package bsdb

import (
	"runtime"
)

// Build Info (set via linker flags)
var (
	AppVersion    = ""
	GitCommit     = ""
	GitCommitDate = ""
)

type SpVersion struct {
	SpCodeVersion     string
	SpCodeCommit      string
	SpGoVersion       string
	SpArchitecture    string
	SpOperatingSystem string
}

func (b *BsDBImpl) GetSpVersion() *SpVersion {
	return &SpVersion{
		SpCodeVersion:     AppVersion,
		SpCodeCommit:      GitCommit,
		SpGoVersion:       runtime.Version(),
		SpArchitecture:    runtime.GOARCH,
		SpOperatingSystem: runtime.GOOS,
	}
}
