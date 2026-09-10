// GET OR CREATE DATABASE FILES
package files

import (
	"fmt"
	"os"
	"strings"
)

const (
	DB_DRIVER           string = "sqlite3"
	DB_FILEPATH_ENV_VAR string = "CHAR_INV_DATABASE"
	DB_FILE_EXTENSION   string = ".db"
	XDG_DATA_DIR        string = "$HOME/.local/share/char-inv/char-inv.db"

	// FILE ERRORS
	ErrEmptyEnvVar		string = "Set environment variable '%s' is empty"
	ErrFileNotExist		string = "stat %s: no such file or directory"
	ErrIrregularFile	string = "File '%s' is not a regular file"
	ErrNotDatabase		string = "File %s is not a database with file extentsion %s"
	ErrNonWriteable 	string = "You do not have write permissions for file '%s'"
	ErrTargetDirectory	string = "%s is a directory not a file"
	ErrXDGNotCreated	string = "Defaulted to XDG database (%s) but it hasn't been created yet.\nUse 'character-inventory new' to generate new database"
)

// KEEP TRACK OF HOW DATABASE FILE WAS ACQUIRED WITH enum
type HandlerMode int
const (
	XDG HandlerMode = iota
	ENV
	ARG	
)

func (h HandlerMode) String() string {
	return []string{"XDG", "ENV", "ARG"}[h]
}

type FileHandler struct {
	Driver        string
	ErrorMessage  error
	Filepath      string
	FileExtension string
	Mode          HandlerMode 
}

func (f *FileHandler) affirmDatabase(fileInfo os.FileInfo) {
	if !strings.HasSuffix(fileInfo.Name(), f.FileExtension) {
		f.ErrorMessage = fmt.Errorf(ErrNotDatabase, f.Filepath, f.FileExtension)
	}
}

func (f *FileHandler) affirmNotDirectory(fileInfo os.FileInfo) {
	if fileInfo.IsDir() {
		f.ErrorMessage = fmt.Errorf(ErrTargetDirectory, f.Filepath)
	}
}

func (f *FileHandler) affirmRegularFile(fileInfo os.FileInfo) {
	if !fileInfo.Mode().IsRegular() {
		f.ErrorMessage = fmt.Errorf(ErrIrregularFile, f.Filepath)
	}
	return
}

func (f *FileHandler) affirmWritePermission(fileInfo os.FileInfo) {
	if fileInfo.Mode().Perm() < 0o600 {
		// 0o600 IS OCTAL LITERAL EQUIVALENT TO USER WRITE PERMISSIONS
		f.ErrorMessage = fmt.Errorf(ErrNonWriteable, f.Filepath)
	}
	return
}

func (f *FileHandler) checkEdgeCase(fileErr error) (edgeErr error, overrideErr bool) {
	if fileErr.Error() == fmt.Errorf(ErrFileNotExist, XDG_DATA_DIR).Error() {
		return fmt.Errorf(ErrXDGNotCreated, XDG_DATA_DIR), true
	}
	return
}

func (f *FileHandler) Init(path ...string) {
	f.FileExtension = DB_FILE_EXTENSION
	f.Driver = DB_DRIVER

	// IF PROVIDED FILE PASSED IN FROM CMD ARGS SET f.Filepath, OTHERWISE FIND IN env OR XDG
	fpSet, argSet := len(f.Filepath) > 0, len(path) > 0
	if fpSet || argSet { // FILEPATH SET ON struct INSTANTIATION
		f.Mode = ARG
	} else if !fpSet && argSet { // FILEPATH PASSED INTO FUNC
		f.Filepath = path[0]
	} else if !fpSet && !argSet { // FILEPATH NOT SET ON struct INSTANTIATION
		f.SearchForDatabase()
	}

	// VALIDATE PROVIDED FILEPATH BEFORE RETURN
	f.ValidateFilepath()

	// CHECK FOR ERRORS SET TO FileHandler
	if f.ErrorMessage != nil {
		fmt.Println(f.ErrorMessage)
		os.Exit(1)
	}

	return
}

func (f *FileHandler) CreateDatabaseFile(desiredFilepath string) error {
	// directory := f.DatabaseFilepath
	// TODO: make proccess for file creation
	return nil
}

// LOOK IN ENV VAR OR XDG_DATA_DIR FOR DATABASE FILE
func (f *FileHandler) SearchForDatabase() {
	env, envSet := os.LookupEnv(DB_FILEPATH_ENV_VAR)
	if envSet {
		f.Mode = ENV
		if env == "" {
			f.ErrorMessage = fmt.Errorf(ErrEmptyEnvVar, DB_FILEPATH_ENV_VAR)
			return
		}
	}

	// PICK BETWEEN DB_FILEPATH_ENV_VAR & XDG_DATA_DIR
	var targetFilepath string
	if envSet {
		targetFilepath = env
	} else {
		targetFilepath = XDG_DATA_DIR
	}

	// SET FILEPATH - MOST VALIDATION HANDLED BY FileValidator.ValidateFilepath()
	f.Filepath = targetFilepath
	return
}

func (f *FileHandler) ValidateFilepath() {
	// err OUPUT OF os.Stat WILL HANDLE FILES THAT DO NOT EXIST
	fileInfo, err := os.Stat(f.Filepath)
	if err != nil {
		edgeErr, overrideErr := f.checkEdgeCase(err)
		if overrideErr {
			err = edgeErr
		}
		f.ErrorMessage = err
		return
	}

	var tests = []func(os.FileInfo){
		f.affirmNotDirectory,
		f.affirmRegularFile,
		f.affirmWritePermission,
		f.affirmDatabase,
	}
	// LOOP THROUGH TESTS AND BREAK IF ONE SETS AN ERROR != nil
	for _, test := range tests {
		test(fileInfo)
		if f.ErrorMessage != nil {
			break
		}
	}
	return
}
