package keystore

// Spec says what key to make or to take: its name before the algorithm,
// as deploy for deploy_rsa4096, its purpose and scope, and for a key to
// generate its algorithm and size. An upload reads them from the key.
type Spec struct {
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
	Scope   string `json:"scope"`
	Algo    string `json:"algo"`
	Bits    int    `json:"bits"`
	Comment string `json:"comment"`
}
