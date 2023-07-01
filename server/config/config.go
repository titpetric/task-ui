package config

// Load function exposed from TaskfileLoader as a global ref.
var Load = (*TaskfileLoader)(nil).Load
