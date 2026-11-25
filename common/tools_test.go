//go:build unit_test

/*
Package ginterfaces
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Test runners.
*/
package common

import (
	"encoding/base64"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func init() {
	SetupTest()
}

func TestSimpleReplacer(t *testing.T) {

	data1 := map[string]interface{}{
		"APPLE": "ORANGE",
		"Green": "Blue",
		"North": "South",
	}

	text1 := `jsjsjs jsjsjws <<< APPLE >>> WEF,M SDJDFJDF < JASJDASJDASJ <<<Green>>> asdfdsfsdf asd <<< NORTH >>> asfa>sdasd`
	s, err := SimpleReplacer(text1, data1, false)
	if err != nil {
		assert.NoError(t, err, "Bad replacement.")
		return
	}
	sp, _ := testutil.NewScanPack(s, nil)
	assert.GreaterOrEqual(t, sp.Seek("ORANGE"), 0, "APPLE")
	assert.GreaterOrEqual(t, sp.Seek("DF < JA"), 0, "Broken left as is.")
	assert.GreaterOrEqual(t, sp.Seek("Blue"), 0, "Green")
	assert.Less(t, sp.Seek("NORTH"), 0, "NORTH nuked")

	text2 := `jsjsjs jsjsjws <<< APPLE >>> WE<<<Green>>> asdfdsfsdf asd <<< asdadsdd`
	s, err = SimpleReplacer(text2, data1, false)
	assert.Error(t, err, "Dangling check.")
}

// TestSimpleReplacer2 - AI is fun!  It is very convenient, but the generated tests had two bugs.
func TestSimpleReplacer2(t *testing.T) {
	var tests = []struct {
		name      string
		input     string
		mapInput  map[string]interface{}
		wantErr   bool
		expectRes string
	}{
		{
			name:     "empty input",
			input:    "",
			mapInput: map[string]interface{}{},
			// BUG 1: Nil input is not a failure.
			wantErr:   false,
			expectRes: "",
		},
		{
			name:      "no replacement needed",
			input:     func() string { s := "hello world"; return s }(),
			mapInput:  map[string]interface{}{},
			wantErr:   false,
			expectRes: "hello world",
		},
		{
			name:     "replacement not found in map",
			input:    func() string { s := "hello <<<name>>>"; return s }(),
			mapInput: map[string]interface{}{"surname": "go"},
			// BUG 2: Not matching is not a failure.
			wantErr:   false,
			expectRes: "hello ",
		},
		{
			name:     "dangling replacer",
			input:    func() string { s := "hello <<<name"; return s }(),
			mapInput: map[string]interface{}{"name": "go"},
			wantErr:  true,
		},
		{
			name:      "successful replacement",
			input:     func() string { s := "hello <<<name>>>"; return s }(),
			mapInput:  map[string]interface{}{"name": "go"},
			wantErr:   false,
			expectRes: "hello go",
		},
		{
			name:      "bug #1",
			input:     func() string { s := "http://<<<host>>>:<<<port>>>/xxx"; return s }(),
			mapInput:  map[string]interface{}{"host": "localhost", "port": "8080"},
			wantErr:   false,
			expectRes: "http://localhost:8080/xxx",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resStr, err := SimpleReplacer(tc.input, tc.mapInput, false)
			if (err != nil) != tc.wantErr {
				t.Errorf("SimpleReplacer() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if err == nil && resStr != "" && resStr != tc.expectRes {
				t.Errorf("SimpleReplacer() = %v, want %v", resStr, tc.expectRes)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name       string
		beforeCopy func() (string, string, error)       // Function to prepare test data
		afterCopy  func(srcPath, destPath string) error // Function to validate test results and clear the test data
		wantErr    bool
	}{
		{
			name: "CopyFileNormal",
			beforeCopy: func() (string, string, error) {
				// Create a temporary file
				srcFile, err := os.CreateTemp("", "src")
				if err != nil {
					return "", "", err
				}
				defer func(srcFile *os.File) {
					_ = srcFile.Close()
				}(srcFile)

				// Write some data to the file
				_, _ = srcFile.WriteString("This is a test data")

				// Create a destination file
				destFile, err := os.CreateTemp("", "dest")
				if err != nil {
					return "", "", err
				}

				return srcFile.Name(), destFile.Name(), nil
			},
			afterCopy: func(srcPath, destPath string) error {
				defer func(name string) {
					_ = os.Remove(name)
				}(srcPath)
				defer func(name string) {
					_ = os.Remove(name)
				}(destPath)

				// Verify the file copy
				gotData, err := os.ReadFile(destPath)
				if err != nil {
					return err
				}

				wantData, err := os.ReadFile(srcPath)
				if err != nil {
					return err
				}

				if diff := cmp.Diff(wantData, gotData); diff != "" {
					t.Errorf("Mismatch in copied data (-want +got):\n%s", diff)
				}

				return nil
			},
			wantErr: false,
		},
		{
			name: "CopyFileNonExistant",
			beforeCopy: func() (string, string, error) {
				return "NonExistingFile", "NonExistingFile", nil
			},
			afterCopy: func(_ string, _ string) error {
				// Nothing to clean here, as the files don't exist
				return nil
			},
			wantErr: true,
		},
		// ADD MORE TEST CASES HERE TO COVER OTHER CORNER CASES
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			srcPath, destPath, err := tc.beforeCopy()
			if err != nil {
				t.Fatalf("failed to prepare test data: %v", err)
			}

			gotErr := CopyFile(srcPath, destPath)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("CopyFile() error = %v, wantErr %v", gotErr, tc.wantErr)
				return
			}

			if err := tc.afterCopy(srcPath, destPath); err != nil {
				t.Errorf("failed to validate test results: %v", err)
			}
		})
	}
}

func TestBase64ToJsonMap(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "empty input",
			data:    "",
			wantErr: true,
		},
		{
			name:    "invalid base64",
			data:    "+++",
			wantErr: true,
		},
		{
			name:    "valid base64, invalid json",
			data:    base64.StdEncoding.EncodeToString([]byte(`{"foo": "bar",}`)),
			wantErr: true,
		},
		{
			name:    "valid base64, valid json",
			data:    base64.StdEncoding.EncodeToString([]byte(`{"foo": "bar"}`)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base64ToJsonMap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Base64ToJsonMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got["foo"] != "bar" {
				t.Errorf("Base64ToJsonMap() = %v, want %v", got, map[string]interface{}{"foo": "bar"})
			}
		})
	}
}

