package middleware

import (
	"testing"
	"unicode"
)

func Test_generateNonce(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Run("generateNonce", func(t *testing.T) {
			nonce, err := generateNonce()
			if err != nil {
				t.Errorf("generateNonce() returned error: %v", err)
				return
			}
			for _, c := range nonce {
				if !unicode.IsDigit(c) && (c < 'A' || c > 'F') {
					t.Errorf("generateNonce() %s, is not HEXA value", nonce)
				}
			}
		})
	}
}
