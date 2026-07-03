package main

import (
	"os"

	"github.com/deepawasthi/devstack/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
