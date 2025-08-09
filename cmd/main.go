package main

import (
	"log"
	"time"

	"github.com/joho/godotenv"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/application"
)

var (
	version    = "unknown"
	commitHash = "unknown"
	buildDate  = time.Now().Format(time.RFC3339)
)

func main() {
	_ = godotenv.Load()

	ver := application.NewVersion(version, commitHash, buildDate)
	app := application.NewCLI(ver)

	if err := app.Execute(); err != nil {
		log.Fatal(err)
	}
}
