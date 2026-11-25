//go:build functional_test

/*
Package tool
Copyright (c) 2025 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Test the testhost tools.
*/
package tool

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/common"
)

func TestLoadConfigProv(t *testing.T) {

	common.SetupTest()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(currentDir)
	}()

	// Table-driven tests
	tests := []struct {
		name         string
		setup        func() string // Setup function to control file environment
		expectedName string
		expectErr    bool
	}{
		{
			name: "valid hosting file and config provider",
			setup: func() string {
				// Setup temporary working directory
				dir := base.GetGinfraRandomTmpTestDirAndMake()
				hostFilePath := filepath.Join(dir, app.PathHostingFilePath)
				// Create valid hosting file
				err := common.WriteJsonFile(manage.HostFile{
					ConfigUrl:         "file://" + filepath.Join(base.GetGinfraHome(), "common", "config", "test", "1.json"),
					ConfigAccessToken: "valid-token",
					DeploymentName:    "test-deployment",
				}, hostFilePath, "hosting metadata")
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				_ = os.Chdir(dir)
				return dir
			},
			expectedName: "test-deployment",
			expectErr:    false,
		},
		{
			name: "missing hosting file",
			setup: func() string {
				// Setup temporary working directory without hosting file
				dir := base.GetGinfraRandomTmpTestDirAndMake()
				_ = os.Chdir(dir)
				return dir
			},
			expectedName: "",
			expectErr:    true,
		},
		{
			name: "invalid hosting file",
			setup: func() string {
				// Setup temporary working directory
				dir := base.GetGinfraRandomTmpTestDirAndMake()
				hostFilePath := filepath.Join(dir, app.PathHostingFilePath)
				// Create an invalid hosting file
				err := os.WriteFile(hostFilePath, []byte("{invalid-json}"), 0644)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				_ = os.Chdir(dir)
				return dir
			},
			expectedName: "",
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			dir := tt.setup()
			defer func() {
				_ = os.RemoveAll(dir)
			}()

			// Call the function under test
			_, name, err := loadConfigProv()

			// Validate results
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if name != tt.expectedName {
				t.Errorf("expected deployment name: %s, got: %s", tt.expectedName, name)
			}
		})
	}
}

type ErrorTestStruct struct {
	name        string
	errorType   ErrorType
	expectedErr string
}

