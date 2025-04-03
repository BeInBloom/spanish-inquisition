package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	signed = "HashSHA256"
)

func CheckHash(key string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		f := func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("CheckHash: %+v\n", r.Header)
			hash := r.Header.Get(signed)

			if hash != "" {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}

				if !isCorrectHash(hash, bodyBytes, key) {
					http.Error(w, "invalid hash", http.StatusBadRequest)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			fmt.Printf("all fine\n")
			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(f)
	}

}

func isCorrectHash(hash string, bodyBytes []byte, key string) bool {
	h := hmac.New(sha256.New, []byte(key))

	h.Write(bodyBytes)

	bodyHash := h.Sum(nil)

	hexHash := hex.EncodeToString(bodyHash)

	fmt.Printf("hash: %s, hexHash: %s\n", hash, hexHash)

	return strings.EqualFold(hash, hexHash)
}
