package main

import (
	"fmt"
	"runtime"
)

var (
	Version    string
	CommitID   string
	BranchName string
	BuildTime  string
)

func DumpVersion() string {
	return fmt.Sprintf("Version : %s\n"+
		"Branch  : %s\n"+
		"Commit  : %s\n"+
		"Build   : %s %s %s %s\n",
		Version,
		BranchName,
		CommitID,
		runtime.Version(), runtime.GOOS, runtime.GOARCH, BuildTime)
}
