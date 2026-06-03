// Command todoz is a standalone, append-only todo store exposed as a CLI.
//
// Apps invoke it as a subprocess: `todoz <command> [--flag value ...]`.
// It prints a single JSON envelope to stdout and exits 0 on success or
// non-zero on failure. See USAGE.md for the full integration contract.
package main

import (
	"fmt"
	"os"
	"todoz/internal/cli"
)

func main() {
	out, code := cli.Run(os.Args[1:])
	fmt.Println(out)
	os.Exit(code)
}
