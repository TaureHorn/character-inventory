// GET OR CREATE DATABASE FILES
package files

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// CONST
const (
	DB_DRIVER           string = "sqlite3"
	DB_FILEPATH_ENV_VAR string = "CHAR_INV_DATABASE"
	DEFAULT_DB_FILEPATH string = "files/default.db"
	XDG_DATA_DIR        string = "$HOME/.local/share/char-inv/char-inv.db"

	// FILE ERRORS
	ErrAlreadyExists    string = "%s already exists"
	ErrEmptyEnvVar      string = "Set environment variable '%s' is empty"
	ErrFileNotExist     string = "stat %s: no such file or directory"
	ErrIrregularFile    string = "File '%s' is not a regular file"
	ErrPermissionDenied string = "stat %s: permission denied" // unused currently
	ErrTargetDirectory  string = "%s is a directory not a file"
	ErrXDGNotCreated    string = "Defaulted to XDG database (%s) but it hasn't been created yet.\nUse 'character-inventory new' to generate new database"
)

// KEEP TRACK OF HOW DATABASE FILE WAS ACQUIRED WITH enum
type HandlerMode int
type ValidationContext int

const (
	XDG HandlerMode = iota
	ENV
	ARG
)

const (
	CREATE ValidationContext = iota
	INIT
)

func (h HandlerMode) String() string {
	return [...]string{"XDG", "ENV", "ARG"}[h]
}

func (v ValidationContext) String() string {
	return [...]string{"CREATE", "INIT"}[v]
}

type FileHandler struct {
	Context          ValidationContext
	Driver           string
	EnvVar           string
	EnvVarSet        bool
	ErrorMessage     error
	Filepath         string
	Mode             HandlerMode
}

// VALIDATION FUNCTIONS
// Sets an error to FileHandler if a file matching passed in os.FileInfo exists
func (f *FileHandler) affirmDoesNotExist(fileInfo os.FileInfo) {
	if fileInfo != nil {
		f.ErrorMessage = errors.New(fmt.Sprintf(ErrAlreadyExists, f.Filepath))
	}
}

// Sets an error to FileHandler if file from passed in os.FileInfo is a directory
func (f *FileHandler) affirmNotDirectory(fileInfo os.FileInfo) {
	if fileInfo.IsDir() {
		f.ErrorMessage = errors.New(fmt.Sprintf(ErrTargetDirectory, f.Filepath))
	}
}

// Sets an error to FileHandler if file from passed in os.FileInfo not a regular file
func (f *FileHandler) affirmRegularFile(fileInfo os.FileInfo) {
	if !fileInfo.Mode().IsRegular() {
		f.ErrorMessage = errors.New(fmt.Sprintf(ErrIrregularFile, f.Filepath))
	}
	return
}

// Sets an error to FileHandler if user does not have write permissions for file from passed in os.FileInfo file
func (f *FileHandler) affirmWritePermission(fileInfo os.FileInfo) {
	if fileInfo.Mode().Perm() < 0o600 {
		// 0o600 IS OCTAL LITERAL EQUIVALENT TO USER WRITE PERMISSIONS
		f.ErrorMessage = errors.New(fmt.Sprintf(ErrPermissionDenied, f.Filepath))
	}
	return
}

func (f *FileHandler) checkEdgeCase(err error) (edgeErr error, overrideErr bool) {
	switch err.Error() {
	case fmt.Sprintf(ErrFileNotExist, XDG_DATA_DIR):
		return errors.New(fmt.Sprintf(ErrXDGNotCreated, XDG_DATA_DIR)), true
	case fmt.Sprintf(ErrFileNotExist, f.Filepath):
		if f.Context == CREATE {
			return nil, true
		}
	}
	return
}

// Create a new database file in a specified location
func (f *FileHandler) CreateDatabaseFile() {
	f.Context = CREATE

	// GET ABSOLUTE PATH OF f.Filepath AND VALIDATE PARENT DIRECTORY AND FILE IF EXISTS
	var procs = []func(){
		f.getFullFilepath,
		f.ValidateDirectory,
		f.ValidateFilepath,
	}
	for _, fn := range procs {
		fn()
		if f.ErrorMessage != nil {
			fmt.Println(f.ErrorMessage)
			os.Exit(1)
			break
		}
	}

	// CREATE NEW FILE, WRITE DEFAULT DATABASE DATA TO IT
	// TODO: make proccess for file creation
	newDatabase, fileCreateErr := os.Create(f.Filepath)
	if fileCreateErr != nil {
		fmt.Println(fileCreateErr)
	}
	defaultDatabase, fileReadErr := os.ReadFile(DEFAULT_DB_FILEPATH)
	if fileReadErr != nil {
		fmt.Println(fileReadErr)
	}
	newDatabase.Write(defaultDatabase)
	fmt.Printf("New file %s created!\n", newDatabase.Name())
	return
}

