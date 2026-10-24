package syncnet

// Requests between paired devices besides sync sessions: listing a
// device's profiles and creating and removing links (device-sync spec:
// Profile links). Each is the first message of a connection, answered by
// one message.
const (
	MsgListProfiles = "listProfiles"
	MsgProfiles     = "profiles"
	MsgCreateLink   = "createLink"
	MsgLinked       = "linked"
	MsgUnlink       = "unlink"
	MsgFailed       = "failed"
)

// ProfileInfo names a profile in a MsgProfiles answer (as JSON in Data).
type ProfileInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateLinkRequest asks a paired device to link profile Profile, or a new
// profile named NewName, with profile FromProfile (named FromName) of the
// asking device. It is the JSON Data of a MsgCreateLink message; the
// answer's Data is the linked ProfileInfo.
type CreateLinkRequest struct {
	LinkID      string `json:"linkId"`
	Profile     string `json:"profile,omitempty"`
	NewName     string `json:"newName,omitempty"`
	FromProfile string `json:"fromProfile"`
	FromName    string `json:"fromName"`
}
