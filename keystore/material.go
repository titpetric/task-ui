package keystore

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

// generate makes private key material.
func generate(algo string, bits int) (any, int, error) {
	switch algo {
	case "ed25519":
		_, key, err := ed25519.GenerateKey(rand.Reader)
		return key, 0, err
	case "ecdsa":
		curves := map[int]elliptic.Curve{0: elliptic.P256(), 256: elliptic.P256(), 384: elliptic.P384(), 521: elliptic.P521()}
		curve, ok := curves[bits]
		if !ok {
			return nil, 0, fmt.Errorf("ecdsa is 256, 384 or 521 bits, not %d", bits)
		}
		key, err := ecdsa.GenerateKey(curve, rand.Reader)
		return key, curve.Params().BitSize, err
	case "rsa":
		if bits == 0 {
			bits = 4096
		}
		if bits != 2048 && bits != 3072 && bits != 4096 {
			return nil, 0, fmt.Errorf("rsa is 2048, 3072 or 4096 bits, not %d", bits)
		}
		key, err := rsa.GenerateKey(rand.Reader, bits)
		return key, bits, err
	}
	return nil, 0, fmt.Errorf("unsupported key algorithm %q: ed25519, ecdsa or rsa", algo)
}

// kindOf returns the algorithm and size of private key material.
func kindOf(private any) (string, int, error) {
	switch key := private.(type) {
	case ed25519.PrivateKey:
		return "ed25519", 0, nil
	case *ecdsa.PrivateKey:
		return "ecdsa", key.Curve.Params().BitSize, nil
	case *rsa.PrivateKey:
		return "rsa", key.N.BitLen(), nil
	}
	return "", 0, fmt.Errorf("unsupported key type %T", private)
}