func TestStringToJsonMap(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name:    "Empty String",
			data:    "",
			want:    nil,
			wantErr: true,
		},
		{
			name: "Valid JSON",
			data: `{"key":"value", "key2": 2}`,
			want: map[string]interface{}{
				"key":  "value",
				"key2": float64(2), // json unmarshaling treats numbers as float64 by default
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			data:    `{"key": value}`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringToJsonmap(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToJsonmap() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToJsonmap() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

type TestToolJsonFileCasesData struct {
	Name string `json:"Name"`
}

func TestToolJsonFileCases(t *testing.T) {
	t.Run("ReadJsonFileMap file doesn't exist", func(t *testing.T) {
		var target map[string]interface{}
		err := ReadJsonFileMap("sdfsdfsdiowejnc", "xxx", &target)
		assert.Error(t, err)
	})
	t.Run("ReadJsonFileMap bad file", func(t *testing.T) {
		var target map[string]interface{}
		err := ReadJsonFileMap(filepath.Join(base.GetGinfraHome(), "README.md"), "xxx", &target)
		assert.Error(t, err)
	})
	t.Run("WriteJsonFileA bad target file path", func(t *testing.T) {
		var target = map[string]interface{}{"hi": "low"}
		err := WriteJsonFileA(target, "../../../../../../../../../../../../../", "xxx")
		assert.Error(t, err)
	})
	t.Run("ReadJsonFile bad file", func(t *testing.T) {
		_, err := ReadJsonFile[int](filepath.Join(base.GetGinfraHome(), "README.md"), "xxx")
		assert.Error(t, err)
	})
	t.Run("WriteJsonFile bad file", func(t *testing.T) {
		d := TestToolJsonFileCasesData{Name: "Name"}
		err := WriteJsonFile[TestToolJsonFileCasesData](d, "../../../../../../../../../../../../../", "xxx")
		assert.Error(t, err)
	})
}

func TestToolCases(t *testing.T) {
	t.Run("GetInterfaceName", func(t *testing.T) {
		assert.Contains(t, GetInterfaceName("GetInterfaceName"), "GetInterfaceName")
	})

	t.Run("Bad file destination", func(t *testing.T) {
		err := CopyFile(filepath.Join(base.GetGinfraHome(), "README.md"), "/")
		assert.Error(t, err)
	})

	// This really sucks.  This test passes on my dev laptop and when run as a single package.  But when
	// I try to run all the tests on my main server, it fails.  The copy returns nil instead of an error.
	// If I attach a debugger to it remotely, it even works then.  WTF
	// UPDATE: it is flaky everywhere now.
	/*
		t.Run("Bad copyfiles destination path error", func(t *testing.T) {
			if testutil.SkipFlakyCheck() {
				return
			}
			werr := CopyFiles(base.GetGinfraTmpTestDir(), "../../sdfasdfasdf/../.../zdfasdfa/.../sdfasdf/../../../..", false, false,
				false)
			assert.ErrorContains(t, werr, "no such file")
		})
	*/

	t.Run("Bad dir destination exists as file", func(t *testing.T) {
		err := MakeDir(filepath.Join(base.GetGinfraHome(), "README.md"))
		assert.Error(t, err)
	})

	t.Run("Bad dir path", func(t *testing.T) {
		err := MakeDir(filepath.Join("\000"))
		assert.Error(t, err)
	})

	t.Run("Copy files - Bad dir path", func(t *testing.T) {
		err := CopyFiles("sdfasdf", "\000", false, false, false)
		assert.Error(t, err)
	})

	t.Run("ReadAllString good", func(t *testing.T) {
		s, e := ReadAllString(filepath.Join(base.GetGinfraHome(), "README.md"))
		assert.NoError(t, e)
		assert.Less(t, 1, len(s))
	})

	t.Run("ReadAllString bad", func(t *testing.T) {
		_, e := ReadAllString(filepath.Join(base.GetGinfraHome(), "kwhjer09283h2odhdwqohdoewgf892gfolB"))
		assert.Error(t, e)
	})

	t.Run("Load array", func(t *testing.T) {
		d, err := ReadFileIntoArray(filepath.Join(base.GetGinfraHome(), "LICENSE"))
		assert.NoError(t, err)
		if len(d) < 2 {
			t.Errorf("Expected at least 2 lines in the file, got %d", len(d))
		}
		assert.Contains(t, d[1], "Apache")
	})

	t.Run("Load array - bad path", func(t *testing.T) {
		_, err := ReadFileIntoArray(filepath.Join(base.GetGinfraHome(), "nerigiqbevb ajxvkasdvandksv"))
		assert.Error(t, err)
	})

	t.Run("Clean URL - no parts", func(t *testing.T) {
		r := CleanUrl("thing/go.html")
		assert.Equal(t, "thing/go.html", r)
	})

}

func TestSimpleReplacerSnarfer(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
		expectError   bool
		expectedState int
	}{
		{
			name:          "Basic input without markers",
			input:         "hello world",
			expected:      "hello world",
			expectError:   false,
			expectedState: StateReplacerOpen,
		},
		{
			name:          "Closed replacement marker",
			input:         "hello>>>",
			expected:      "hello",
			expectError:   false,
			expectedState: StateReplacerExit,
		},
		{
			name:          "Nested replacement markers",
			input:         "hello<<<def>>>",
			expected:      "hello<<<def",
			expectError:   false,
			expectedState: StateReplacerExit,
		},
		{
			name:          "Empty input",
			input:         "",
			expected:      "",
			expectError:   false,
			expectedState: StateReplacerOpen,
		},
		{
			name:          "Replacement marker as entire input",
			input:         "<<<>>>",
			expected:      "<<<",
			expectError:   false,
			expectedState: StateReplacerExit,
		},
		{
			name:          "Input ends abruptly with replacement marker",
			input:         "abc<<<",
			expected:      "abc<<<",
			expectError:   false,
			expectedState: StateReplacerOpen,
		},
		{
			name:          "Invalid sequence with EOF",
			input:         "abc<<<hello",
			expected:      "abc<<<hello",
			expectError:   false,
			expectedState: StateReplacerOpen,
		},
		{
			name:          "Broken exit",
			input:         "hello<<<def>>a",
			expected:      "hello<<<def>>a",
			expectError:   false,
			expectedState: StateReplacerOpen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			result, err, state := simpleReplacerSnarfer(reader)
			if (err != nil) != tt.expectError {
				t.Errorf("simpleReplacerSnarfer() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if result != tt.expected {
				t.Errorf("simpleReplacerSnarfer() = %v, expected %v", result, tt.expected)
			}
			if state != tt.expectedState {
				t.Errorf("simpleReplacerSnarfer() state = %v, expectedState %v", state, tt.expectedState)
			}
		})
	}
}
