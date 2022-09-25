package path

import (
	"math/rand"
	"unsafe"
)

const randLetters = "abcdefghijklmnopqrstuvwxyz1234567890"

func generateUniqueId(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = randLetters[rand.Int63()%int64(len(randLetters))]
	}
	return *(*string)(unsafe.Pointer(&b))
}
