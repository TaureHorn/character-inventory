package main

import (
	"fmt"
	"os"

	"github.com/TaureHorn/character-inventory/files"
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
	creationErr := f.CreateDatabaseFile()
	if creationErr != nil {
		fmt.Println(creationErr)
		os.Exit(1)
	}

	d := new(files.DataHandler)
	d.Init(files.DB_DRIVER, f.Filepath, files.CREATE)
	d.WriteDefaultData()

	if d.ErrorMessage != nil {
		fmt.Println(d.ErrorMessage)
		os.Exit(1)
	}

	os.Exit(0)
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
	// Get FileHandler
	fileHandler := new(files.FileHandler)
	databaseFile, providedFilepath := parseCmdArguments(os.Args)

	var initErr error
	if providedFilepath {
		initErr = fileHandler.Init(databaseFile)
	} else {
		initErr = fileHandler.Init()
	}
	if initErr != nil {
		fmt.Println(initErr)
		os.Exit(1)
	}

	// Connect to Database
	dataHandler := new(files.DataHandler)
	dataHandler.Init(files.DB_DRIVER, fileHandler.Filepath, fileHandler.Context)
	defer dataHandler.Database.Close()
}
