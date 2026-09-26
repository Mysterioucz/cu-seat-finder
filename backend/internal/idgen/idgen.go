// Package idgen makes short random IDs for primary keys.
// ponytail: crypto/rand hex string instead of a uuid dependency — plenty unique for this app's scale.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
)

func New() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Reference makes a short, human-typeable booking reference like "CU-7F3K9Q".
func Reference() string {
	const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ" // no 0/O/1/I ambiguity
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	out := make([]byte, 6)
	for i, c := range b {
		out[i] = alphabet[int(c)%len(alphabet)]
	}
	return "CU-" + string(out)
}
