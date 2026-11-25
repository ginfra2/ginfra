//go:build functional_test

/*
Package generate
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Generate services.
*/
package generate

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

func TestGenServProcessNewDir(t *testing.T) {
	/*
		Test fails if you are root.  Need to figure a different way to do it.
		t.Run("T1 Cannot read a file", func(t *testing.T) {
			err := common.CopyFiles(filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T1"),
				filepath.Join(base.GetGinfraTmpTestDir(), "T1"), true, true, true)
			if err != nil {
				t.Error(err)
				return
			}

			target := filepath.Join(base.GetGinfraTmpTestDir(), "T1", "README.md")
			err = os.Chmod(target, 222)
			if err != nil {
				t.Error(err)
				return
			}
			defer func(t string) {
				err = os.Remove(t)
				if err != nil {
					panic(err)
				}
			}(target)

			assert.Panics(t, func() {
				err = genServProcessNewDir(app.GetRootContext(), filepath.Join(base.GetGinfraTmpTestDir(), "T1"),
					filepath.Join(base.GetGinfraTmpTestDir(), "dT1"), map[string]interface{}{})
			}, "Cannot read file should have panicked.")
		})
	*/
	t.Run("T2 Bad file replacer", func(t *testing.T) {
		err := genServProcessNewDir(app.GetRootContext(), filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T2"),
			filepath.Join(base.GetGinfraTmpTestDir(), "dT2"), map[string]interface{}{})
		assert.ErrorContains(t, err, "Failed to process file")
	})
}

func TestGenUpdateService(t *testing.T) {
	t.Run("T3 Update, missing specifics", func(t *testing.T) {
		gctx := &app.GContext{
			ServiceHomePath: filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T3"),
		}
		err := UpdateService(gctx, "xxx")
		assert.ErrorContains(t, err, "specifics.json: no such file or directory ")
	})

	t.Run("T4 Update, missing specifics", func(t *testing.T) {
		gctx := &app.GContext{
			ServiceHomePath: filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T4"),
		}
		err := UpdateService(gctx, "xxxx")
		assert.ErrorContains(t, err, "Specifics file does not exist")
	})
}

func TestGenUpdateServiceMerge(t *testing.T) {
	t.Run("T3 Update, no metafile", func(t *testing.T) {
		src := filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T5b")
		dest := filepath.Join(base.GetGinfraTmpTestDir(), "generate", "T5b")

		err := common.CopyFiles(src, dest, false, true, true)
		if err != nil {
			t.Error(err)
			return
		}

		err = updateServProcessNewDir(app.GetRootContext(), dest, src, map[string]interface{}{})
		assert.ErrorContains(t, err, "Dangling block marker")
	})
}

func TestGenCases(t *testing.T) {

	t.Run("template names", func(t *testing.T) {
		m, err := getTemplateNames(base.GetGinfraTemplatesDir())
		assert.NoError(t, err)
		assert.Less(t, 2, len(m))
	})

	t.Run("template bad dir", func(t *testing.T) {
		_, err := getTemplateNames(filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T1"))
		assert.Error(t, err)
	})

	t.Run("template bad dir path", func(t *testing.T) {
		_, err := getTemplateNames(filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "TDOESNOTEXIST"))
		assert.Error(t, err)
	})

	t.Run("template not a dir", func(t *testing.T) {
		err := GenerateService(app.GetRootContext(), "xx", "xx", "xx", "server",
			"xx", filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T6"))
		assert.ErrorContains(t, err, "Template path is not a directory")
	})
}

func TestUpdateCases(t *testing.T) {

	cd, _ := os.Getwd()
	defer os.Chdir(cd)
	err := os.Chdir(filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T6"))
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("template dir bad", func(t *testing.T) {
		err = UpdateService(app.GetRootContext(), "/wefoqij2345234t235gherugijorijvjoerjojqnjeojngkjengbenkrbneofkvanvcmscjfhi")
		assert.ErrorContains(t, err, "Template path does not exist or is not accessible")
	})

	t.Run("template dir is a file", func(t *testing.T) {
		err = UpdateService(app.GetRootContext(), filepath.Join(base.GetGinfraHome(), "app", "generate", "test", "T6"))
		assert.ErrorContains(t, err, "Template path is not a directory")
	})

	t.Run("dest not found", func(t *testing.T) {
		_, err = updateServProcessFile("xxx", "/SdfadfasdvasdvqawefvqerwSDZvcasdfqwefvasdvasdv")
		assert.ErrorContains(t, err, "Could not read original destination file")
	})

}
