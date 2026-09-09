package files

import (
	"fmt"
	"testing"
)

func TestValidator(t *testing.T) {
	var tests = []struct {
		name          string
		filepath      string
		expectedError error
	}{
		{
			name:          "directory",
			filepath:      "test-data/test",
			expectedError: fmt.Errorf(FILE_ERRORS["IrregularFile"], "test-data/test"),
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
			name:          "file does exist but no write permissions",
			filepath:      "test-data/testfile_400",
			expectedError: fmt.Errorf(FILE_ERRORS["NoWritePermsissions"], "test-data/testfile_400"),
		}, {
			name:          "irregular file",
			filepath:      "/dev/nvme0n1",
			expectedError: fmt.Errorf(FILE_ERRORS["IrregularFile"], "/dev/nvme0n1"),
		}, {
			name:          "file does exist & has write permissions not database",
			filepath:      "test-data/testfile_644",
			expectedError: fmt.Errorf(FILE_ERRORS["NotDatabase"], "test-data/testfile_644", DB_FILE_EXTENSION),
		}, {
			name:          "file exists, is writeable and is database",
			filepath:      "test-data/database.db",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := new(FileValidator{filepath: tt.filepath})
			v.ValidateFilepath()

			// CONVERT TO STRINGS TO EASILY USE nil OR fs.PathError IN COMPARISON
			vErr, tErr := fmt.Sprint(v.ErrorMessage), fmt.Sprint(tt.expectedError)
			if vErr != tErr {
				t.Errorf("%s | Expected '%s', but got '%s'\n", t.Name(), tt.expectedError, v.ErrorMessage)
			}
		})
	}
}
