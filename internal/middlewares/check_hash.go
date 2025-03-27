package middlewares

import (
	"bytes"
	"crypto/sha256"
	"io"
	"net/http"
)

const (
	signed = "HashSHA256"
)

func CheckHash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hash := r.Header.Get(signed)

		if hash != "" {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}

			if !isCorrectHash(hash, bodyBytes) {
				http.Error(w, "invalid hash", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		next.ServeHTTP(w, r)
	})
}

func isCorrectHash(hash string, bodyBytes []byte) bool {
	bodyHash := sha256.Sum256(bodyBytes)

	return bytes.Equal(bodyHash[:], []byte(hash))
}
