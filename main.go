package main

import (
	"os"
	"strings"

	"github.com/yv3000/pathman_doctor/cmd"
	"github.com/yv3000/pathman_doctor/internal"
)

var version = "1.3.0"

func main() {
	app := cmd.NewApp(internal.NewStore(), os.Stdin, os.Stdout, normalizeVersion(version))
	os.Exit(app.Run(os.Args[1:]))
}

// normalizeVersion accepts both "1.2.3" and the "v1.2.3" form injected from
// a release tag, so --version never prints a doubled prefix.
func normalizeVersion(value string) string {
	return strings.TrimPrefix(value, "v")
}
