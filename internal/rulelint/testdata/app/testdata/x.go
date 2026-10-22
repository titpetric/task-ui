package x

func f(s interface{ Allowed(string) bool }) { s.Allowed("ignored/view") }
