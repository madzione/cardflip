// Package sync carries out a sync session between two devices (sync design
// decisions 3, 4 and 7): the hello checks, the first-sync choice, the
// exchange of deltas in batches with a resume cursor, media files, and the
// merge and replace flows. It works on a Conn of messages and a Collection,
// without networking, so both directions and three-device convergence are
// tested in memory.
package sync

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"

	"cardflip/golang/core"
	"cardflip/golang/internal/usermsg"
)

// ProtocolVersion is the version of the sync protocol this build speaks;
// MinProtocolVersion the oldest it can still speak.
const (
	ProtocolVersion    = 1
	MinProtocolVersion = 1
)

// MaxClockSkew is the largest difference between two devices' clocks that
// they sync with (spec: Clocks must agree).
const MaxClockSkew = 10 * time.Minute

// Choice is the user's answer to the first-sync question, from the point
// of view of the device that asked.
type Choice string

const (
	ChoiceNone Choice = ""
	// ChoiceMerge combines both collections.
	ChoiceMerge Choice = "merge"
	// ChoiceReplaceOther replaces the peer's collection with this one.
	ChoiceReplaceOther Choice = "replaceOther"
	// ChoiceReplaceThis replaces this collection with the peer's.
	ChoiceReplaceThis Choice = "replaceThis"
)

// mirrored returns the choice as the peer sees it.
func (c Choice) mirrored() Choice {
	switch c {
	case ChoiceReplaceOther:
		return ChoiceReplaceThis
	case ChoiceReplaceThis:
		return ChoiceReplaceOther
	}
	return c
}

func (c Choice) valid() bool {
	return c == ChoiceNone || c == ChoiceMerge || c == ChoiceReplaceOther || c == ChoiceReplaceThis
}

// Link is what a profile remembers about its link to a profile on a paired
// device. The app stores it next to the profile's collection, outside
// backups.
type Link struct {
	ID            string `json:"id"`
	Peer          string `json:"peer"`
	RemoteProfile string `json:"remoteProfile"`
	// Epoch and PeerEpoch are both collections' lineages at the last
	// sync; empty before the first.
	Epoch     string `json:"epoch,omitempty"`
	PeerEpoch string `json:"peerEpoch,omitempty"`
	// Received is the peer's sequence number up to which its changes have
	// all been received (-1 for none). Cursor is the position in the
	// peer's delta after Received that an interrupted session got to, and
	// CursorSeq the peer's sequence number when that delta started, which
	// becomes Received once it is complete: changes after it, before the
	// cursor, come with the next delta.
	Received  int64             `json:"received"`
	Cursor    *core.DeltaCursor `json:"cursor,omitempty"`
	CursorSeq int64             `json:"cursorSeq,omitempty"`
	// Echo are the local sequence numbers of changes received through the
	// link, which are not sent back.
	Echo core.SeqRanges `json:"echo,omitempty"`
	// PeerReceived is this collection's sequence number up to which the
	// peer has received everything, and LastSeq this collection's sequence
	// number when the last sync finished.
	PeerReceived int64     `json:"peerReceived"`
	LastSeq      int64     `json:"lastSeq"`
	LastSynced   time.Time `json:"lastSynced,omitzero"`
	// Pending is a first-sync choice that is being carried out; a session
	// interrupted in the middle of it continues it. NeedsChoice is set
	// while the first-sync question waits for the user.
	Pending     Choice `json:"pending,omitempty"`
	NeedsChoice bool   `json:"needsChoice,omitempty"`
}

// NewLink returns a link that has never synced.
func NewLink(id, peer, remoteProfile string) Link {
	return Link{ID: id, Peer: peer, RemoteProfile: remoteProfile, Received: -1, PeerReceived: -1}
}

// firstSync reports whether the next sync of the link must ask the user
// how to combine the collections (decision 7): it never synced, a
// collection was replaced since, or it has not synced for longer than
// tombstones are kept.
func (l *Link) firstSync(epoch, peerEpoch string, now time.Time) bool {
	return l.Epoch == "" || l.Epoch != epoch || l.PeerEpoch != peerEpoch ||
		now.Sub(l.LastSynced) > core.TombstoneRetention
}

// Hello opens a session. Link names the link; Epoch is the sender's
// lineage and SeenEpoch the receiver's as the sender saw it at the last
// sync; FirstSync says whether the sender's link needs the first-sync
// question, and Choice is the initiator's answer to it, from the
// initiator's point of view.
type Hello struct {
	Protocol    int            `json:"protocol"`
	MinProtocol int            `json:"minProtocol"`
	Snapshot    int            `json:"snapshot"`
	Device      string         `json:"device"`
	Name        string         `json:"name"`
	Time        int64          `json:"time"`
	Link        string         `json:"link"`
	Epoch       string         `json:"epoch"`
	SeenEpoch   string         `json:"seenEpoch,omitempty"`
	FirstSync   bool           `json:"firstSync,omitempty"`
	Choice      Choice         `json:"choice,omitempty"`
	Slots       map[string]int `json:"slots,omitempty"`
}

