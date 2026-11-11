package folder

// Catalogue lists the folders and finds one by ID. Discovery is one; the
// composition builds a fresh one per request, so configuration changes
// show without a restart.
type Catalogue interface {
	List() []*Folder
	Find(id string) *Folder
}
