package shared

import (
	"archive/zip"
	"bytes"
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"path/filepath"
	"testing"
)

func TestBundleDetect(t *testing.T) {
	t.Parallel()

	// Create temporary test environment
	tempDir := base.GetGinfraRandomTmpTestDir()
	rootPath := filepath.Join(tempDir, "root")
	err := os.MkdirAll(rootPath, 0755)
	if err != nil {
		t.Fatalf("failed to create root directory: %v", err)
	}
	tempBundlePath := filepath.Join(rootPath, BundleFileName)
	err = os.WriteFile(tempBundlePath, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test bundle file: %v", err)
	}

	tests := []struct {
		name     string
		setup    func() (string, func())
		expected string
	}{
		{
			name: "file in specified root folder",
			setup: func() (string, func()) {
				// Clean-up automatically handled by the test temp folder
				return rootPath, nil
			},
			expected: filepath.Join(rootPath, BundleFileName),
		},
		{
			name: "file in current directory",
			setup: func() (string, func()) {
				currentDir, err := os.Getwd()
				if err != nil {
					t.Fatalf("failed to get current working directory: %v", err)
				}

				testPath := filepath.Join(currentDir, BundleFileName)
				err = os.WriteFile(testPath, []byte("test content"), 0644)
				if err != nil {
					t.Fatalf("failed to create test file in current directory: %v", err)
				}

				return "", func() {
					_ = os.Remove(testPath)
				}
			},
			expected: filepath.Join(".", BundleFileName),
		},
		{
			name: "file not found",
			setup: func() (string, func()) {
				err = os.Remove(tempBundlePath)
				if err != nil {
					panic(err)
				}

				return rootPath, nil
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, cleanup := test.setup()
			if cleanup != nil {
				defer cleanup()
			}

			result := BundleDetect(root)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestBundleExtractFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() (string, string, []string, func())
		expectErr bool
	}{
		{
			name: "successful file extraction",
			setup: func() (string, string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()

				// Create a bundle with a test file
				bundlePath := filepath.Join(tempDir, BundleFileName)
				err := os.MkdirAll(filepath.Dir(bundlePath), 0755)
				if err != nil {
					panic(err)
				}

				fileContent := []byte("test content")
				fileName := "test.txt"

				buf := new(bytes.Buffer)
				writer := zip.NewWriter(buf)
				f, err := writer.Create(fileName)
				if err != nil {
					t.Fatalf("failed to create file in zip: %v", err)
				}
				_, err = f.Write(fileContent)
				if err != nil {
					t.Fatalf("failed to write content to file in zip: %v", err)
				}
				_ = writer.Close()

				err = os.WriteFile(bundlePath, buf.Bytes(), 0644)
				if err != nil {
					t.Fatalf("failed to write the zip file: %v", err)
				}

				destDir := filepath.Join(tempDir, "output")
				err = os.Mkdir(destDir, 0755)
				if err != nil {
					t.Fatalf("failed to create destination directory: %v", err)
				}

				return bundlePath, destDir, []string{fileName}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expectErr: false,
		},
		{
			name: "empty items list",
			setup: func() (string, string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()
				bundlePath := filepath.Join(tempDir, BundleFileName)
				_ = os.MkdirAll(filepath.Dir(bundlePath), 0755)

				err := os.WriteFile(bundlePath, []byte{}, 0644)
				if err != nil {
					t.Fatalf("failed to write the empty zip file: %v", err)
				}

				destDir := filepath.Join(tempDir, "output")
				err = os.Mkdir(destDir, 0755)
				if err != nil {
					t.Fatalf("failed to create destination directory: %v", err)
				}

				return bundlePath, destDir, []string{}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expectErr: false,
		},
		{
			name: "file not found in bundle",
			setup: func() (string, string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()

				// Create a zip bundle without the target file
				bundlePath := filepath.Join(tempDir, BundleFileName)
				_ = os.MkdirAll(filepath.Dir(bundlePath), 0755)

				writer := zip.NewWriter(bytes.NewBuffer(nil))
				_ = writer.Close()
				err := os.WriteFile(bundlePath, []byte{}, 0644)
				if err != nil {
					t.Fatalf("failed to write empty zip file: %v", err)
				}

				destDir := filepath.Join(tempDir, "output")
				err = os.Mkdir(destDir, 0755)
				if err != nil {
					t.Fatalf("failed to create destination directory: %v", err)
				}

				return bundlePath, destDir, []string{"nonexistent.txt"}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expectErr: true,
		},
		{
			name: "invalid bundle path",
			setup: func() (string, string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()
				destDir := filepath.Join(tempDir, "output")
				err := os.MkdirAll(destDir, 0755)
				if err != nil {
					t.Fatalf("failed to create destination directory: %v", err)
				}

				return "invalid_path.zip", destDir, []string{"file.txt"}, func() {
					_ = os.RemoveAll(tempDir)
				}
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundlePath, destDir, items, cleanup := test.setup()
			defer cleanup()

			err := BundleExtractFiles(bundlePath, destDir, items)
			if test.expectErr && err == nil {
				t.Errorf("expected an error but got none")
			} else if !test.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBundleExtract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() (string, []string, func())
		expected  [][]byte
		expectErr bool
	}{
		{
			name: "valid items in bundle",
			setup: func() (string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()

				// Create a zip file with two files
				bundlePath := filepath.Join(tempDir, BundleFileName)
				_ = os.MkdirAll(filepath.Dir(bundlePath), 0755)
				buf := new(bytes.Buffer)
				writer := zip.NewWriter(buf)

				files := map[string][]byte{
					"file1.txt": []byte("content1"),
					"file2.txt": []byte("content2"),
				}

				for name, content := range files {
					f, err := writer.Create(name)
					if err != nil {
						t.Fatalf("failed to create file in zip: %v", err)
					}
					_, err = f.Write(content)
					if err != nil {
						t.Fatalf("failed to write content to zip: %v", err)
					}
				}

				_ = writer.Close()
				_ = os.WriteFile(bundlePath, buf.Bytes(), 0644)

				return bundlePath, []string{"file1.txt", "file2.txt"}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expected:  [][]byte{[]byte("content1"), []byte("content2")},
			expectErr: false,
		},
		{
			name: "empty items list",
			setup: func() (string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()

				// Create an empty zip file
				bundlePath := filepath.Join(tempDir, BundleFileName)
				_ = os.MkdirAll(filepath.Dir(bundlePath), 0755)
				buf := new(bytes.Buffer)
				writer := zip.NewWriter(buf)
				_ = writer.Close()
				_ = os.WriteFile(bundlePath, buf.Bytes(), 0644)

				return bundlePath, []string{}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expected:  [][]byte{},
			expectErr: false,
		},
		{
			name: "item not found in bundle",
			setup: func() (string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()

				// Create a zip file without the required file
				bundlePath := filepath.Join(tempDir, BundleFileName)
				_ = os.MkdirAll(filepath.Dir(bundlePath), 0755)
				buf := new(bytes.Buffer)
				writer := zip.NewWriter(buf)
				_ = writer.Close()
				_ = os.WriteFile(bundlePath, buf.Bytes(), 0644)

				return bundlePath, []string{"nonexistent.txt"}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expected:  [][]byte{nil}, // File not found, so data is nil
			expectErr: true,
		},
		{
			name: "invalid bundle path",
			setup: func() (string, []string, func()) {
				return "invalid_path.zip", []string{"file.txt"}, func() {}
			},
			expected:  [][]byte{nil},
			expectErr: true,
		},
		{
			name: "partial match in bundle",
			setup: func() (string, []string, func()) {
				tempDir := base.GetGinfraRandomTmpTestDir()

				// Create a zip file with one matching and one missing file
				bundlePath := filepath.Join(tempDir, BundleFileName)
				_ = os.MkdirAll(filepath.Dir(bundlePath), 0755)
				buf := new(bytes.Buffer)
				writer := zip.NewWriter(buf)

				f, err := writer.Create("file1.txt")
				if err != nil {
					t.Fatalf("failed to create file in zip: %v", err)
				}
				_, err = f.Write([]byte("content1"))
				if err != nil {
					t.Fatalf("failed to write content to zip: %v", err)
				}

				_ = writer.Close()
				_ = os.WriteFile(bundlePath, buf.Bytes(), 0644)

				return bundlePath, []string{"file1.txt", "file2.txt"}, func() {
					os.RemoveAll(tempDir)
				}
			},
			expected:  [][]byte{[]byte("content1"), nil},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundlePath, items, cleanup := test.setup()
			defer cleanup()

			result, err := BundleExtract(bundlePath, items)

			if test.expectErr && err == nil {
				t.Errorf("expected an error but got none")
			} else if !test.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			for i, expectedData := range test.expected {
				if !bytes.Equal(result[i], expectedData) {
					t.Errorf("for item %d, expected %v, got %v", i, expectedData, result[i])
				}
			}
		})
	}
}
