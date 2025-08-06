package application

import (
	"fmt"
	"runtime"
)

type Version struct {
	AppVersion string `json:"version,omitempty"`
	Commit     string `json:"commit,omitempty"`
	BuildDate  string `json:"buildDate,omitempty"`
	GoVersion  string `json:"goVersion,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

func NewVersion(appVer, hash, date string) *Version {
	return &Version{
		AppVersion: appVer,
		Commit:     hash,
		BuildDate:  date,
		GoVersion:  runtime.Version(),
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
