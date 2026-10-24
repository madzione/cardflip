//go:build js

package storage

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"syscall/js"
	"time"
)

// IndexedDB is the web Backend. It uses seven object stores:
//
//	values       key → bytes             (Put/Get)
//	records      [key, seq] → bytes      (Append/Records)
//	logs         key → next seq          (which logs exist)
//	backups      name → bytes            (BackupStore)
//	backupSizes  name → size in bytes    (lists backups without reading them)
//	trash        name → bytes            (MediaTrash)
//	trashInfo    name → [size, ms]       (lists the trash without reading it)
//
// Every mutation is a single transaction with strict durability, and returns
// after the transaction completes. A transaction is all-or-nothing, so logs
// never have torn records. Promises and requests are awaited by blocking on
// a channel, which is safe because fluffy runs every call in a goroutine.
type IndexedDB struct {
	db js.Value

	mu   sync.Mutex
	next map[string]int // next record sequence number per log
}

// idbVersion 2 added the backup stores, 3 the media trash.
const idbVersion = 3

// OpenBackend opens the platform backend. On the web the data directory
// names the database: "" is the app's own collection.
func OpenBackend(dataDir string) (Backend, error) {
	return OpenIndexedDB(dbName(dataDir))
}

// RemoveBackend deletes everything OpenBackend(dataDir) stored.
func RemoveBackend(dataDir string) error { return DeleteIndexedDB(dbName(dataDir)) }

func dbName(dataDir string) string {
	if dataDir == "" {
		return "cardflip"
	}
	return "cardflip-" + dataDir
}

// OpenIndexedDB opens (creating if needed) the database name.
func OpenIndexedDB(name string) (*IndexedDB, error) {
	idb := js.Global().Get("indexedDB")
	if idb.IsUndefined() || idb.IsNull() {
		return nil, errors.New("storage: this browser has no IndexedDB (private mode?)")
	}
	req := idb.Call("open", name, idbVersion)
	upgrade := js.FuncOf(func(this js.Value, args []js.Value) any {
		db := req.Get("result")
		stores := db.Get("objectStoreNames")
		for _, s := range []string{"values", "records", "logs", "backups", "backupSizes", "trash", "trashInfo"} {
			if !stores.Call("contains", s).Bool() {
				db.Call("createObjectStore", s)
			}
		}
		return nil
	})
	defer upgrade.Release()
	req.Set("onupgradeneeded", upgrade)
	db, err := awaitRequest(req)
	if err != nil {
		return nil, fmt.Errorf("storage: open IndexedDB: %w", err)
	}
	b := &IndexedDB{db: db, next: map[string]int{}}

	// Load the log counters; there is one entry per log, not per record.
	tx := b.tx("readonly", "logs")
	store := tx.Call("objectStore", "logs")
	keys, err := awaitRequest(store.Call("getAllKeys"))
	if err != nil {
		return nil, err
	}
	vals, err := awaitRequest(store.Call("getAll"))
	if err != nil {
		return nil, err
	}
	for i := 0; i < keys.Length(); i++ {
		b.next[keys.Index(i).String()] = vals.Index(i).Int()
	}
	return b, nil
}

// DeleteIndexedDB removes the database name; tests use it to start fresh.
func DeleteIndexedDB(name string) error {
	_, err := awaitRequest(js.Global().Get("indexedDB").Call("deleteDatabase", name))
	return err
}

func (b *IndexedDB) tx(mode string, stores ...string) js.Value {
	names := make([]any, len(stores))
	for i, s := range stores {
		names[i] = s
	}
	opts := js.Global().Get("Object").New()
	opts.Set("durability", "strict")
	return b.db.Call("transaction", js.ValueOf(names), mode, opts)
}

func (b *IndexedDB) Get(key string) ([]byte, error) {
	if err := ValidKey(key); err != nil {
		return nil, err
	}
	v, err := awaitRequest(b.tx("readonly", "values").Call("objectStore", "values").Call("get", key))
	if err != nil {
		return nil, err
	}
	if v.IsUndefined() {
		return nil, ErrNotFound
	}
	return toGo(v), nil
}

func (b *IndexedDB) Put(key string, value []byte) error {
	if err := ValidKey(key); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	tx := b.tx("readwrite", "values", "records", "logs")
	tx.Call("objectStore", "values").Call("put", toJS(value), key)
	if _, ok := b.next[key]; ok {
		tx.Call("objectStore", "records").Call("delete", logRange(key))
		tx.Call("objectStore", "logs").Call("delete", key)
	}
	if err := awaitTx(tx); err != nil {
		return err
	}
	delete(b.next, key)
	return nil
}

