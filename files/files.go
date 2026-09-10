// GET OR CREATE DATABASE FILES
package files

import (
	"fmt"
	"os"
	"path"
)

// CONST
const (
	DB_DRIVER           string = "sqlite3"
	DB_FILEPATH_ENV_VAR string = "CHAR_INV_DATABASE"
	DB_FILE_EXTENSION   string = ".db"
	XDG_DATA_DIR        string = "$HOME/.local/share/char-inv/char-inv.db"

	// FILE ERRORS
	ErrAlreadyExists   string = "%s already exists"
	ErrEmptyEnvVar     string = "Set environment variable '%s' is empty"
	ErrFileNotExist    string = "stat %s: no such file or directory"
	ErrIrregularFile   string = "File '%s' is not a regular file"
	ErrNotDatabase     string = "File %s is not a database with file extentsion %s"
	ErrNonWriteable    string = "You do not have write permissions for file '%s'"
	ErrTargetDirectory string = "%s is a directory not a file"
	ErrXDGNotCreated   string = "Defaulted to XDG database (%s) but it hasn't been created yet.\nUse 'character-inventory new' to generate new database"
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
	EnvVar        string
	EnvVarSet     bool
	ErrorMessage  error
	Filepath      string
	FileExtension string
	Mode          HandlerMode
}

// VALIDATION FUNCTIONS
func (f *FileHandler) affirmDoesNotExist(fileInfo os.FileInfo) {
	if fileInfo != nil {
		f.ErrorMessage = fmt.Errorf(ErrAlreadyExists, f.Filepath)
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

// TODO: make proccess for file creation
func (f *FileHandler) CreateDatabaseFile() {
	f.FileExtension = DB_FILE_EXTENSION

	// VALIDATE GIVEN FILEPATH
	f.ValidateNewFilepath()
	if f.ErrorMessage != nil {
		fmt.Println(f.ErrorMessage)
		os.Exit(1)
	}

	// CREATE NEW FILE

	return
}

func (f *FileHandler) GetEnvironmentVariable() {
	f.EnvVar, f.EnvVarSet = os.LookupEnv(DB_FILEPATH_ENV_VAR)
	return
}

func (f *FileHandler) Init(path ...string) {
	f.FileExtension = DB_FILE_EXTENSION
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

	// VALIDATE PROVIDED FILEPATH BEFORE RETURN
	f.ValidateExistingFilepath()

	// CHECK FOR ERRORS SET TO FileHandler
	if f.ErrorMessage != nil {
		fmt.Println(f.ErrorMessage)
		os.Exit(1)
	}

	return
}

// LOOK IN ENV VAR OR XDG_DATA_DIR FOR DATABASE FILE
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

func (f *FileHandler) ValidateExistingFilepath() {
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

	tests := []func(os.FileInfo){
		f.affirmNotDirectory,
		f.affirmRegularFile,
		f.affirmWritePermission,
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

func (f *FileHandler) ValidateNewFilepath() {
	// databaseToCreate := path.Base(f.Filepath)
	databaseDirectory := path.Dir(f.Filepath)

	// CHECK BASE DIRECTORY EXISTS AND IS WRITABLE
	dirInfo, err := os.Stat(databaseDirectory)
	if err != nil {
		f.ErrorMessage = err
	}
	f.affirmWritePermission(dirInfo)
	if f.ErrorMessage != nil {
		return
	}

	// CHECK FILE DOES NOT EXIST AND IS NOT DIRECTORY
	fileInfo, err := os.Stat(f.Filepath)
	tests := []func(os.FileInfo){
		f.affirmNotDirectory,
		f.affirmDoesNotExist,
	}
	for _, test := range tests {
		test(fileInfo)
		if f.ErrorMessage != nil {
			break
		}
	}

	return
}
