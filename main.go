package main

import (
	"fmt"
	"os"

	"github.com/folcoz/milestone1-code/secrets"
	"github.com/folcoz/milestone1-code/server"
)

func main() {
	err := secrets.InitFile()
	if err != nil {
		quit(err, 1)
	}

	server.StartListener("0.0.0.0:8080")
}

func quit(err error, exitCode int) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(exitCode)
}
