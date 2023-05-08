package main

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
)

const (
	StorageProviderLogo = `Greenfield Storage Provider
    __                                                       _     __
    _____/ /_____  _________ _____ ____     ____  _________ _   __(_)___/ /__  _____
    / ___/ __/ __ \/ ___/ __  / __  / _ \   / __ \/ ___/ __ \ | / / / __  / _ \/ ___/
    (__  ) /_/ /_/ / /  / /_/ / /_/ /  __/  / /_/ / /  / /_/ / |/ / / /_/ /  __/ /
    /____/\__/\____/_/   \__,_/\__, /\___/  / .___/_/   \____/|___/_/\__,_/\___/_/
    /____/       /_/
    `
)

// DumpLogo output greenfield storage provider logo
func DumpLogo() string {
	return StorageProviderLogo
}

var (
	Version    string
	CommitID   string
	BranchName string
	BuildTime  string
)

// DumpVersion output the storage provider version information
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

var VersionCmd = &cli.Command{
	Action:      versionAction,
	Name:        "version",
	Aliases:     []string{"v"},
	Usage:       "Print version information",
	Category:    "MISCELLANEOUS COMMANDS",
	Description: `The output of this command is supposed to be machine-readable.`,
}

func versionAction(ctx *cli.Context) error {
	fmt.Print(DumpLogo() + "\n" + DumpVersion())
	return nil
}
