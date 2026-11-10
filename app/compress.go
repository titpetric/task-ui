package app

import (
	"net/http"
)

// uncompressedRanges wraps the compressor so a range request passes it by.
//
// A range of a compressed body is a range of the gzip stream under an
// unchanged Content-Range, so a client resuming a download of an artifact
// writes compressed bytes at the offsets of the plain ones.
func uncompressedRanges(compress func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		compressed := compress(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Range") != "" {
				next.ServeHTTP(w, r)
				return
			}
			compressed.ServeHTTP(w, r)
		})
	}
}
