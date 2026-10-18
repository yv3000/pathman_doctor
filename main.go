package main

import (
	"os"

	"github.com/yv3000/pathman_doctor/cmd"
	"github.com/yv3000/pathman_doctor/internal"
)

var version = "1.1.0"

func main() {
	app := cmd.NewApp(internal.NewStore(), os.Stdin, os.Stdout, version)
	os.Exit(app.Run(os.Args[1:]))
}
