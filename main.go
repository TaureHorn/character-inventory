package main

import (
	"fmt"
	"github.com/TaureHorn/character-inventory/files"
	"os"
)

const helpString = `character-inventory
	
[Usage]
character-inventory [file]
character-inventory [command]

[Commands]
character-inventory new
character-inventory help
`

func createFile(index int, args []string) {
	f := new(files.FileHandler)
	f.GetEnvironmentVariable()

	if len(args) >= index+2 {
		f.Mode = files.ARG
		f.Filepath = args[index+1]
	} else if f.EnvVarSet {
		f.Mode = files.ENV
		f.Filepath = f.EnvVar
	} else {
		f.Filepath = files.XDG_DATA_DIR
	}
	f.CreateDatabaseFile()
}

// ITERATE OVER CMD ARGS FIND COMMANDS OR PROVIDED DATABASE FILEPATH
func parseCmdArguments(args []string) (string, bool) {

	if len(args) <= 1 {
		return "", false
	}

	var filepath string
	for i, flag := range args {
		switch flag {
		case "help":
			printHelp()
		case "new":
			createFile(i, args)
		default:
			filepath = flag
		}
	}

	if len(filepath) == 0 {
		return "", false
	} else {
		return filepath, true
	}

}

// PRINT const helpString
func printHelp() {
	fmt.Print(helpString)
	os.Exit(0)
	return
}

func main() {
	fileHandler := new(files.FileHandler)
	databaseFile, providedFilepath := parseCmdArguments(os.Args)

	if providedFilepath {
		fileHandler.Init(databaseFile)
	} else {
		fileHandler.Init()
	}

}
