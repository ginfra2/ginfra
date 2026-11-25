/*
Package run
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Test for Utils.
*/
package run

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"testing"
)

func TestRunHostFile(t *testing.T) {

	err := common.CopyFiles(filepath.Join(base.GetGinfraHome(), "app", "run", "test", "hostingfile"),
		filepath.Join(base.GetGinfraTmpTestDir(), "run", "hostingfile"), false, false, false)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Good hosting file", func(t *testing.T) {
		var buf bytes.Buffer
		var gctx, _ = app.NewContext(&buf)
		trg := filepath.Join(base.GetGinfraTmpTestDir(), "run", "hostingfile", "goodhostingfile1")
		_ = os.Rename(filepath.Join(trg, "xHosting.json"), filepath.Join(trg, "Hosting.json"))
		gctx.HostDirPath = filepath.Join(trg)
		var sctx = ScriptContext{HostedSeen: false}

		checkHostFile(gctx, &sctx)
		sp, _ := testutil.NewScanPack(buf.String(), nil)
		assert.Equal(t, sp.Seek("Could not load"), -1)
	})

	t.Run("Bad hosting file - format", func(t *testing.T) {
		var buf bytes.Buffer
		var gctx, _ = app.NewContext(&buf)
		trg := filepath.Join(base.GetGinfraTmpTestDir(), "run", "hostingfile", "badhostingfile1")
		_ = os.Rename(filepath.Join(trg, "xHosting.json"), filepath.Join(trg, "Hosting.json"))
		gctx.HostDirPath = filepath.Join(trg)
		var sctx = ScriptContext{HostedSeen: false}

		checkHostFile(gctx, &sctx)
		sp, _ := testutil.NewScanPack(buf.String(), nil)
		assert.Greater(t, sp.Seek("not load hosting"), 0)
	})

	t.Run("Bad hosting file - config provider", func(t *testing.T) {
		var buf bytes.Buffer
		var gctx, _ = app.NewContext(&buf)
		trg := filepath.Join(base.GetGinfraTmpTestDir(), "run", "hostingfile", "badhostingfile2")
		_ = os.Rename(filepath.Join(trg, "xHosting.json"), filepath.Join(trg, "Hosting.json"))
		gctx.HostDirPath = filepath.Join(trg)
		var sctx = ScriptContext{HostedSeen: false}

		checkHostFile(gctx, &sctx)
		sp, _ := testutil.NewScanPack(buf.String(), nil)
		assert.Greater(t, sp.Seek(" provider could not be loaded"), 0)
	})

}
