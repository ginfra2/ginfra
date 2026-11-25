/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes deployments.  Utilities.

I tried to hijack nerdctl code and containerd client to do this internally,
but setting them up where extremely complex.  Using the docker client is way easier.  Maybe another day.
*/
package kube

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// #####################################################################################################################
// #

func getContainerId(name, tag string) (string, string, error) {

	var (
		b   []byte
		err error
	)

	cmd := exec.Command("nerdctl", "-n", "k8s.io", "images", "--digests", fmt.Sprintf("--filter=reference=%s:%s", name, tag))
	if b, err = cmd.CombinedOutput(); err != nil {
		return "", "", base.NewGinfraErrorChildAFlags("Cmd failed.", err, base.ERRFLAG_NOT_RECOVERABLE,
			base.LM_ERROR, string(b))
	}

	s := string(b)
	if !strings.Contains(s, name) {
		return "", "", base.NewGinfraErrorFlags("Image not present.", base.ERRFLAG_RECOVERABLE)
	}

	s = s[strings.Index(s, name):]
	ss := strings.Fields(s)
	if len(ss) < 4 {
		return "", "", base.NewGinfraErrorChildAFlags("Bad data from nerdctl", err, base.ERRFLAG_NOT_RECOVERABLE,
			base.LM_ERROR, s)
	}

	return ss[3], ss[2], nil
}

func pullContainer(name, tag string) error {

	var (
		b   []byte
		err error
	)

	cmd := exec.Command("nerdctl", "-n", "k8s.io", "pull", fmt.Sprintf("%s:%s", name, tag))
	if b, err = cmd.CombinedOutput(); err != nil {
		return base.NewGinfraErrorChildAFlags("Cmd failed.", err, base.ERRFLAG_NOT_RECOVERABLE,
			base.LM_ERROR, string(b))
	}
	return nil
}

func readContainerFile(id string) (string, error) {
	tfile := filepath.Join(base.GetGinfraTmpDir(), id)

	cmd := exec.Command("nerdctl", "-n", "k8s.io", "cp", id+":/specifics.json", tfile)
	if b, err := cmd.CombinedOutput(); err != nil {
		return "", base.NewGinfraErrorChildAFlags("Could not pluck specifics.json file.", err,
			base.ERRFLAG_NOT_FOUND, base.LM_ERROR, string(b))
	}
	defer func() {
		_ = os.Remove(tfile)
	}()

	b, err := os.ReadFile(tfile)
	if err != nil {
		// This should never happen
		panic(err)
	}

	return string(b), nil
}

func getSpecificsCacheFilePrefix(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}

func getSpecificsCacheFileName(id string, name string) string {
	return getSpecificsCacheFilePrefix(name) + "_" + strings.ReplaceAll(id, "/", "_")
}

func cacheContainerSpecifics(gctx *app.GContext, imageId, target string, values map[string]interface{}) (*shared.Specifics, error) {

	var (
		b   []byte
		err error
	)

	// Create container.
	cmd := exec.Command("nerdctl", "-n", "k8s.io", "create", imageId)
	if b, err = cmd.CombinedOutput(); err != nil {
		return nil, base.NewGinfraErrorChildA("Could not create image to pluck the file.", err,
			base.LM_ERROR, string(b))
	}
	cid := strings.TrimSpace(string(b))

	defer func() {
		cmdr := exec.Command("nerdctl", "-n", "k8s.io", "rm", cid)
		if b, err = cmdr.CombinedOutput(); err != nil {
			gctx.GetLogger().Warnw("Could not remove the container.", zap.Error(err),
				zap.String(base.LM_ID, base.SnipId(cid)))
		}
	}()

	// Get the file
	p := filepath.Join("/", app.PathSpecificsFile)
	s, err := readContainerFile(cid)
	if err != nil {
		if (base.GetErrorFlags(err) & base.ERRFLAG_NOT_FOUND) > 0 {
			return nil, base.NewGinfraErrorChildAFlags("Specifics file not found in container", err,
				base.ERRFLAG_NOT_FOUND&base.ERRFLAG_RECOVERABLE, base.LM_PATH, p)
		} else {
			return nil, base.NewGinfraErrorChildFlags("Could not get specifics from file", err,
				base.ERRFLAG_RECOVERABLE)
		}
	}

	sp, err := shared.SpecificsLoad(s, p, values)
	if err != nil {
		return nil, base.NewGinfraErrorChildFlags("Specifics file invalid.", err, base.ERRFLAG_RECOVERABLE)
	}

	// Cache it.
	if err = os.WriteFile(target, []byte(s), 0644); err != nil {
		err = base.NewGinfraErrorChildFlags("Could not cache file.", err, base.ERRFLAG_RECOVERABLE)
	}

	return sp, err
}

func nukeImage(name string) {
	n := name[strings.LastIndex(name, "_")+1:]
	cmd := exec.Command("nerdctl", "-n", "k8s.io", "rmi", "-f", n)
	_ = cmd.Run()
}

func getContainerSpecifics(gctx *app.GContext, image string, values map[string]interface{}) (*shared.Specifics, string, error) {

	var (
		cspc      *shared.Specifics
		err       error
		id, sha   string
		name, tag string
	)

	sp := strings.Split(image, ":")
	if len(sp) == 1 {
		name = sp[0]
		tag = "latest"
	} else if len(sp) == 2 {
		name = sp[0]
		tag = sp[1]
	} else {
		return nil, "", base.NewGinfraErrorA("Bad image spec", base.LM_VALUE, image)
	}

	id, sha, err = getContainerId(name, tag)
	if err != nil && base.GetErrorFlags(err) == base.ERRFLAG_RECOVERABLE {
		// Pull and try again.
		err = pullContainer(name, tag)
		if err == nil {
			id, sha, err = getContainerId(name, tag)
		}
	}
	if err != nil {
		return nil, "", err
	}

	// =======================================================================================
	// Check cache

	target := getSpecificsCacheFileName(id, image)
	dirl, err := os.ReadDir(base.GetGinfraTmpCacheDir())
	if err != nil {
		return nil, "", base.NewGinfraErrorChild("Cache dir is bad or missing.  Environment corrupt.  You may need to reinstall Ginfra.", err)
	}
	for _, i := range dirl {
		if i.Name() == target {
			cspc, _ = shared.SpecificsLoadFile(filepath.Join(base.GetGinfraTmpCacheDir(), target), values)
			break
		}
	}

	if cspc == nil {
		// Wipe out cache members with same name
		for _, i := range dirl {
			if strings.HasPrefix(i.Name(), getSpecificsCacheFilePrefix(name)) {
				_ = os.Remove(filepath.Join(base.GetGinfraTmpCacheDir(), target))
				nukeImage(i.Name()) // Opportunistic only.
			}
		}

		cspc, err = cacheContainerSpecifics(gctx, id, filepath.Join(base.GetGinfraTmpCacheDir(), target), values)
	}

	return cspc, sha, err
}
