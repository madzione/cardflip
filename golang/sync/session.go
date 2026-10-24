package sync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/storage"
)

// Collection is what a session needs of a collection; *core.Collection
// implements it.
type Collection interface {
	Epoch() string
	Seq() int64
	Slots() map[string]int
	SetSlots(map[string]int) error
	Delta(core.DeltaRequest) (core.DeltaPage, error)
	ApplyBatch([]core.DeltaItem, core.BatchSource) (core.ApplyResult, error)
	Identities() (core.Identities, error)
	Renumber(core.Identities) (int, error)
	BuildMediaManifest() error
	MissingMedia() []core.MediaFileRef
	MediaBlob(name, sha string) ([]byte, error)
	PutSyncedMedia(name string, data []byte) error
	Snapshot() ([]byte, error)
	RevlogMonths() ([]string, error)
	ReadRevlog(month string) ([]json.RawMessage, error)
	BeginReplace() (*core.Replacing, error)
}

var _ Collection = (*core.Collection)(nil)

// Conn carries the messages of a session. The network layer frames,
// compresses and encrypts them; a closed or broken connection makes Send
// and Recv fail.
type Conn interface {
	Send(*Message) error
	Recv() (*Message, error)
}

// Options configure a session.
type Options struct {
	// Device and Name are this device's ID and name.
	Device string
	Name   string
	// Now is the clock; nil means time.Now.
	Now func() time.Time
	// Backup makes an automatic backup of the collection before a
	// first-sync choice changes it.
	Backup func() error
	// SaveLink saves the link's state. It is called after every batch, so
	// an interrupted session continues where it stopped.
	SaveLink func(Link) error
	// Progress, if set, is told how far the session got.
	Progress func(Progress)
	// Protocol and MinProtocol override the protocol versions (tests);
	// MaxItems the batch size.
	Protocol, MinProtocol int
	MaxItems              int
}

// Progress is how far a session got.
type Progress struct {
	Phase    string `json:"phase"`
	Received int    `json:"received"`
	Sent     int    `json:"sent"`
}

// Session phases.
const (
	PhaseReceiving = "receiving"
	PhaseSending   = "sending"
	PhaseMedia     = "media"
	PhaseReplacing = "replacing"
)

// Summary is what a session did: how many changes it received and sent,
// and how many media files it fetched. Mode is ModeSync, ModeMerge,
// ModeReplaced (this collection was replaced) or ModeReplacedPeer.
type Summary struct {
	Received int    `json:"received"`
	Sent     int    `json:"sent"`
	Media    int    `json:"media"`
	Mode     string `json:"mode"`
}

// Session modes.
const (
	ModeSync         = "sync"
	ModeMerge        = "merge"
	ModeReplaced     = "replaced"
	ModeReplacedPeer = "replacedPeer"
)

// Transfer sizes: data chunks of media files and snapshots, review log
// entries and identities per message.
const (
	chunkBytes    = 4 << 20
	revlogChunk   = 2000
	identityChunk = 100000
)

type session struct {
	ctx       context.Context
	conn      Conn
	col       Collection
	link      *Link
	opts      Options
	peer      *Hello
	initiator bool
	sum       Summary
}

func (s *session) now() time.Time {
	if s.opts.Now != nil {
		return s.opts.Now()
	}
	return time.Now()
}

func (s *session) protocol() (int, int) {
	p, m := ProtocolVersion, MinProtocolVersion
	if s.opts.Protocol != 0 {
		p = s.opts.Protocol
	}
	if s.opts.MinProtocol != 0 {
		m = s.opts.MinProtocol
	}
	return p, m
}

func (s *session) send(m *Message) error {
	if err := s.ctx.Err(); err != nil {
		return err
	}
	return s.conn.Send(m)
}

