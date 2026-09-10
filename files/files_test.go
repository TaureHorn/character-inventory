package files

import (
	"fmt"
	"os"
	"testing"
)

func TestValidate(t *testing.T) {

	os.Unsetenv(DB_FILEPATH_ENV_VAR)
	v := new(FileHandler)
	v.FileExtension = DB_FILE_EXTENSION

	tests := []struct {
		name          string
		filepath      string
		expectedError error
		function      func()
	}{
		{
			name:          "directory",
			filepath:      "test-data/test",
			expectedError: fmt.Errorf(ErrTargetDirectory, "test-data/test"),
			function:      v.ValidateExistingFilepath,
		}, {
			name:          "empty filepath",
			filepath:      "",
			expectedError: fmt.Errorf(ErrFileNotExist, ""),
			function:      v.ValidateExistingFilepath,
		}, {
			name:          "filled filpath but file doesn't exist",
			filepath:      "test-data/testfile_xxx",
			expectedError: fmt.Errorf(ErrFileNotExist, "test-data/testfile_xxx"),
			function:      v.ValidateExistingFilepath,
		}, {
			name:          "file exists, is NOT writeable",
			filepath:      "test-data/testfile_400",
			expectedError: fmt.Errorf(ErrNonWriteable, "test-data/testfile_400"),
			function:      v.ValidateExistingFilepath,
		}, {
			name:          "irregular file",
			filepath:      "/dev/nvme0n1",
			expectedError: fmt.Errorf(ErrIrregularFile, "/dev/nvme0n1"),
			function:      v.ValidateExistingFilepath,
		}, {
			name:          "file exists, is writeable",
			filepath:      "test-data/testfile_644",
			expectedError: nil,
			function:      v.ValidateExistingFilepath,
		}, {
			name:          "file already exists",
			filepath:      "test-data/database.db",
			expectedError: fmt.Errorf(ErrAlreadyExists, "test-data/database.db"),
			function:      v.ValidateNewFilepath,
		}, {
			name:          "directory not writeable",
			filepath:      "test-data/no-write-dir/file",
			expectedError: fmt.Errorf(ErrNonWriteable, "test-data/no-write-dir/file"),
			function:      v.ValidateNewFilepath,
		}, {
			name:          "file is directory",
			filepath:      "test-data/test",
			expectedError: fmt.Errorf(ErrTargetDirectory, "test-data/test"),
			function:      v.ValidateNewFilepath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v.ErrorMessage = nil
			v.Filepath = tt.filepath
			tt.function()

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