// Set specified environemnt variable to FileHandler as well as a boolean for if variable is set
func (f *FileHandler) GetEnvironmentVariable() {
	f.EnvVar, f.EnvVarSet = os.LookupEnv(DB_FILEPATH_ENV_VAR)
	return
}

// Modifty FileHandler.Filepath to expand environemnt variables and get absolute path
func (f *FileHandler) getFullFilepath() {
	path := os.ExpandEnv(f.Filepath)
	path, err := filepath.Abs(path)
	if err != nil {
		f.ErrorMessage = err
		return
	}
	f.Filepath = path
	return
}

// Return a slice of functions depending on FileHandler.Context
func (f *FileHandler) getValidationTests(context ValidationContext) (tests []func(os.FileInfo)) {
	switch context {
	case CREATE:
		tests = []func(os.FileInfo){
			f.affirmNotDirectory,
			f.affirmDoesNotExist,
		}
	case INIT:
		tests = []func(os.FileInfo){
			f.affirmNotDirectory,
			f.affirmRegularFile,
			f.affirmWritePermission,
		}
	}
	return
}

// Initialse FileHandler for normal use
func (f *FileHandler) Init(path ...string) {
	f.Context = INIT
	f.Driver = DB_DRIVER
	f.GetEnvironmentVariable()

	// IF PROVIDED FILE PASSED IN FROM CMD ARGS SET f.Filepath, OTHERWISE FIND IN env OR XDG
	fpSet, argSet := len(f.Filepath) > 0, len(path) > 0
	if fpSet || argSet { // FILEPATH SET ON struct INSTANTIATION
		f.Mode = ARG
	} else if !fpSet && argSet { // FILEPATH PASSED INTO FUNC
		f.Filepath = path[0]
	} else if !fpSet && !argSet { // FILEPATH NOT SET ON struct INSTANTIATION
		f.SearchForDatabase()
	}

	// GET ABSOLUTE PATH OF f.Filepath & VALIDATE
	var procs = []func(){
		f.getFullFilepath,
		f.ValidateFilepath,
	}
	for _, fn := range procs {
		fn()
		if f.ErrorMessage != nil {
			fmt.Println(f.ErrorMessage)
			os.Exit(1)
			break
		}
	}
	return
}

// Look in ENV VAR or XDG_DATA_DIR for database file
func (f *FileHandler) SearchForDatabase() {

	// PICK BETWEEN DB_FILEPATH_ENV_VAR & XDG_DATA_DIR
	var targetFilepath string
	if f.EnvVarSet {
		f.Mode = ENV
		if f.EnvVar == "" {
			f.ErrorMessage = fmt.Errorf(ErrEmptyEnvVar, DB_FILEPATH_ENV_VAR)
			return
		}
		targetFilepath = f.EnvVar
	} else {
		// NO NEED TO SET f.Mode; IT INITIALISES TO DESIRED VALUE FOR CASE
		targetFilepath = XDG_DATA_DIR
	}

	// SET FILEPATH - MOST VALIDATION HANDLED BY FileValidator.ValidateExistingFilepath()
	f.Filepath = targetFilepath
	return
}

func (f *FileHandler) ValidateDirectory() {
	// CHECK BASE DIRECTORY EXISTS AND IS WRITABLE
	_, err := os.Stat(path.Dir(f.Filepath))
	if err != nil {
		edgeErr, overrideErr := f.checkEdgeCase(err)
		if overrideErr {
			err = edgeErr
		}
		f.ErrorMessage = err
		return
	}
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

	tests := f.getValidationTests(f.Context)
	// LOOP THROUGH TESTS AND BREAK IF ONE SETS AN ERROR != nil
	for _, test := range tests {
		test(fileInfo)
		if f.ErrorMessage != nil {
			break
		}
	}
	return
}