// recv receives the next message, which must be of one of types.
func (s *session) recv(types ...string) (*Message, error) {
	if err := s.ctx.Err(); err != nil {
		return nil, err
	}
	m, err := s.conn.Recv()
	if err != nil {
		return nil, err
	}
	switch m.Type {
	case msgRefuse:
		if m.Refusal != nil {
			return nil, m.Refusal
		}
	case msgFail:
		return nil, fmt.Errorf("the other device stopped: %s", m.Error)
	}
	for _, t := range types {
		if m.Type == t {
			return m, nil
		}
	}
	return nil, fmt.Errorf("unexpected sync message %q", m.Type)
}

// fail tells the peer the session failed, and returns err.
func (s *session) fail(err error) error {
	var r *Refusal
	if errors.As(err, &r) {
		return err
	}
	_ = s.conn.Send(&Message{Type: msgFail, Error: err.Error()})
	return err
}

func (s *session) save() error {
	if s.opts.SaveLink == nil {
		return nil
	}
	return s.opts.SaveLink(*s.link)
}

func (s *session) progress(phase string) {
	if s.opts.Progress != nil {
		s.opts.Progress(Progress{Phase: phase, Received: s.sum.Received, Sent: s.sum.Sent})
	}
}

// hello returns this device's hello.
func (s *session) hello(first bool, choice Choice) *Hello {
	p, m := s.protocol()
	return &Hello{Protocol: p, MinProtocol: m, Snapshot: core.SnapshotVersion, Device: s.opts.Device, Name: s.opts.Name,
		Time: s.now().UnixMilli(), Link: s.link.ID, Epoch: s.col.Epoch(), SeenEpoch: s.link.PeerEpoch,
		FirstSync: first, Choice: choice, Slots: s.col.Slots()}
}

// check checks the peer's hello (spec: Compatible versions, Clocks must
// agree). It returns the refusal for this device and the one for the peer.
func (s *session) check(h *Hello, rtt time.Duration) (mine, theirs *Refusal) {
	p, m := s.protocol()
	if agreed := min(p, h.Protocol); agreed < max(m, h.MinProtocol) {
		if p < h.Protocol {
			return &Refusal{Code: RefuseVersion, Device: s.opts.Name, Mine: true},
				&Refusal{Code: RefuseVersion, Device: s.opts.Name}
		}
		return &Refusal{Code: RefuseVersion, Device: h.Name},
			&Refusal{Code: RefuseVersion, Device: h.Name, Mine: true}
	}
	skew := time.UnixMilli(h.Time).Add(rtt / 2).Sub(s.now())
	if skew.Abs() > MaxClockSkew {
		return &Refusal{Code: RefuseClock, Device: h.Name, Skew: skew.Milliseconds()},
			&Refusal{Code: RefuseClock, Device: s.opts.Name, Skew: -skew.Milliseconds()}
	}
	return nil, nil
}

// mergeSlots merges the slot tables and saves the result (decision 4).
func (s *session) mergeSlots() error {
	merged := MergeSlots(s.col.Slots(), s.peer.Slots, s.opts.Device, s.peer.Device)
	if maps.Equal(merged, s.col.Slots()) {
		return nil
	}
	return s.col.SetSlots(merged)
}

// Initiate runs a session with the peer of link as the device that starts
// it. choice is the user's answer to the first-sync question, if they gave
// one; without one, a link that needs it ends the session with a Refusal
// of code RefuseNeedsChoice and is marked NeedsChoice. The link is saved
// through opts.SaveLink as the session goes.
func Initiate(ctx context.Context, conn Conn, col Collection, link *Link, choice Choice, opts Options) (Summary, error) {
	s := &session{ctx: ctx, conn: conn, col: col, link: link, opts: opts, initiator: true}
	if !choice.valid() {
		return Summary{}, fmt.Errorf("invalid choice %q", choice)
	}
	if choice == ChoiceNone {
		choice = link.Pending
	}
	if err := col.BuildMediaManifest(); err != nil {
		return Summary{}, err
	}
	first := link.firstSync(col.Epoch(), link.PeerEpoch, s.now())
	start := s.now()
	if err := s.send(&Message{Type: msgHello, Hello: s.hello(first, choice)}); err != nil {
		return Summary{}, err
	}
	m, err := s.recv(msgHello)
	if err != nil {
		var r *Refusal
		if errors.As(err, &r) && r.Code == RefuseNeedsChoice {
			link.NeedsChoice = true
			_ = s.save()
		}
		return Summary{}, err
	}
	s.peer = m.Hello
	if mine, theirs := s.check(s.peer, s.now().Sub(start)); mine != nil {
		_ = s.send(&Message{Type: msgRefuse, Refusal: theirs})
		return Summary{}, mine
	}
	if err := s.mergeSlots(); err != nil {
		return Summary{}, s.fail(err)
	}
	return s.run(first || s.peer.FirstSync, s.peer.Choice)
}

