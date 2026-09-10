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
	ErrAlreadyExists    string = "%s already exists"
	ErrEmptyEnvVar      string = "Set environment variable '%s' is empty"
	ErrFileNotExist     string = "stat %s: no such file or directory"
	ErrIrregularFile    string = "File '%s' is not a regular file"
	ErrNotDatabase      string = "File %s is not a database with file extentsion %s"
	ErrNonWriteable     string = "You do not have write permissions for file '%s'"
	ErrPermissionDenied string = "stat %s: permission denied"
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
	Context       ValidationContext
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
	f.Context = CREATE
	f.FileExtension = DB_FILE_EXTENSION

	// VALIDATE GIVEN FILEPATH
	for _, validation := range []func(){f.ValidateDirectory, f.ValidateFilepath} {
		validation()
		if f.ErrorMessage != nil {
			fmt.Println(f.ErrorMessage)
			os.Exit(1)
			break
		}
	}

	// CREATE NEW FILE

	return
}

func (f *FileHandler) GetEnvironmentVariable() {
	f.EnvVar, f.EnvVarSet = os.LookupEnv(DB_FILEPATH_ENV_VAR)
	return
}

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

func (f *FileHandler) Init(path ...string) {
	f.Context = INIT
	f.Driver = DB_DRIVER
	f.FileExtension = DB_FILE_EXTENSION
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
	f.ValidateFilepath()

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

func (f *FileHandler) ValidateDirectory() {

	// CHECK BASE DIRECTORY EXISTS AND IS WRITABLE
	dirInfo, err := os.Stat(path.Dir(f.Filepath))
	if err != nil {
		f.ErrorMessage = err
		return
	}
	f.affirmWritePermission(dirInfo)
	if f.ErrorMessage != nil {
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
