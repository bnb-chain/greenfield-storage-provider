package main

import (
	"fmt"
	"os"
	"strings"

	linenoise "github.com/GeertJohan/go.linenoise"
	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/cmd/gnfd_sp_cli/conf"
	"github.com/bnb-chain/greenfield-storage-provider/cmd/gnfd_sp_cli/crypto"
)

const (
	HistoryFile = ".gnfd_sp_cli_history"
)

var commands = map[string]cli.Command{}

var confCommands = []cli.Command{
	conf.DumpConfigCommand,
}

var addrCommands = []cli.Command{
	crypto.CreateSpAddressCommand,
	crypto.CreateP2PCommand,
}

func init() {
	for _, cmd := range confCommands {
		commands[cmd.Name] = cmd
		if len(cmd.ShortName) != 0 {
			commands[cmd.ShortName] = cmd
		}
	}
	for _, cmd := range addrCommands {
		commands[cmd.Name] = cmd
		if len(cmd.ShortName) != 0 {
			commands[cmd.ShortName] = cmd
		}
	}
}

func showHelp() {
	fmt.Println("config commands:")
	for _, cmd := range confCommands {
		fmt.Printf("%-3s%-14s  -  %-50s\n", "   ", cmd.Name, cmd.Usage)
	}
	fmt.Println("crypto commands:")
	for _, cmd := range addrCommands {
		fmt.Printf("%-3s%-14s  -  %-50s\n", "   ", cmd.Name, cmd.Usage)
	}
}

func main() {
	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		showHelp()
		os.Exit(0)
	}

	_, err := os.Stat(HistoryFile)
	if err != nil && os.IsNotExist(err) {
		_, err = os.Create(HistoryFile)
		if err != nil {
			fmt.Println("fail to create command history file, error:", err)
			os.Exit(1)
		}
	}
	err = linenoise.LoadHistory(HistoryFile)
	if err != nil {
		fmt.Println("fail to load command history file, error:", err)
	}
	for {
		str, err := linenoise.Line("gnfd_sp_cli>")
		if err != nil {
			if err == linenoise.KillSignalError {
				os.Exit(1)
			}
			fmt.Printf("Unexpected error: %s\n", err)
			os.Exit(1)
		}
		linenoise.AddHistory(str)
		linenoise.SaveHistory(HistoryFile)
		fields := strings.Fields(str)
		if len(fields) == 0 {
			continue
		}

		switch fields[0] {
		case "help":
			showHelp()
			continue
		case "quit":
			os.Exit(0)
		}

		cmd, ok := commands[fields[0]]
		if !ok {
			fmt.Println("Error: unknown command.")
			continue
		}
		app := cli.NewApp()
		app.Name = cmd.Name
		app.Commands = []cli.Command{cmd}
		app.Run(append(os.Args[:1], fields...))
	}
}
