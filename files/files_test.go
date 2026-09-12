package files

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"
)

const (
	MEMORY 		string = ":memory:"
	SQL_VERSION string = "3.53.4"
)

func TestDatabase(t *testing.T) {
	d := new(DataHandler)
	d.Init(DB_DRIVER, MEMORY, INIT)

	var version string
	err := d.Database.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)
	if err != nil {
		t.Error(err)
	} else if version != SQL_VERSION {
		t.Errorf("Expected '%s', got '%s'\n", SQL_VERSION, version)
	}
	
}

func TestNewFileCreation(t *testing.T) {
	os.Unsetenv(DB_FILEPATH_ENV_VAR)
	defaultData, readErr := os.ReadFile("default.db")
	if readErr != nil {
		t.Error(readErr)
	}

	tests := []struct {
		filepath string
		mode     HandlerMode
	}{
		{mode: ARG, filepath: "$HOME/Desktop/argTest.db"},
		{mode: ENV, filepath: "$HOME/Desktop/envVar.db"},
		{mode: XDG, filepath: XDG_DATA_DIR},
	}
	for _, tt := range tests {
		_ = os.Remove(tt.filepath)
		t.Run(tt.mode.String(), func(t *testing.T) {
			f := new(FileHandler)
			switch tt.mode {
			case ARG:
				f.Mode = tt.mode
				f.Filepath = tt.filepath
			case ENV:
				f.Mode = tt.mode
				os.Setenv(DB_FILEPATH_ENV_VAR, tt.filepath)
				f.GetEnvironmentVariable()
				f.Filepath = f.EnvVar
			case XDG:
				f.Mode = tt.mode
				f.Filepath = tt.filepath
			}
			f.CreateDatabaseFile()

			newFile, newFileReadErr := os.ReadFile(f.Filepath)
			if newFileReadErr != nil {
				t.Error(tt.mode, newFileReadErr)
			}

			if !reflect.DeepEqual(defaultData, newFile) {
				t.Error("new file does not match default data")
			}

			deleteErr := os.Remove(f.Filepath)
			if deleteErr != nil {
				t.Error(tt.mode, deleteErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {

	os.Unsetenv(DB_FILEPATH_ENV_VAR)

	tests := []struct {
		name          string
		filepath      string
		expectedError string
		context       ValidationContext
	}{
		// INITIATE NEW FILE HANDER TESTS
		{
			name:          "directory",
			filepath:      "test-data/test",
			expectedError: ErrTargetDirectory,
			context:       INIT,
		}, {
			name:          "empty filepath",
			filepath:      "",
			expectedError: ErrTargetDirectory,
			context:       INIT,
		}, {
			name:          "filled filpath but file doesn't exist",
			filepath:      "test-data/testfile_xxx",
			expectedError: ErrFileNotExist,
			context:       INIT,
		}, {
			name:          "file exists, is NOT writeable",
			filepath:      "test-data/testfile_400",
			expectedError: ErrPermissionDenied,
			context:       INIT,
		}, {
			name:          "irregular file",
			filepath:      "/dev/nvme0n1",
			expectedError: ErrIrregularFile,
			context:       INIT,
		}, {
			name:          "file exists, is writeable",
			filepath:      "test-data/testfile_644",
			expectedError: "",
			context:       INIT,
		},
		// CREATE NEW DATABASE TESTS
		{
			name:          "file doesnt exist",
			filepath:      "test-data/newDatabase",
			expectedError: "",
			context:       CREATE,
		}, {
			name:          "file doesnt exist (CWD)",
			filepath:      "newDatabase",
			expectedError: "",
			context:       CREATE,
		}, {
			name:          "path contains $VAR",
			filepath:      "$HOME/database.db",
			expectedError: "",
			context:       CREATE,
		}, {
			name:          "file already exists",
			filepath:      "test-data/database.db",
			expectedError: ErrAlreadyExists,
			context:       CREATE,
		}, {
			name:          "directory doesn't exist",
			filepath:      "fake-directory/database",
			expectedError: ErrFileNotExist,
			context:       CREATE,
		}, {
			name:          "directory not writeable",
			filepath:      "test-data/no-write-dir/file",
			expectedError: ErrPermissionDenied,
			context:       CREATE,
		}, {
			name:          "file is directory",
			filepath:      "test-data/test",
			expectedError: ErrTargetDirectory,
			context:       CREATE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FILEHANDLER SETUP
			v := new(FileHandler{Context: tt.context, Filepath: tt.filepath})
			v.getFullFilepath()
			wantedErr := func() error {
				if tt.expectedError == "" {
					return nil
				} else {
					filepath := v.Filepath
					if tt.name == "directory doesn't exist" {
						filepath = path.Dir(v.Filepath)
					}

					return fmt.Errorf(tt.expectedError, filepath)
				}
			}()

			// RUN
			switch tt.context {
			case CREATE:
				v.CreateDatabaseFile()
				if wantedErr == nil {
					os.Remove(v.Filepath)
				}
			case INIT:
				v.Init()
			}

			// CONVERT TO STRINGS TO EASILY USE nil OR fs.PathError IN COMPARISON
			vErr, tErr := fmt.Sprint(v.ErrorMessage), fmt.Sprint(wantedErr)
			if vErr != tErr {
				t.Errorf("%s | Expected '%s', but got '%s'\n\n", t.Name(), wantedErr, v.ErrorMessage)
			}
		})
	}
}

func TestSearchForDatabase(t *testing.T) {
	tests := []struct {
		name          string
		setEnv        bool
		envVar        string
		expectedError error
		expectedMode  HandlerMode
	}{
		{
			name:          "xdg",
			setEnv:        false,
			envVar:        "",
			expectedError: nil,
			expectedMode:  XDG,
		}, {
			name:          "env var empty",
			setEnv:        true,
			envVar:        "",
			expectedError: fmt.Errorf(ErrEmptyEnvVar, DB_FILEPATH_ENV_VAR),
			expectedMode:  ENV,
		}, {
			name:          "env var",
			setEnv:        true,
			envVar:        "files/test-data/database.db",
			expectedError: nil,
			expectedMode:  ENV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(DB_FILEPATH_ENV_VAR, tt.envVar)
			} else {
				os.Unsetenv(DB_FILEPATH_ENV_VAR)
			}

			env, envSet := os.LookupEnv(DB_FILEPATH_ENV_VAR)
			f := new(FileHandler)
			f.GetEnvironmentVariable()
			f.SearchForDatabase()

			// CONVERT TO STRINGS TO EASILY USE nil OR fs.PathError IN COMPARISON
			fErr, tErr := fmt.Sprint(f.ErrorMessage), fmt.Sprint(tt.expectedError)
			if fErr != tErr {
				t.Errorf("%s | Expected '%s', but got '%s'\n", t.Name(), tt.expectedError, f.ErrorMessage)
				if envSet {
					t.Errorf("%s | %s: '%s'\n\n", t.Name(), DB_FILEPATH_ENV_VAR, env)
				}
			}
			if f.Mode != tt.expectedMode {
				t.Errorf("%s | Expected mode %d, but got %d \n", t.Name(), tt.expectedMode, f.Mode)
			}
		})
	}
}


