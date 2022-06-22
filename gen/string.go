package gen

import (
	"math/rand"
	"strings"
	"time"
)

var alphabet = "abcdefghijklmnopqrstuvwxyz"

const defaultLength = 10

// RandLowercaseString returns a random string of length "defaultLength" made from alphabet.
func RandLowercaseString() string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	var sb strings.Builder
	sb.Grow(defaultLength)
	for i := 0; i < defaultLength; i++ {
		sb.WriteByte(alphabet[r.Intn(len(alphabet))])
	}
	return sb.String()
}
