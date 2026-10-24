// Command syncpeer is a headless sync peer for manual and integration
// testing: a device with one profile that pairs with whoever asks
// (confirming the code without a user), accepts links and serves sync
// sessions. It is not part of the app.
//
// Commands, one per line on standard input; each answer is a line of JSON
// on standard output:
//
//	status   the device ID, address, paired devices, notes, cards, reviews and due cards
//	notes    the first field of every note
//	answer   answer the first due card Good
//	pair     be visible for pairing for 5 minutes (the default at start with -pair)
//	quit     stop
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/scheduler"
	"cardflip/golang/storage"
	cfsync "cardflip/golang/sync"
	"cardflip/golang/syncnet"
)

const profileID = storage.DefaultProfile

// peer is the headless device.
type peer struct {
	dir  string
	col  *core.Collection
	node *syncnet.Node
	out  *json.Encoder

	mu    sync.Mutex // guards the links file and out
	state struct {
		Key     []byte            `json:"key"`
		Name    string            `json:"name"`
		Paired  []syncnet.Device  `json:"paired,omitempty"`
		Revoked []syncnet.Revoked `json:"revoked,omitempty"`
	}
}

// autoUI accepts every pairing request and confirms every code, printing
// it.
type autoUI struct{ p *peer }

func (u autoUI) PairingRequest(ctx context.Context, name string) (bool, error) {
	log.Printf("pairing request from %s", name)
	return true, nil
}

func (u autoUI) ConfirmPairing(ctx context.Context, name, code string) (bool, error) {
	log.Printf("pairing code with %s: %s", name, code)
	return true, nil
}

func (u autoUI) PairingFinished(name string, paired bool) {
	log.Printf("paired with %s: %v", name, paired)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: syncpeer [flags]\n\n"+
			"A headless Cardflip sync peer for testing. It pairs with any device that asks,\n"+
			"accepts links and serves sync sessions. Commands on standard input: status,\n"+
			"notes, answer, pair, quit.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	dir := flag.String("data", "", "data directory (default: a new temporary one)")
	name := flag.String("name", "syncpeer", "device name")
	port := flag.Int("port", 0, "port to listen on (0: any free port)")
	pair := flag.Bool("pair", true, "be visible for pairing at start")
	flag.Parse()
	log.SetPrefix("syncpeer: ")
	log.SetFlags(0)
	if *dir == "" {
		d, err := os.MkdirTemp("", "syncpeer-")
		if err != nil {
			log.Fatal(err)
		}
		*dir = d
	}
	p, err := open(*dir, *name)
	if err != nil {
		log.Fatal(err)
	}
	l, err := syncnet.Listen(p.node.Identity(), *port, p.node.Verify, p.node.Handle)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	if *pair {
		p.node.OpenPairing(syncnet.PairingWindow)
	}
	log.Printf("device %s listening on port %d, data in %s", p.node.ID(), l.Port(), *dir)
	p.reply(map[string]any{"ready": true, "device": p.node.ID(), "port": l.Port()})
	p.commands(os.Stdin, l)
}

// open opens the peer's collection and device state in dir.
func open(dir, name string) (*peer, error) {
	b, err := storage.OpenBackend(dir)
	if err != nil {
		return nil, err
	}
	col, _, err := core.Open(b, core.Options{})
	if err != nil {
		return nil, err
	}
	p := &peer{dir: dir, col: col, out: json.NewEncoder(os.Stdout)}
	data, err := storage.ReadSync(dir)
	switch {
	case errors.Is(err, storage.ErrNotFound):
		p.state.Key, p.state.Name = syncnet.NewKey(), name
	case err != nil:
		return nil, err
	default:
		if err := json.Unmarshal(data, &p.state); err != nil {
			return nil, err
		}
	}
	id, err := syncnet.IdentityFromSeed(p.state.Key)
	if err != nil {
		return nil, err
	}
	col.SetDevice(id.ID())
	p.node = syncnet.NewNode(id, p.state.Name, autoUI{p}, time.Now, p.state.Paired, p.state.Revoked)
	p.node.Save = p.save
	p.node.OnUnpaired = func(id string) {
		p.mu.Lock()
		defer p.mu.Unlock()
		links, _ := p.links()
		_ = p.saveLinks(slices.DeleteFunc(links, func(l cfsync.Link) bool { return l.Peer == id }))
	}
	p.node.OnSession = p.serve
	return p, p.save()
}

func (p *peer) save() error {
	p.state.Paired, p.state.Revoked = p.node.Devices(), p.node.RevokedKeys()
	data, err := json.Marshal(p.state)
	if err != nil {
		return err
	}
	return storage.WriteSync(p.dir, data)
}

// links returns the profile's links. Callers hold mu.
func (p *peer) links() ([]cfsync.Link, error) {
	data, err := storage.ReadSyncLinks(p.dir, profileID)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []cfsync.Link
	return out, json.Unmarshal(data, &out)
}

// saveLinks saves the profile's links. Callers hold mu.
func (p *peer) saveLinks(links []cfsync.Link) error {
	if links == nil {
		links = []cfsync.Link{}
	}
	data, err := json.Marshal(links)
	if err != nil {
		return err
	}
	return storage.WriteSyncLinks(p.dir, profileID, data)
}