func (b *IndexedDB) Append(key string, record []byte) error {
	return b.AppendAll(key, [][]byte{record})
}

// AppendAll appends records in one transaction.
func (b *IndexedDB) AppendAll(key string, records [][]byte) error {
	if err := ValidKey(key); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	seq := b.next[key]
	tx := b.tx("readwrite", "records", "logs")
	store := tx.Call("objectStore", "records")
	for i, r := range records {
		store.Call("put", toJS(r), js.ValueOf([]any{key, seq + i}))
	}
	tx.Call("objectStore", "logs").Call("put", seq+len(records), key)
	if err := awaitTx(tx); err != nil {
		return err
	}
	b.next[key] = seq + len(records)
	return nil
}

func (b *IndexedDB) Records(key string) ([][]byte, error) {
	if err := ValidKey(key); err != nil {
		return nil, err
	}
	vals, err := awaitRequest(b.tx("readonly", "records").Call("objectStore", "records").Call("getAll", logRange(key)))
	if err != nil {
		return nil, err
	}
	out := make([][]byte, vals.Length())
	for i := range out {
		out[i] = toGo(vals.Index(i))
	}
	return out, nil
}

func (b *IndexedDB) List(prefix string) ([]string, error) {
	store := b.tx("readonly", "values").Call("objectStore", "values")
	var keys js.Value
	var err error
	if prefix == "" {
		keys, err = awaitRequest(store.Call("getAllKeys"))
	} else {
		keys, err = awaitRequest(store.Call("getAllKeys",
			js.Global().Get("IDBKeyRange").Call("bound", prefix, prefix+"￿")))
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for i := 0; i < keys.Length(); i++ {
		if k := keys.Index(i).String(); strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	b.mu.Lock()
	for k := range b.next {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	b.mu.Unlock()
	sort.Strings(out)
	return out, nil
}

func (b *IndexedDB) Delete(key string) error {
	if err := ValidKey(key); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	tx := b.tx("readwrite", "values", "records", "logs")
	tx.Call("objectStore", "values").Call("delete", key)
	tx.Call("objectStore", "records").Call("delete", logRange(key))
	tx.Call("objectStore", "logs").Call("delete", key)
	if err := awaitTx(tx); err != nil {
		return err
	}
	delete(b.next, key)
	return nil
}

func (b *IndexedDB) Close() error {
	b.db.Call("close")
	return nil
}

func (b *IndexedDB) Backups() BackupStore { return idbBackups{b} }

type idbBackups struct{ b *IndexedDB }

func (ib idbBackups) List() ([]BackupFile, error) {
	store := ib.b.tx("readonly", "backupSizes").Call("objectStore", "backupSizes")
	keys, err := awaitRequest(store.Call("getAllKeys"))
	if err != nil {
		return nil, err
	}
	sizes, err := awaitRequest(store.Call("getAll"))
	if err != nil {
		return nil, err
	}
	out := make([]BackupFile, keys.Length())
	for i := range out {
		out[i] = BackupFile{Name: keys.Index(i).String(), Size: int64(sizes.Index(i).Float())}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (ib idbBackups) Write(name string, data []byte) error {
	if err := ValidBackupName(name); err != nil {
		return err
	}
	tx := ib.b.tx("readwrite", "backups", "backupSizes")
	tx.Call("objectStore", "backups").Call("put", toJS(data), name)
	tx.Call("objectStore", "backupSizes").Call("put", len(data), name)
	return awaitTx(tx)
}

func (ib idbBackups) Read(name string) ([]byte, error) {
	if err := ValidBackupName(name); err != nil {
		return nil, err
	}
	v, err := awaitRequest(ib.b.tx("readonly", "backups").Call("objectStore", "backups").Call("get", name))
	if err != nil {
		return nil, err
	}
	if v.IsUndefined() {
		return nil, ErrNotFound
	}
	return toGo(v), nil
}

func (ib idbBackups) Delete(name string) error {
	if err := ValidBackupName(name); err != nil {
		return err
	}
	tx := ib.b.tx("readwrite", "backups", "backupSizes")
	tx.Call("objectStore", "backups").Call("delete", name)
	tx.Call("objectStore", "backupSizes").Call("delete", name)
	return awaitTx(tx)
}

func (b *IndexedDB) Trash() MediaTrash { return idbTrash{b} }

type idbTrash struct{ b *IndexedDB }

func (it idbTrash) List() ([]TrashedMedia, error) {
	store := it.b.tx("readonly", "trashInfo").Call("objectStore", "trashInfo")
	keys, err := awaitRequest(store.Call("getAllKeys"))
	if err != nil {
		return nil, err
	}
	infos, err := awaitRequest(store.Call("getAll"))
	if err != nil {
		return nil, err
	}
	out := make([]TrashedMedia, keys.Length())
	for i := range out {
		info := infos.Index(i)
		out[i] = TrashedMedia{Name: keys.Index(i).String(), Size: int64(info.Index(0).Float()),
			Trashed: time.UnixMilli(int64(info.Index(1).Float()))}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (it idbTrash) Put(name string, data []byte, trashed time.Time) error {
	if err := validTrashName(name); err != nil {
		return err
	}
	tx := it.b.tx("readwrite", "trash", "trashInfo")
	tx.Call("objectStore", "trash").Call("put", toJS(data), name)
	tx.Call("objectStore", "trashInfo").Call("put", js.ValueOf([]any{len(data), float64(trashed.UnixMilli())}), name)
	return awaitTx(tx)
}

func (it idbTrash) Get(name string) ([]byte, error) {
	if err := validTrashName(name); err != nil {
		return nil, err
	}
	v, err := awaitRequest(it.b.tx("readonly", "trash").Call("objectStore", "trash").Call("get", name))
	if err != nil {
		return nil, err
	}
	if v.IsUndefined() {
		return nil, ErrNotFound
	}
	return toGo(v), nil
}

func (it idbTrash) Delete(name string) error {
	if err := validTrashName(name); err != nil {
		return err
	}
	tx := it.b.tx("readwrite", "trash", "trashInfo")
	tx.Call("objectStore", "trash").Call("delete", name)
	tx.Call("objectStore", "trashInfo").Call("delete", name)
	return awaitTx(tx)
}

// RequestPersistence asks the browser not to evict the site's storage and
// reports whether it agreed.
func RequestPersistence() bool {
	storage := js.Global().Get("navigator").Get("storage")
	if storage.IsUndefined() || storage.Get("persist").IsUndefined() {
		return false
	}
	v, err := awaitPromise(storage.Call("persist"))
	return err == nil && v.Truthy()
}

func logRange(key string) js.Value {
	return js.Global().Get("IDBKeyRange").Call("bound",
		js.ValueOf([]any{key, 0}), js.ValueOf([]any{key, js.Global().Get("Infinity")}))
}

func toJS(b []byte) js.Value {
	arr := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(arr, b)
	return arr
}

func toGo(v js.Value) []byte {
	if v.InstanceOf(js.Global().Get("ArrayBuffer")) {
		v = js.Global().Get("Uint8Array").New(v)
	}
	out := make([]byte, v.Get("length").Int())
	js.CopyBytesToGo(out, v)
	return out
}

func domError(v js.Value) error {
	if v.IsUndefined() || v.IsNull() {
		return errors.New("storage: IndexedDB error")
	}
	return fmt.Errorf("storage: IndexedDB: %s: %s", v.Get("name").String(), v.Get("message").String())
}

// awaitRequest waits for an IDBRequest and returns its result.
func awaitRequest(req js.Value) (js.Value, error) {
	type result struct {
		v   js.Value
		err error
	}
	ch := make(chan result, 1)
	ok := js.FuncOf(func(js.Value, []js.Value) any {
		ch <- result{v: req.Get("result")}
		return nil
	})
	fail := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			args[0].Call("preventDefault")
		}
		ch <- result{err: domError(req.Get("error"))}
		return nil
	})
	defer ok.Release()
	defer fail.Release()
	req.Set("onsuccess", ok)
	req.Set("onerror", fail)
	r := <-ch
	return r.v, r.err
}

// awaitTx waits until a transaction has committed.
func awaitTx(tx js.Value) error {
	ch := make(chan error, 1)
	done := js.FuncOf(func(js.Value, []js.Value) any { ch <- nil; return nil })
	fail := js.FuncOf(func(js.Value, []js.Value) any { ch <- domError(tx.Get("error")); return nil })
	defer done.Release()
	defer fail.Release()
	tx.Set("oncomplete", done)
	tx.Set("onerror", fail)
	tx.Set("onabort", fail)
	return <-ch
}

func awaitPromise(p js.Value) (js.Value, error) {
	type result struct {
		v   js.Value
		err error
	}
	ch := make(chan result, 1)
	ok := js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- result{v: args[0]}
		return nil
	})
	fail := js.FuncOf(func(_ js.Value, args []js.Value) any {
		ch <- result{err: fmt.Errorf("storage: %s", args[0].Call("toString").String())}
		return nil
	})
	defer ok.Release()
	defer fail.Release()
	p.Call("then", ok, fail)
	r := <-ch
	return r.v, r.err
}
