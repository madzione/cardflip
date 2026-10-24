package storage

// DefaultProfile is the ID of the profile whose data is where the collection
// was kept before there were profiles, so an existing installation opens it
// unchanged.
const DefaultProfile = "default"

// The profile registry (which profiles exist, which was used last) is kept
// outside every collection: ReadProfiles and WriteProfiles store it next to
// the default profile's data, and deleting a profile's data never touches
// it. ProfileDataDir says where each profile's backend is.