// Lookup finds the collection and link a peer's hello names. release is
// called when the session ends. A nil link means the link is unknown.
type Lookup func(peer, link string) (col Collection, l *Link, release func(), err error)

// Respond runs a session that the paired device peer started.
func Respond(ctx context.Context, conn Conn, peer string, lookup Lookup, opts Options) (Summary, error) {
	s := &session{ctx: ctx, conn: conn, opts: opts}
	m, err := s.recv(msgHello)
	if err != nil {
		return Summary{}, err
	}
	s.peer = m.Hello
	if s.peer.Device != peer {
		return Summary{}, s.fail(fmt.Errorf("hello from %s on a connection of %s", s.peer.Device, peer))
	}
	if mine, theirs := s.check(s.peer, 0); mine != nil {
		_ = s.send(&Message{Type: msgRefuse, Refusal: theirs})
		return Summary{}, mine
	}
	col, link, release, err := lookup(peer, s.peer.Link)
	if err != nil {
		return Summary{}, s.fail(err)
	}
	if release != nil {
		defer release()
	}
	if link == nil || link.Peer != peer {
		_ = s.send(&Message{Type: msgRefuse, Refusal: &Refusal{Code: RefuseUnknownLink, Device: opts.Name}})
		return Summary{}, &Refusal{Code: RefuseUnknownLink, Device: s.peer.Name}
	}
	s.col, s.link = col, link
	if err := col.BuildMediaManifest(); err != nil {
		return Summary{}, s.fail(err)
	}
	first := link.firstSync(col.Epoch(), s.peer.Epoch, s.now()) || s.peer.FirstSync
	choice := s.peer.Choice
	if choice == ChoiceNone {
		choice = link.Pending.mirrored()
	}
	if !first {
		choice = ChoiceNone
	}
	if first && choice == ChoiceNone {
		_ = s.send(&Message{Type: msgRefuse, Refusal: &Refusal{Code: RefuseNeedsChoice, Device: opts.Name}})
		return Summary{}, &Refusal{Code: RefuseNeedsChoice, Device: s.peer.Name}
	}
	if err := s.send(&Message{Type: msgHello, Hello: s.hello(first, choice)}); err != nil {
		return Summary{}, err
	}
	if err := s.mergeSlots(); err != nil {
		return Summary{}, s.fail(err)
	}
	return s.run(first, choice)
}

// run carries out the session after the hellos. choice is from the
// initiator's point of view.
func (s *session) run(first bool, choice Choice) (Summary, error) {
	if !first {
		s.sum.Mode = ModeSync
		return s.finish(s.exchange())
	}
	mine := choice
	if !s.initiator {
		mine = choice.mirrored()
	}
	if s.link.Pending != mine {
		// A new first sync: start from scratch.
		s.link.Pending, s.link.NeedsChoice = mine, false
		s.link.Received, s.link.Cursor, s.link.CursorSeq, s.link.Echo = -1, nil, 0, nil
		if err := s.save(); err != nil {
			return Summary{}, s.fail(err)
		}
	}
	switch mine {
	case ChoiceMerge:
		s.sum.Mode = ModeMerge
		return s.finish(s.merge())
	case ChoiceReplaceOther:
		s.sum.Mode = ModeReplacedPeer
		return s.finish(s.sendReplacement())
	case ChoiceReplaceThis:
		s.sum.Mode = ModeReplaced
		return s.finish(s.receiveReplacement())
	}
	return Summary{}, s.fail(fmt.Errorf("invalid choice %q", choice))
}

