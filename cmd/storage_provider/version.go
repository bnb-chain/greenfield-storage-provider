package main

import (
	"fmt"
	"runtime"
)

const (
	StorageProviderLogo = `Green Field Storage Provider
         __                                                       _     __         
   _____/ /_____  _________ _____ ____     ____  _________ _   __(_)___/ /__  _____
  / ___/ __/ __ \/ ___/ __  / __  / _ \   / __ \/ ___/ __ \ | / / / __  / _ \/ ___/
 (__  ) /_/ /_/ / /  / /_/ / /_/ /  __/  / /_/ / /  / /_/ / |/ / / /_/ /  __/ /
/____/\__/\____/_/   \__,_/\__, /\___/  / .___/_/   \____/|___/_/\__,_/\___/_/
                          /____/       /_/
`
)

func DumpLogo() string {
	return StorageProviderLogo
}

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
