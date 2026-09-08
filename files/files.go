// GET OR CREATE DATABASE FILES
package files

import (
	"errors"
	"os"
	"strings"
)

var (
	DB_DRIVER           string = "sqlite3"
	DB_FILEPATH_ENV_VAR string = "CHAR_INV_DATABASE"
	DB_FILE_EXTENSION   string = ".db"
	XDG_DATA_DIR        string = "$HOME/.local/share/char-inv/char-inv.db"

	FILE_ERRORS = map[string]error{
		"DoesNotExist":        errors.New("File %s does not exist\n"),
		"EmptyFileString":     errors.New("String provided for database filepath is empty\n"),
		"FileNotFound":        errors.New("No database files with file extentsion %s found in directory %s\n"),
		"NoWritePermsissions": errors.New("You do not have write permissions for file %s\n"),
	}
)

type FileHandler struct {
	Filepath      string
	FileExtension string
	Driver        string
}

func (f *FileHandler) Init(path ...string) error {
	f.FileExtension = DB_FILE_EXTENSION
	f.Driver = DB_DRIVER
	if len(path) > 0 {
		f.Filepath = path[0]
	}
	err := f.GetDatabaseFilepath()
	// TODO: Figure out how best to handle errors here
	return err
}

func (f *FileHandler) CreateDatabaseFile() error {
	// directory := f.DatabaseFilepath
	return nil
}

func (f *FileHandler) GetDatabaseFilepath() error {

	// IF f.Filepath NOT PROVIDED DURING f.Init() FIND FILEPATH
	if f.Filepath == "" {
		// TODO: make process for searching for db file
		directoryList := []string{
			DB_FILEPATH_ENV_VAR,
			XDG_DATA_DIR,
		}

		for _, directory := range directoryList {
			if strings.Contains(directory, "$") {
				directory = os.ExpandEnv(directory)
			}
		}
	}

	// VALIDATE PROVIDED FILEPATH BEFORE RETURN
	v := new(FileValidator{filepath: f.Filepath})
	v.ValidateFilepath()

	if v.ErrorMessage != nil {
		return v.ErrorMessage
	} else {
		return nil
	}
}

type FileValidator struct {
	ErrorMessage error
	filepath     string
}

func (v *FileValidator) checkFileString() {
	if v.filepath == "" {
		v.ErrorMessage = FILE_ERRORS["EmptyFileString"]
	}
	return
}

func (v *FileValidator) checkFileExists() {
	_, err := os.Stat(v.filepath)
	if err != nil {
		v.ErrorMessage = FILE_ERRORS["DoesNotExist"]
	} 
	return
}

func (v *FileValidator) checkFileWritePermission() {
	fileInfo, err := os.Stat(v.filepath)
	if err != nil {
		v.ErrorMessage = err
	}
	if fileInfo.Mode().Perm() < 0o600 {
		// 0o600 IS OCTAL LITERAL EQUIVALENT TO USER WRITE PERMISSIONS
		v.ErrorMessage = FILE_ERRORS["NoWritePermsissions"]
	}
	return
}

func (v *FileValidator) ValidateFilepath() {
	var tests = []func(){
		v.checkFileString,
		v.checkFileExists,
		v.checkFileWritePermission,
	}
	// LOOP THROUGH TESTS AND BREAK IF ONE SETS AN ERROR != nil
	for _, test := range tests {
		test()
		if v.ErrorMessage != nil {
			break
		}
	}
	return
}
