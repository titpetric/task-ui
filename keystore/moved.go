package keystore

// Moved is a key the migration took, by its old file, its new name and
// its fingerprint.
type Moved struct {
	From, Name, Fingerprint string
}