// serve answers a paired device's requests and sessions.
func (p *peer) serve(dev syncnet.Device, first *cfsync.Message, c syncnet.SessionConn) {
	defer c.Close()
	switch first.Type {
	case syncnet.MsgListProfiles:
		data, _ := json.Marshal([]syncnet.ProfileInfo{{ID: profileID, Name: "Default"}})
		_ = c.Send(&cfsync.Message{Type: syncnet.MsgProfiles, Data: data})
	case syncnet.MsgCreateLink:
		var req syncnet.CreateLinkRequest
		if err := json.Unmarshal(first.Data, &req); err != nil {
			return
		}
		p.mu.Lock()
		links, err := p.links()
		if err == nil {
			links = slices.DeleteFunc(links, func(l cfsync.Link) bool { return l.Peer == dev.ID })
			err = p.saveLinks(append(links, cfsync.NewLink(req.LinkID, dev.ID, req.FromProfile)))
		}
		p.mu.Unlock()
		if err != nil {
			_ = c.Send(&cfsync.Message{Type: syncnet.MsgFailed, Error: err.Error()})
			return
		}
		log.Printf("linked with profile %q of %s", req.FromName, dev.Name)
		data, _ := json.Marshal(syncnet.ProfileInfo{ID: profileID, Name: "Default"})
		_ = c.Send(&cfsync.Message{Type: syncnet.MsgLinked, Data: data})
	case syncnet.MsgUnlink:
		p.mu.Lock()
		links, _ := p.links()
		_ = p.saveLinks(slices.DeleteFunc(links, func(l cfsync.Link) bool { return l.ID == first.Name }))
		p.mu.Unlock()
	case "hello":
		lookup := func(peerID, linkID string) (cfsync.Collection, *cfsync.Link, func(), error) {
			p.mu.Lock()
			defer p.mu.Unlock()
			links, err := p.links()
			if err != nil {
				return nil, nil, nil, err
			}
			for _, l := range links {
				if l.ID == linkID && l.Peer == peerID {
					return p.col, &l, nil, nil
				}
			}
			return nil, nil, nil, nil
		}
		opts := cfsync.Options{Device: p.node.ID(), Name: p.node.Name(), SaveLink: func(l cfsync.Link) error {
			p.mu.Lock()
			defer p.mu.Unlock()
			links, err := p.links()
			if err != nil {
				return err
			}
			for i := range links {
				if links[i].ID == l.ID {
					links[i] = l
				}
			}
			return p.saveLinks(links)
		}}
		sum, err := cfsync.Respond(context.Background(), c, dev.ID, lookup, opts)
		if err != nil {
			log.Printf("sync with %s failed: %v", dev.Name, err)
			return
		}
		log.Printf("synced with %s (%s): received %d, sent %d, media %d", dev.Name, sum.Mode, sum.Received, sum.Sent, sum.Media)
	}
}

// commands runs the commands of r.
func (p *peer) commands(r io.Reader, l *syncnet.Listener) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		switch cmd := strings.TrimSpace(sc.Text()); cmd {
		case "":
		case "status":
			p.reply(p.status(l))
		case "notes":
			var fronts []string
			for _, id := range p.col.AllNoteIDs() {
				if n, err := p.col.Note(id); err == nil && len(n.Fields) > 0 {
					fronts = append(fronts, n.Fields[0])
				}
			}
			p.reply(map[string]any{"notes": fronts})
		case "answer":
			id, err := p.answer()
			if err != nil {
				p.reply(map[string]any{"error": err.Error()})
				continue
			}
			p.reply(map[string]any{"answered": id})
		case "pair":
			p.node.OpenPairing(syncnet.PairingWindow)
			p.reply(map[string]any{"pairing": true})
		case "quit":
			return
		default:
			p.reply(map[string]any{"error": "unknown command " + strconv.Quote(cmd)})
		}
	}
}

func (p *peer) reply(v any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	_ = p.out.Encode(v)
}

// status describes the peer.
func (p *peer) status(l *syncnet.Listener) map[string]any {
	due := 0
	counts := p.col.DeckCounts()
	for _, d := range p.col.Decks() {
		if !strings.Contains(d.Name, core.DeckSeparator) {
			due += counts[d.ID].Total() // with its subdecks
		}
	}
	reviews := 0
	months, _ := p.col.RevlogMonths()
	for _, m := range months {
		e, _ := p.col.ReadRevlog(m)
		reviews += len(e)
	}
	var paired []string
	for _, d := range p.node.Devices() {
		paired = append(paired, d.Name)
	}
	_, addr := syncnet.LocalNetwork()
	return map[string]any{"device": p.node.ID(), "name": p.node.Name(), "port": l.Port(),
		"address": net.JoinHostPort(addr, strconv.Itoa(l.Port())), "paired": paired,
		"notes": len(p.col.AllNoteIDs()), "cards": p.col.CardCount(), "reviews": reviews, "due": due}
}

// answer answers the first due card of the collection Good.
func (p *peer) answer() (core.ID, error) {
	ids, err := p.col.Search("is:due or is:new", core.ColCreated, false)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, errors.New("no card is due")
	}
	if _, err := p.col.Answer(ids[0], scheduler.Good, 3*time.Second); err != nil {
		return 0, err
	}
	return ids[0], nil
}
