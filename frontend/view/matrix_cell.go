package view

// MatrixCell is one path of a policy: its effect, "" to inherit, and
// whether the user may do it, as Effective shows it. An all cell has no
// result of its own.
type MatrixCell struct {
	Path  string
	Title string
	// Effect is permission.Allow, permission.Deny or "".
	Effect  string
	Allowed bool
	// Checked says Allowed means something: no all cell.
	Checked bool
}
