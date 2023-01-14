package main

import (
	"fmt"
	"os"
	"strings"

	linenoise "github.com/GeertJohan/go.linenoise"
	"github.com/urfave/cli"

	"github.com/bnb-chain/greenfield-storage-provider/test/test_tool/context"

	"github.com/bnb-chain/greenfield-storage-provider/test/test_tool/common"
	stoneHub "github.com/bnb-chain/greenfield-storage-provider/test/test_tool/stonehub"
)

var rootCommands = []cli.Command{
	common.ListCommand,
	common.CdCommand,
}

var stoneHubCommands = []cli.Command{
	stoneHub.QueryStoneCommand,
	stoneHub.CreateObjectCommand,
	stoneHub.BeginUploadPayloadCommand,
	stoneHub.DonePrimaryPieceJobCommand,
	stoneHub.DoneSecondaryPieceJobCommand,
}

var commands = map[string]cli.Command{}

func init() {
	for _, cmd := range rootCommands {
		commands[cmd.Name] = cmd
		if len(cmd.ShortName) != 0 {
			commands[cmd.ShortName] = cmd
		}
	}

	for _, cmd := range stoneHubCommands {
		commands[cmd.Name] = cmd
		if len(cmd.ShortName) != 0 {
			commands[cmd.ShortName] = cmd
		}
	}
}

func showHelp() {
	fmt.Println("List of common commands help:")
	for _, cmd := range rootCommands {
		fmt.Printf("%-3s%-14s  -  %-50s\n", "   ", cmd.Name, cmd.Usage)
	}

	fmt.Println("List of stone hub commands help:")
	for _, cmd := range stoneHubCommands {
		fmt.Printf("%-3s%-14s  -  %-50s\n", "   ", cmd.Name, cmd.Usage)
	}
}

func getDir() string {
	switch ctx.CurrentService {
	case context.RootService:
		return "> "
	default:
		return "/" + ctx.CurrentService + "/> "
	}
	return "/>"
}

var ctx *context.Context

func main() {
	if len(os.Args) == 2 && (string(os.Args[1]) == "-h" || string(os.Args[1]) == "--help") {
		showHelp()
		os.Exit(0)
	}

	var (
		conf *context.CliConf
		err  error
	)
	if len(os.Args) == 3 && string(os.Args[1]) == "--config" {
		conf, err = context.LoadCliConf(os.Args[2])
	} else {
		conf = &context.CliConf{
			StoneHubAddr: context.DefaultStoneHubAddr,
			HistoryFile:  context.DefaultHistoryFile,
		}
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = os.Stat(conf.HistoryFile)
	if err != nil && os.IsNotExist(err) {
		_, err = os.Create(conf.HistoryFile)
		if err != nil {
			fmt.Println(conf.HistoryFile + " create failed")
			os.Exit(1)
		}
	}
	ctx = context.GetContext()
	ctx.Cfg = conf

	err = linenoise.LoadHistory(conf.HistoryFile)
	if err != nil {
		fmt.Println(err)
	}
	for {
		str, err := linenoise.Line(getDir())
		if err != nil {
			if err == linenoise.KillSignalError {
				os.Exit(1)
			}
			fmt.Printf("Unexpected error: %s\n", err)
			os.Exit(1)
		}
		linenoise.AddHistory(str)
		err = linenoise.SaveHistory(conf.HistoryFile)
		if err != nil {
			fmt.Println(err)
		}

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
