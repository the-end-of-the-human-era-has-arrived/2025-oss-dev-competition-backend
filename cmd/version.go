package main

import (
	"fmt"
	"runtime"
	"time"
)

var (
	version    = "unknown"
	commitHash = "unknown"
	buildDate  = time.Now().Format(time.RFC3339)
)

type Version struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

func NewVersion() *Version {
	return &Version{
		Version:   version,
		Commit:    commitHash,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func (v *Version) String() string {
	return fmt.Sprintf("%+#v", *v)
}
