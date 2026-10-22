package app

type set struct{}

func (set) Allowed(path string) bool { return path != "" }

type handlers struct{ perms set }

// can hands its rule on: no finding.
func (h handlers) can(path string) bool { return h.perms.Allowed(path) }

func (h handlers) routes(dynamic string) {
	_ = h.can("note/view")
	_ = h.perms.Allowed("note/edit")
	_ = h.can("note/write")
	_ = h.can("note/all")
	_ = h.can(dynamic)
	_ = h.perms.Allowed("run")
}