// finish exchanges the closing messages and records the finished sync.
func (s *session) finish(err error) (Summary, error) {
	if err != nil {
		return s.sum, s.fail(err)
	}
	bye := &Message{Type: msgBye, Seq: s.link.Received, Epoch: s.col.Epoch()}
	if s.sum.Mode == ModeReplaced {
		bye.Have = s.col.Seq()
	}
	var peer *Message
	if s.initiator {
		if err = s.send(bye); err == nil {
			peer, err = s.recv(msgBye)
		}
	} else if peer, err = s.recv(msgBye); err == nil {
		err = s.send(bye)
	}
	if err != nil {
		return s.sum, err
	}
	l := s.link
	if s.sum.Mode == ModeReplacedPeer {
		// Everything the peer has now came from here.
		l.Received, l.Cursor = peer.Have, nil
	}
	l.Epoch, l.PeerEpoch = s.col.Epoch(), peer.Epoch
	l.PeerReceived = peer.Seq
	l.Echo = l.Echo.Prune(l.PeerReceived)
	l.LastSeq, l.LastSynced = s.col.Seq(), s.now()
	l.Pending, l.NeedsChoice = ChoiceNone, false
	return s.sum, s.save()
}

// exchange sends and receives the changes since the last sync: the
// initiator pulls first, then the responder.
func (s *session) exchange() error {
	if s.initiator {
		if err := s.pull(); err != nil {
			return err
		}
		if err := s.send(&Message{Type: msgTurn}); err != nil {
			return err
		}
		return s.serve(msgDone)
	}
	if err := s.serve(msgTurn); err != nil {
		return err
	}
	if err := s.pull(); err != nil {
		return err
	}
	return s.send(&Message{Type: msgDone})
}

// pull receives the peer's delta since the link's watermark, continuing an
// interrupted one, and then the media files that are missing.
func (s *session) pull() error {
	l := s.link
	watermark := int64(-1)
	if l.Cursor != nil {
		watermark = l.CursorSeq
	}
	req := &Message{Type: msgDeltaReq, Since: l.Received, After: l.Cursor}
	s.progress(PhaseReceiving)
	for {
		if err := s.send(req); err != nil {
			return err
		}
		m, err := s.recv(msgDelta)
		if err != nil {
			return err
		}
		if m.Page == nil {
			return errors.New("delta message without a page")
		}
		if watermark < 0 {
			watermark = m.Page.Seq
		}
		if len(m.Page.Items) > 0 {
			res, err := s.col.ApplyBatch(m.Page.Items, core.BatchSource{Link: l.ID, Received: l.Echo, LastSync: l.LastSeq})
			if err != nil {
				return err
			}
			if res.Last > 0 {
				l.Echo = l.Echo.Add(res.First, res.Last)
			}
			s.sum.Received += res.Changed
			cur := m.Page.Items[len(m.Page.Items)-1].Cursor()
			l.Cursor, l.CursorSeq = &cur, watermark
			req.After = &cur
		}
		if m.Page.Done {
			l.Received, l.Cursor, l.CursorSeq = watermark, nil, 0
			if err := s.save(); err != nil {
				return err
			}
			break
		}
		if err := s.save(); err != nil {
			return err
		}
		s.progress(PhaseReceiving)
	}
	return s.pullMedia()
}

