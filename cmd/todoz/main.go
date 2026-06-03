// Command todoz is a standalone, append-only todo store exposed as a CLI.
//
// Apps invoke it as a subprocess: `todoz <command> [--flag value ...]`.
// It prints a single JSON envelope to stdout and exits 0 on success or
// non-zero on failure. See USAGE.md for the full integration contract.
package main

import "os"

func main() {
	// Wired to the CLI router in Task 26. For now it is an inert skeleton
	// so the module compiles from the very first commit.
	os.Exit(0)
}
