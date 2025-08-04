package main

import (
	"fmt"
	"os"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/server"
)

func main() {
	if err := server.NewCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
