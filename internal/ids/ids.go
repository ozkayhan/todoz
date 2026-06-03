// Package ids generates unique identifiers for tasks and lists.
//
// IDs are of the form "<prefix>-<32 hex chars>" using 16 cryptographically
// random bytes. Random IDs (rather than sequential counters) keep generation
// lock-free and collision-free across concurrent todoz processes.
package ids

import (
	"crypto/rand"
	"encoding/hex"
)

// New returns a unique id with the given semantic prefix (e.g. "task", "list").
func New(prefix string) string {
	var b [16]byte
	// crypto/rand.Read never returns a short read on supported platforms;
	// any error here is catastrophic (no entropy) and worth surfacing.
	if _, err := rand.Read(b[:]); err != nil {
		panic("ids: cannot read random bytes: " + err.Error())
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}
