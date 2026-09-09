// GET OR CREATE DATABASE FILES
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	DB_DRIVER           string = "sqlite3"
	DB_FILEPATH_ENV_VAR string = "CHAR_INV_DATABASE"
	DB_FILE_EXTENSION   string = ".db"
	XDG_DATA_DIR        string = "$HOME/.local/share/char-inv/char-inv.db"

	FILE_ERRORS = map[string]string{
		"EmptyEnvVar":		   "Set environment variable '%s' is empty",
		"FileNotExist":        "stat %s: no such file or directory",
		"FileNotFound":        "No database files with file extentsion '%s' found in directory '%s'",
		"IrregularFile":       "File '%s' is not a regular file",
		"NotDatabase":         "File %s is not a database with file extentsion %s",
		"NoFileInFilepath":	   "Target filepath '%s' is a directory",
		"NoWritePermsissions": "You do not have write permissions for file '%s'",
		"TargetIsDirectory":   "%s is a directory not a file",
	}
)

type FileHandler struct {
	Driver        string
	ErrorMessage  error
	Filepath      string
	FileExtension string
}

func (f *FileHandler) Init(path ...string) {
	f.FileExtension = DB_FILE_EXTENSION
	f.Driver = DB_DRIVER
	if len(path) > 0 {
		f.Filepath = path[0]
	}
	err := f.GetDatabaseFilepath()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return
}

func (f *FileHandler) CreateDatabaseFile() error {
	// directory := f.DatabaseFilepath
	// TODO: make proccess for file creation
	return nil
}

func (f *FileHandler) GetDatabaseFilepath() error {

	// IF f.Filepath NOT PROVIDED DURING f.Init() FIND FILEPATH
	if f.Filepath == "" {
		f.SearchForDatabase()
		if f.ErrorMessage != nil {
			return f.ErrorMessage
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

// LOOK IN ENV VAR OR XDG_DATA_DIR FOR DATABASE FILE
func (f *FileHandler) SearchForDatabase() {
	env, envSet := os.LookupEnv(DB_FILEPATH_ENV_VAR)
	if envSet && env == "" {
		f.ErrorMessage = fmt.Errorf(FILE_ERRORS["EmptyEnvVar"], DB_FILEPATH_ENV_VAR)
	}

	// PICK BETWEEN DB_FILEPATH_ENV_VAR & XDG_DATA_DIR
	var targetFilepath string
	if envSet {
		targetFilepath = env
	} else {
		targetFilepath = XDG_DATA_DIR
	}

	// CHECK FOR FILES
	_, file := filepath.Split(targetFilepath)
	if file == "" {
		f.ErrorMessage = fmt.Errorf(FILE_ERRORS["NoFileInFilepath"], targetFilepath)
	}

	f.Filepath = targetFilepath
	return
}

type FileValidator struct {
	ErrorMessage error
	filepath     string
}

func (v *FileValidator) affirmDatabase(fileInfo os.FileInfo) {
	if !strings.HasSuffix(fileInfo.Name(), DB_FILE_EXTENSION) {
		v.ErrorMessage = fmt.Errorf(FILE_ERRORS["NotDatabase"], v.filepath, DB_FILE_EXTENSION)
	}
}

func (v *FileValidator) affirmNotDirectory(fileInfo os.FileInfo) {
	if fileInfo.IsDir() {
		v.ErrorMessage = fmt.Errorf(FILE_ERRORS["TargetIsDirectory"], v.filepath)
	}
}

func (v *FileValidator) affirmRegularFile(fileInfo os.FileInfo) {
	if !fileInfo.Mode().IsRegular() {
		v.ErrorMessage = fmt.Errorf(FILE_ERRORS["IrregularFile"], v.filepath)
	}
	return
}

func (v *FileValidator) affirmWritePermission(fileInfo os.FileInfo) {
	if fileInfo.Mode().Perm() < 0o600 {
		// 0o600 IS OCTAL LITERAL EQUIVALENT TO USER WRITE PERMISSIONS
		v.ErrorMessage = fmt.Errorf(FILE_ERRORS["NoWritePermsissions"], v.filepath)
	}
	return
}

func (v *FileValidator) ValidateFilepath() {
	// err OUPUT OF os.Stat WILL HANDLE FILES THAT DO NOT EXIST
	fileInfo, err := os.Stat(v.filepath)
	if err != nil {
		v.ErrorMessage = err
		return
	}

	var tests = []func(os.FileInfo){
		v.affirmWritePermission,
		v.affirmRegularFile,
		v.affirmDatabase,
	}
	// LOOP THROUGH TESTS AND BREAK IF ONE SETS AN ERROR != nil
	for _, test := range tests {
		test(fileInfo)
		if v.ErrorMessage != nil {
			break
		}
	}
	return
}