// pullMedia fetches the media files the collection lacks from the peer.
func (s *session) pullMedia() error {
	missing := s.col.MissingMedia()
	if len(missing) > 0 {
		s.progress(PhaseMedia)
	}
	for _, f := range missing {
		name := f.Name
		if err := s.send(&Message{Type: msgMediaReq, Name: name, SHA1: f.SHA1}); err != nil {
			return err
		}
		var data []byte
		for {
			m, err := s.recv(msgMedia)
			if err != nil {
				return err
			}
			if m.Name != name {
				return fmt.Errorf("asked for media file %s, got %s", name, m.Name)
			}
			if m.Error != "" {
				data = nil
				break
			}
			data = append(data, m.Data...)
			if !m.More {
				break
			}
		}
		if data == nil {
			continue // the peer does not have it either
		}
		if err := s.col.PutSyncedMedia(name, data); err != nil {
			return err
		}
		s.sum.Media++
	}
	return nil
}

// serve answers the peer's requests for its delta and media files until it
// sends end.
func (s *session) serve(end string) error {
	maxItems := s.opts.MaxItems
	for {
		m, err := s.recv(msgDeltaReq, msgMediaReq, end)
		if err != nil {
			return err
		}
		switch m.Type {
		case end:
			return nil
		case msgDeltaReq:
			req := core.DeltaRequest{Since: m.Since, Skip: s.link.Echo, MaxItems: maxItems}
			if m.After != nil {
				req.After = *m.After
			}
			page, err := s.col.Delta(req)
			if err != nil {
				return err
			}
			s.sum.Sent += len(page.Items)
			if err := s.send(&Message{Type: msgDelta, Page: &page}); err != nil {
				return err
			}
			s.progress(PhaseSending)
		case msgMediaReq:
			if err := s.sendMedia(m.Name, m.SHA1); err != nil {
				return err
			}
		}
	}
}

// sendMedia sends a media file with the content sha in chunks.
func (s *session) sendMedia(name, sha string) error {
	data, err := s.col.MediaBlob(name, sha)
	if errors.Is(err, storage.ErrNotFound) {
		return s.send(&Message{Type: msgMedia, Name: name, Error: "missing"})
	}
	if err != nil {
		return err
	}
	for {
		n := min(len(data), chunkBytes)
		if err := s.send(&Message{Type: msgMedia, Name: name, Data: data[:n], More: n < len(data)}); err != nil {
			return err
		}
		data = data[n:]
		if len(data) == 0 {
			return nil
		}
	}
}

// merge carries out the first-sync choice Merge (decision 7): automatic
// backups, the responder renumbers what collides with the initiator's
// objects (decision 4), and both exchange everything. Both keep their
// lineage: their other links stay valid.
func (s *session) merge() error {
	if s.opts.Backup != nil {
		if err := s.opts.Backup(); err != nil {
			return fmt.Errorf("could not make the backup before merging: %w", err)
		}
	}
	if s.initiator {
		ids, err := s.col.Identities()
		if err != nil {
			return err
		}
		if err := s.sendIdentities(ids); err != nil {
			return err
		}
		if _, err := s.recv(msgRenumbered); err != nil {
			return err
		}
	} else {
		ids := core.Identities{Notes: map[core.ID]uint64{}, Cards: map[core.ID]uint64{}, Reviews: map[core.ID]uint64{}}
		for {
			m, err := s.recv(msgIdentities)
			if err != nil {
				return err
			}
			if m.Identities != nil {
				maps.Copy(ids.Notes, m.Identities.Notes)
				maps.Copy(ids.Cards, m.Identities.Cards)
				maps.Copy(ids.Reviews, m.Identities.Reviews)
			}
			if !m.More {
				break
			}
		}
		n, err := s.col.Renumber(ids)
		if err != nil {
			return err
		}
		if err := s.send(&Message{Type: msgRenumbered, Count: n}); err != nil {
			return err
		}
	}
	return s.exchange()
}