func TestTestCases(t *testing.T) {
	common.SetupTest()

	err := RunService(44244)
	if err != nil {
		t.Error(err)
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(currentDir)
		_ = StopService()
	}()
	tf := filepath.Join(base.GetGinfraHome(), "app", "tool", "test", "test.txt")
	t.Logf("tf value: %s", tf)

	if err = os.Chdir(filepath.Join(base.GetGinfraHome(), "app", "tool", "test")); err != nil {
		t.Fatalf("Could not cd to test directory: %v", err)
	}

	// -- RUN ------------------------------------------------------------------------------------------------

	// I am not putting these in a table for now since the API is not stable.

	t.Run("Run - no file and no config", func(t *testing.T) {
		err = TestRun(app.GetRootContext(), "1", "xxx", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required")
	})

	gctx := app.GetCopyRootContext()
	t.Run("Run - No auth", func(t *testing.T) {
		gctx.Host = "http://localhost:44244"
		err = TestRun(&gctx, "1", "xxx", tf)
		assert.ErrorContains(t, err, "Authorization and host are required.")
	})

	t.Run("Run - OK", func(t *testing.T) {
		gctx.Authtoken = "AAAAAAAAAAAA"
		err = TestRun(&gctx, "1", "xxx", tf)
		assert.NoError(t, err)
	})

	tests := []ErrorTestStruct{
		{"Run - Not authorized", AUTHERROR, "Authorization is required"},
		{"Run - Internal server error", INTSRVERROR, "Could not run test"},
		{"Run - Forbidden", FORBIDDENERROR, "Bad User id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ForceTestError(tt.errorType)
			err = TestRun(&gctx, "1", "xxx", tf)
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

	t.Run("Run - bad host", func(t *testing.T) {
		gctx.Host = "http://x:44244"
		err = TestRun(&gctx, "1", "host", tf)
		assert.ErrorContains(t, err, "Error connecting to test host")
	})

	// -- STOP ------------------------------------------------------------------------------------------------

	t.Run("Stop - no file and no config", func(t *testing.T) {
		err = TestStop(app.GetRootContext(), "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required")
	})

	gctx = app.GetCopyRootContext()
	t.Run("Stop - No auth", func(t *testing.T) {
		gctx.Host = "http://localhost:44244"
		err = TestStop(&gctx, "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required.")
	})

	t.Run("Stop - OK", func(t *testing.T) {
		gctx.Authtoken = "AAAAAAAAAAAA"
		err = TestStop(&gctx, "1", "xxx")
		assert.NoError(t, err)
	})

	tests = []ErrorTestStruct{
		{"Stop - Not authorized", AUTHERROR, "Authorization is required"},
		{"Stop - Not found", NOTFOUNDERROR, "Id not found"},
		{"Stop - Internal server error", INTSRVERROR, "Internal server error"},
		{"Stop - Forbidden", FORBIDDENERROR, "User id (uid) is forbidden for this test id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ForceTestError(tt.errorType)
			err = TestStop(&gctx, "1", "xxx")
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

	t.Run("Stop - bad host", func(t *testing.T) {
		gctx.Host = "http://x:44244"
		err = TestStop(&gctx, "1", "host")
		assert.ErrorContains(t, err, "Error connecting to test host")
	})

	// -- STATUS ------------------------------------------------------------------------------------------------

	t.Run("Status - no file and no config", func(t *testing.T) {
		err = TestStatus(app.GetRootContext(), "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required")
	})

	gctx = app.GetCopyRootContext()
	t.Run("Status - No auth", func(t *testing.T) {
		gctx.Host = "http://localhost:44244"
		err = TestStatus(&gctx, "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required.")
	})

	t.Run("Status - OK", func(t *testing.T) {
		gctx.Authtoken = "AAAAAAAAAAAA"
		err = TestStatus(&gctx, "1", "xxx")
		assert.NoError(t, err)
	})

	tests = []ErrorTestStruct{
		{"Status - Not authorized", AUTHERROR, "Authorization is required"},
		{"Status - not found", NOTFOUNDERROR, "Id not found"},
		{"Status - Internal server error", INTSRVERROR, "Could not get status"},
		{"Status - Forbidden", FORBIDDENERROR, "User id (uid) is forbidden for this test id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ForceTestError(tt.errorType)
			err = TestStatus(&gctx, "1", "xxx")
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

	t.Run("Status - bad host", func(t *testing.T) {
		gctx.Host = "http://x:44244"
		err = TestStatus(&gctx, "1", "host")
		assert.ErrorContains(t, err, "Error connecting to test host")
	})

	// -- STATUS ALL ----------------------------------------------------------------------------------------------

	t.Run("Status All - no file and no config", func(t *testing.T) {
		err = TestAll(app.GetRootContext(), "1")
		assert.ErrorContains(t, err, "Authorization and host are required")
	})

	gctx = app.GetCopyRootContext()
	t.Run("Status All - No auth", func(t *testing.T) {
		gctx.Host = "http://localhost:44244"
		err = TestAll(&gctx, "1")
		assert.ErrorContains(t, err, "Authorization and host are required.")
	})

	t.Run("Status All - OK", func(t *testing.T) {
		gctx.Authtoken = "AAAAAAAAAAAA"
		err = TestAll(&gctx, "1")
		assert.NoError(t, err)
	})

	tests = []ErrorTestStruct{
		{"Status All - Not authorized", AUTHERROR, "Authorization is required"},
		{"Status All - Internal server error", INTSRVERROR, "Could not get status"},
		{"Status All - Forbidden", FORBIDDENERROR, "Bad User id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ForceTestError(tt.errorType)
			err = TestAll(&gctx, "1")
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

	t.Run("Status All - bad host", func(t *testing.T) {
		gctx.Host = "http://x:44244"
		err = TestAll(&gctx, "1")
		assert.ErrorContains(t, err, "Error connecting to test host")
	})

	// -- LOG ------------------------------------------------------------------------------------------------

	t.Run("Log - no file and no config", func(t *testing.T) {
		err = TestLog(app.GetRootContext(), "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required")
	})

	gctx = app.GetCopyRootContext()
	t.Run("Log - No auth", func(t *testing.T) {
		gctx.Host = "http://localhost:44244"
		err = TestLog(&gctx, "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required.")
	})

	t.Run("Log - OK", func(t *testing.T) {
		gctx.Authtoken = "AAAAAAAAAAAA"
		err = TestLog(&gctx, "1", "xxx")
		assert.NoError(t, err)
	})

	tests = []ErrorTestStruct{
		{"Log - Not authorized", AUTHERROR, "Authorization is required"},
		{"Log - not found", NOTFOUNDERROR, "Id not found"},
		{"Log - Internal server error", INTSRVERROR, "Could not get log"},
		{"Log - Forbidden", FORBIDDENERROR, "User id (uid) is forbidden for this test id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ForceTestError(tt.errorType)
			err = TestLog(&gctx, "1", "xxx")
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

	t.Run("Log - bad host", func(t *testing.T) {
		gctx.Host = "http://x:44244"
		err = TestLog(&gctx, "1", "host")
		assert.ErrorContains(t, err, "Error connecting to test host")
	})

	// -- REMOVE ------------------------------------------------------------------------------------------------

	t.Run("Remove - no file and no config", func(t *testing.T) {
		err = TestRemove(app.GetRootContext(), "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required")
	})

	gctx = app.GetCopyRootContext()
	t.Run("Remove - No auth", func(t *testing.T) {
		gctx.Host = "http://localhost:44244"
		err = TestRemove(&gctx, "1", "xxx")
		assert.ErrorContains(t, err, "Authorization and host are required.")
	})

	t.Run("Remove - OK", func(t *testing.T) {
		gctx.Authtoken = "AAAAAAAAAAAA"
		err = TestRemove(&gctx, "1", "xxx")
		assert.NoError(t, err)
	})

	t.Run("Remove - id required", func(t *testing.T) {
		err = TestRemove(&gctx, "1")
		assert.ErrorContains(t, err, "Id is required")
	})

	tests = []ErrorTestStruct{
		{"Remove - Not authorized", AUTHERROR, "Authorization is required"},
		{"Remove - not found", NOTFOUNDERROR, "Id not found"},
		{"Remove - Internal server error", INTSRVERROR, "Could not remove"},
		{"Remove - Forbidden", FORBIDDENERROR, "Bad User id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ForceTestError(tt.errorType)
			err = TestRemove(&gctx, "xxx", "1")
			assert.ErrorContains(t, err, tt.expectedErr)
		})
	}

	t.Run("Remove - bad host", func(t *testing.T) {
		gctx.Host = "http://x:44244"
		err = TestRemove(&gctx, "1", "host")
		assert.ErrorContains(t, err, "Error connecting to test host")
	})

}
