package files

import (
	"testing"
)

func TestValidator(t *testing.T) {
	var tests = []struct {
		name          string
		filepath      string
		expectedError error
	}{
		{
			name:          "empty filepath",
			filepath:      "",
			expectedError: FILE_ERRORS["EmptyFileString"],
		}, {
			name:          "filled filpath but file doesn't exist",
			filepath:      "test-data/testfile_xxx",
			expectedError: FILE_ERRORS["DoesNotExist"],
		}, {
			name:          "file does exist but no write permissions",
			filepath:      "test-data/testfile_400",
			expectedError: FILE_ERRORS["NoWritePermsissions"],
		}, {
			name:          "file does exist & has write permissions",
			filepath:      "test-data/testfile_644",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := new(FileValidator{filepath: tt.filepath})
			v.ValidateFilepath()
			if v.ErrorMessage != tt.expectedError {
				t.Errorf("%s | Expected %s, but got %s", t.Name(), tt.expectedError, v.ErrorMessage)
			}
		})
	}
}