// sendIdentities sends identities in parts of identityChunk entries.
func (s *session) sendIdentities(ids core.Identities) error {
	newPart := func() *core.Identities {
		return &core.Identities{Notes: map[core.ID]uint64{}, Cards: map[core.ID]uint64{}, Reviews: map[core.ID]uint64{}}
	}
	part, n := newPart(), 0
	flush := func(more bool) error {
		err := s.send(&Message{Type: msgIdentities, Identities: part, More: more})
		part, n = newPart(), 0
		return err
	}
	for _, kind := range []struct {
		src map[core.ID]uint64
		dst func(*core.Identities) map[core.ID]uint64
	}{
		{ids.Notes, func(p *core.Identities) map[core.ID]uint64 { return p.Notes }},
		{ids.Cards, func(p *core.Identities) map[core.ID]uint64 { return p.Cards }},
		{ids.Reviews, func(p *core.Identities) map[core.ID]uint64 { return p.Reviews }},
	} {
		for id, h := range kind.src {
			kind.dst(part)[id] = h
			if n++; n >= identityChunk {
				if err := flush(true); err != nil {
					return err
				}
			}
		}
	}
	return flush(false)
}

// snapshotInfo is what a replacement needs of the sender's snapshot.
type snapshotInfo struct {
	Seq   int64  `json:"seq"`
	Epoch string `json:"epoch"`
}

// sendReplacement sends this collection to replace the peer's: the
// snapshot first taken, then the review log, then the snapshot itself,
// then the media files the peer asks for.
func (s *session) sendReplacement() error {
	snap, err := s.col.Snapshot()
	if err != nil {
		return err
	}
	var info snapshotInfo
	if err := json.Unmarshal(snap, &info); err != nil {
		return err
	}
	s.progress(PhaseSending)
	months, err := s.col.RevlogMonths()
	if err != nil {
		return err
	}
	for _, month := range months {
		entries, err := s.col.ReadRevlog(month)
		if err != nil {
			return err
		}
		for len(entries) > 0 {
			n := min(len(entries), revlogChunk)
			if err := s.send(&Message{Type: msgRevlog, Entries: entries[:n]}); err != nil {
				return err
			}
			s.sum.Sent += n
			entries = entries[n:]
		}
	}
	for data := snap; ; {
		n := min(len(data), chunkBytes)
		if err := s.send(&Message{Type: msgSnapshot, Data: data[:n], More: n < len(data), Epoch: info.Epoch, Seq: info.Seq}); err != nil {
			return err
		}
		data = data[n:]
		if len(data) == 0 {
			break
		}
	}
	end := msgDone
	if !s.initiator {
		end = msgTurn
	}
	return s.serve(end)
}

// receiveReplacement replaces this collection with the peer's, after an
// automatic backup, and then fetches the media files it lacks.
func (s *session) receiveReplacement() error {
	if s.opts.Backup != nil {
		if err := s.opts.Backup(); err != nil {
			return fmt.Errorf("could not make the backup before replacing: %w", err)
		}
	}
	rp, err := s.col.BeginReplace()
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			rp.Abort()
		}
	}()
	s.progress(PhaseReplacing)
	var snap []byte
	for {
		m, err := s.recv(msgRevlog, msgSnapshot)
		if err != nil {
			return err
		}
		if m.Type == msgRevlog {
			if err := rp.AddReviews(m.Entries); err != nil {
				return err
			}
			continue
		}
		snap = append(snap, m.Data...)
		if m.More {
			continue
		}
		committed = true
		if err := rp.Commit(snap, m.Epoch); err != nil {
			return err
		}
		l := s.link
		l.Received, l.Cursor, l.CursorSeq, l.Echo = m.Seq, nil, 0, nil
		l.Epoch, l.PeerEpoch = m.Epoch, m.Epoch
		if err := s.save(); err != nil {
			return err
		}
		break
	}
	s.sum.Received++
	if err := s.pullMedia(); err != nil {
		return err
	}
	end := msgDone
	if s.initiator {
		end = msgTurn
	}
	return s.send(&Message{Type: end})
}
