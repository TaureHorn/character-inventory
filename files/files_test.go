package files

import (
	"fmt"
	"os"
	"testing"
)

func TestValidator(t *testing.T) {
	tests := []struct {
		name          string
		filepath      string
		expectedError error
	}{
		{
			name:          "directory",
			filepath:      "test-data/test",
			expectedError: fmt.Errorf(FILE_ERRORS["TargetIsDirectory"], "test-data/test"),
		},
		{
			name:          "empty filepath",
			filepath:      "",
			expectedError: fmt.Errorf(FILE_ERRORS["FileNotExist"], ""),
		}, {
			name:          "filled filpath but file doesn't exist",
			filepath:      "test-data/testfile_xxx",
			expectedError: fmt.Errorf(FILE_ERRORS["FileNotExist"], "test-data/testfile_xxx"),
		}, {
			name:          "file exists, is NOT writeable",
			filepath:      "test-data/testfile_400",
			expectedError: fmt.Errorf(FILE_ERRORS["NoWritePermsissions"], "test-data/testfile_400"),
		}, {
			name:          "irregular file",
			filepath:      "/dev/nvme0n1",
			expectedError: fmt.Errorf(FILE_ERRORS["IrregularFile"], "/dev/nvme0n1"),
		}, {
			name:          "file exists, is writeable, NOT database",
			filepath:      "test-data/testfile_644",
			expectedError: fmt.Errorf(FILE_ERRORS["NotDatabase"], "test-data/testfile_644", DB_FILE_EXTENSION),
		}, {
			name:          "file exists, is writeable, is database",
			filepath:      "test-data/database.db",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(XDG_DATA_DIR)
			v := new(FileHandler{Filepath: tt.filepath})
			v.FileExtension = DB_FILE_EXTENSION
			v.ValidateFilepath()

			// CONVERT TO STRINGS TO EASILY USE nil OR fs.PathError IN COMPARISON
			vErr, tErr := fmt.Sprint(v.ErrorMessage), fmt.Sprint(tt.expectedError)
			if vErr != tErr {
				t.Errorf("%s | Expected '%s', but got '%s'\n", t.Name(), tt.expectedError, v.ErrorMessage)
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
		expectedMode  uint
	}{
		{
			name:          "xdg",
			setEnv:        false,
			envVar:        "",
			expectedError: nil,
			expectedMode:  HANDLER_SOURCE_MODE["XDG"],
		}, {
			name:          "env var empty",
			setEnv:        true,
			envVar:        "",
			expectedError: fmt.Errorf(FILE_ERRORS["EmptyEnvVar"], DB_FILEPATH_ENV_VAR),
			expectedMode:  HANDLER_SOURCE_MODE["ENV"],
		}, {
			name:          "env var",
			setEnv:        true,
			envVar:        "files/test-data/database.db",
			expectedError: nil,
			expectedMode:  HANDLER_SOURCE_MODE["ENV"],
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