// Message is one message of a session.
type Message struct {
	Type       string            `json:"type"`
	Hello      *Hello            `json:"hello,omitempty"`
	Refusal    *Refusal          `json:"refusal,omitempty"`
	Since      int64             `json:"since,omitempty"`
	After      *core.DeltaCursor `json:"after,omitempty"`
	Page       *core.DeltaPage   `json:"page,omitempty"`
	Name       string            `json:"name,omitempty"`
	SHA1       string            `json:"sha1,omitempty"`
	Names      []string          `json:"names,omitempty"`
	Data       []byte            `json:"data,omitempty"`
	More       bool              `json:"more,omitempty"`
	Entries    []json.RawMessage `json:"entries,omitempty"`
	Identities *core.Identities  `json:"ids,omitempty"`
	Seq        int64             `json:"seq,omitempty"`
	Epoch      string            `json:"epoch,omitempty"`
	Count      int               `json:"count,omitempty"`
	Have       int64             `json:"have,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// Message types.
const (
	msgHello      = "hello"
	msgRefuse     = "refuse"
	msgDeltaReq   = "deltaReq"
	msgDelta      = "delta"
	msgTurn       = "turn"
	msgDone       = "done"
	msgMediaReq   = "mediaReq"
	msgMedia      = "media"
	msgIdentities = "identities"
	msgRenumbered = "renumbered"
	msgRevlog     = "revlog"
	msgSnapshot   = "snapshot"
	msgBye        = "bye"
	msgFail       = "fail"
)

// Refusal is why a device refused to sync. It is an error, with a message
// for the user.
type Refusal struct {
	Code string `json:"code"`
	// Device is the device to update (version), or the other device
	// (clock, needsChoice).
	Device string `json:"device,omitempty"`
	// Skew is the difference between the clocks, in milliseconds.
	Skew int64 `json:"skew,omitempty"`
	// Mine is set when the device to update is the one showing the error.
	Mine bool `json:"mine,omitempty"`
}

// Refusal codes.
const (
	RefuseVersion     = "version"
	RefuseClock       = "clock"
	RefuseNeedsChoice = "needsChoice"
	RefuseUnknownLink = "unknownLink"
	RefuseBusy        = "busy"
)

func (r *Refusal) Error() string {
	switch r.Code {
	case RefuseVersion:
		if r.Mine {
			return usermsg.Text("syncUpdateThis", "Update Cardflip on this device to sync with %s.", r.Device)
		}
		return usermsg.Text("syncUpdateOther", "Update Cardflip on %s to sync with it.", r.Device)
	case RefuseClock:
		d := (time.Duration(r.Skew) * time.Millisecond).Abs()
		switch {
		case d >= 48*time.Hour:
			return usermsg.Text("syncClocksDifferDays", "The clocks of this device and %s differ by %d days. Set the correct time on both and try again.",
				r.Device, int(d.Round(24*time.Hour)/(24*time.Hour)))
		case d >= 90*time.Minute:
			return usermsg.Text("syncClocksDifferHours", "The clocks of this device and %s differ by %d hours. Set the correct time on both and try again.",
				r.Device, int(d.Round(time.Hour)/time.Hour))
		default:
			return usermsg.Text("syncClocksDifferMinutes", "The clocks of this device and %s differ by %d minutes. Set the correct time on both and try again.",
				r.Device, int(d.Round(time.Minute)/time.Minute))
		}
	case RefuseNeedsChoice:
		return usermsg.Text("syncNeedsChoice", "Choose how to combine this profile with the one on %s.", r.Device)
	case RefuseUnknownLink:
		return usermsg.Text("syncUnknownLink", "%s no longer has this link. Link the profiles again.", r.Device)
	case RefuseBusy:
		return usermsg.Text("syncBusy", "%s is syncing with another device. Try again in a moment.", r.Device)
	}
	return fmt.Sprintf("sync refused: %s", r.Code)
}

// MergeSlots merges two slot tables (decision 4) and gives the devices that
// have none a slot. Both devices of a session compute it from the same
// tables and get the same result: a device with two slots keeps the
// smaller; of two devices with one slot, the one with the larger ID moves
// to the lowest free slot; new devices take the lowest free slots in the
// order of their IDs. With all 32 slots taken, a device stays without one.
func MergeSlots(a, b map[string]int, devices ...string) map[string]int {
	m := maps.Clone(a)
	if m == nil {
		m = map[string]int{}
	}
	for d, s := range b {
		if old, ok := m[d]; !ok || s < old {
			m[d] = s
		}
	}
	owner := map[int]string{}
	var homeless []string
	for _, d := range slices.Sorted(maps.Keys(m)) {
		s := m[d]
		if _, taken := owner[s]; taken || s < 0 || s > core.MaxSlot {
			delete(m, d)
			homeless = append(homeless, d)
			continue
		}
		owner[s] = d
	}
	for _, d := range devices {
		if _, ok := m[d]; !ok && !slices.Contains(homeless, d) && d != "" {
			homeless = append(homeless, d)
		}
	}
	slices.SortFunc(homeless, cmp.Compare)
	for _, d := range slices.Compact(homeless) {
		for s := 0; s <= core.MaxSlot; s++ {
			if _, taken := owner[s]; !taken {
				owner[s], m[d] = d, s
				break
			}
		}
	}
	return m
}
