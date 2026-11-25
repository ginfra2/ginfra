//go:build integration_test

/*
Package generate
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Create and actually run a service.
*/
package generate

import (
	"bytes"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	common.SetupTest()
}

func testGenLiveRunGoService(name, stype string) error {
	var d []byte

	tg := filepath.Join(base.GetGinfraTmpTestDir(), "generate", name)
	err := os.MkdirAll(tg, 0777)
	if err != nil {
		panic(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(tg)
	if err != nil {
		panic(err)
	}
	defer os.Chdir(wd)

	err = GenerateService(app.GetRootContext(), name, name, name, stype, "gitlab.com/ginfra/gservices/"+name, base.GetGinfraTemplatesDir())

	if err != nil {
		return err
	}

	d, err = exec.Command("task", "stubs").Output()
	if err != nil {
		return base.NewGinfraErrorChildA("Failed refresh the stubs", err, base.LM_CAUSE, string(d))
	}

	_, err = exec.Command("go", "mod", "tidy").Output()
	if err != nil {
		return base.NewGinfraErrorChildA("Failed go mod tidy", err, base.LM_CAUSE, string(d))
	}

	d, err = exec.Command("task", "build").Output()
	if err != nil {
		return base.NewGinfraErrorChildA("Failed build", err, base.LM_CAUSE, string(d))
	}

	cmd := exec.Command(filepath.Join(tg, "build", "ginfra_service.exe"), tg)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	go func(c *exec.Cmd) {
		errr := c.Run()
		if errr != nil {
			es := errr.Error()
			if !strings.Contains(es, "kill") {
				panic(errr)
			}
		}
	}(cmd)
	defer func(c *exec.Cmd) {
		_ = c.Process.Kill()
	}(cmd)

	tries := 50
	var resp *http.Response
	for tries > 0 {
		resp, err = http.Get("http://localhost:8900/ginfra/specifics")
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		return base.NewGinfraErrorChild("Failed get specifics", err)
	}
	if resp.StatusCode != 200 {
		return base.NewGinfraErrorA("Did not get response code 200", base.LM_CODE, resp.StatusCode)
	}

	return nil
}

func TestGenLiveService(t *testing.T) {

	t.Run("APIECHO service", func(t *testing.T) {
		err := testGenLiveRunGoService("testapiecho", "apiecho")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("SERVER service", func(t *testing.T) {
		err := testGenLiveRunGoService("testserver", "server")
		if err != nil {
			t.Error(err)
		}
	})
}
