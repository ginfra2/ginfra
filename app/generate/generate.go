/*
Package generate
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Generate services from templates.
*/
package generate

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"os"
	"path/filepath"
	"strings"
)

func getTemplateNames(path string) (map[string]struct{}, error) {
	var r = make(map[string]struct{})

	e, err := os.ReadDir(path)
	if err != nil {
		return r, base.NewGinfraErrorChildA("Bad template directory for generate.", err, base.LM_PATH,
			path)
	}

	for _, i := range e {
		if i.IsDir() {
			r[i.Name()] = struct{}{}
		}
	}

	if len(r) == 0 {
		return r, base.NewGinfraErrorChildA("Bad template directory for generate.  Does not contain templates.",
			err, base.LM_PATH, path)
	}

	return r, nil
}

// #####################################################################################################################
// # FIRST TIME GENERATE

// GenerateService generates a service scaffold based on the provided parameters and templates directory.
// It validates the given template path and processes it to generate the appropriate service structure.
// Returns an error if the template path is invalid or the service generation fails.
func GenerateService(gctx *app.GContext, name, serviceClass, servicePath, serviceType, moduleName, templatePath string) error {

	v := map[string]interface{}{}
	v["ServiceName"] = name
	v["ServiceType"] = serviceType
	v["ServiceClass"] = serviceClass
	v["ServicePath"] = servicePath
	v["ServiceNameLower"] = strings.ToLower(name)
	v["Module"] = moduleName
	v["ModuleFinalName"] = filepath.Base(moduleName) // not sure about this one.
	v["GINFRA_HOME"] = base.GetGinfraHome()

	assets := filepath.Join(templatePath, serviceType)
	fi, err := os.Stat(assets)
	if err != nil {
		return base.NewGinfraErrorA("Template path does not exist or is not accessible.  Does the template exist for the service type?",
			base.LM_PATH, assets, base.LM_SERVICE_TYPE, serviceType)
	}
	if !fi.IsDir() {
		return base.NewGinfraErrorA("Template path is not a directory.", base.LM_PATH, assets)
	}

	return generateServiceRoot(gctx, assets, v)
}

func genServProcessNewDir(gctx *app.GContext, srcRoot, destRoot string, values map[string]interface{}) error {

	var (
		cfile string
	)

	err := filepath.Walk(srcRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if path != srcRoot {
					dir := filepath.Join(destRoot, path[len(srcRoot):])
					err = common.MakeDir(dir)
				}

			} else {
				df, flags := common.GetDestFilepath(destRoot, srcRoot, path, true, true, true)

				cfile = path
				if err == nil {

					if flags&common.FileFlag_NoTemplate > 0 {
						// Dont process template
						err = common.CopyFile(path, df)

					} else {

						// Get mode from template
						var st os.FileInfo
						st, err = os.Stat(path)
						if err == nil {

							// Process template
							var in []byte
							in, err = os.ReadFile(path)
							if err != nil {
								panic(fmt.Sprintf("Could not read file %s : %v", path, err))
							}

							var ous string
							ins := string(in)
							ous, err = common.SimpleReplacer(ins, values, false)
							if err == nil {
								err = os.WriteFile(df, []byte(ous), st.Mode())
							}
						}

					} // end if not excluded.
				}

				if err == nil {
					gctx.GetLogger().Debugw("Processing file successful.", base.LM_FILEPATH, path)
				} else {
					gctx.GetLogger().Debugw("Processing file failed.", base.LM_FILEPATH, path, base.LM_ERROR, err)
				}
			}

			return err
		})

	if err != nil {
		err = base.NewGinfraErrorChildA("Failed to process file.", err, base.LM_FILEPATH, cfile)
	}
	return err
}

func generateServiceRoot(gctx *app.GContext, assets string, values map[string]interface{}) error {
	var (
		err error = nil
		d   string
	)

	defer func() {
		if r := recover(); r != nil {
			err = base.NewGinfraErrorA("Giving up due to excessive errors or panic.", base.LM_CAUSE, r)
		}
	}()

	d, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	if err == nil {
		err = genServProcessNewDir(gctx, assets, d, values)
	}
	return err
}

// #####################################################################################################################
// # SUBSEQUENT UPDATES

// UpdateService updates a service using provided templates and specifics, handling paths, assets, and configuration.
// It returns an error if specifics loading or file structure setup fails.
func UpdateService(gctx *app.GContext, templates string) error {

	// Get specifics
	dest, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	spec, err := shared.SpecificsLoadFile(filepath.Join(dest, base.SpecificsFileName), map[string]interface{}{})
	if err != nil {
		return err
	}
	if spec == nil {
		panic(err)
	}

	assets := filepath.Join(templates, spec.Meta.Type)
	fi, err := os.Stat(assets)
	if err != nil {
		return base.NewGinfraErrorA("Template path does not exist or is not accessible.", base.LM_PATH, assets)
	}
	if !fi.IsDir() {
		return base.NewGinfraErrorA("Template path is not a directory.", base.LM_PATH, assets)
	}

	v := map[string]interface{}{}
	v["ServiceName"] = spec.Meta.Name
	v["Module"] = spec.Meta.ModuleName
	v["GINFRA_HOME"] = base.GetGinfraHome()

	err = updateServiceRoot(gctx, dest, assets, v)

	if err != nil {
		gctx.GetLogger().Info("IMPORTANT!!!  IMPORTANT!!!  Run 'go mod tidy' now.")
		gctx.GetLogger().Info("IMPORTANT!!!  IMPORTANT!!!  You may need to generate your apis again.   Unless you did something custom, running 'task stubs' should do this.")
	}

	return err
}

type UpdateState int64

const GenTerminalMarker = "// ^^"
const GenTerminalMarkerSh = "##^"
const GenBlockMarker = "// >"

func updateConsumeUntilMarker(s *bufio.Scanner, b *strings.Builder, marker, fpath string) error {
	var (
		err   error = nil
		found       = false
	)
	for s.Scan() {
		t := s.Text()
		b.WriteString(t)
		b.WriteString("\n")
		if strings.HasPrefix(strings.Trim(t, " \t"), marker) {
			found = true
			break
		}
	}
	if !found {
		err = base.NewGinfraErrorA("Dangling block marker in file", base.LM_MARKER, marker, base.LM_FILEPATH, fpath)
	}

	return err
}

func updateServProcessFile(src string, destf string) (*bytes.Buffer, error) {

	var (
		blocks   = make(map[string]string)
		terminal string
		srcb     = bytes.NewBuffer([]byte(src))
	)

	// --- Read the destination. ----------------------------------------------------------------------------------
	inb, err := os.ReadFile(destf)

	if err != nil {
		return nil, base.NewGinfraErrorChildA("Could not read original destination file.", err, base.LM_FILEPATH, destf)
	}
	d := bufio.NewScanner(bytes.NewBuffer(inb))
	for d.Scan() {
		t := d.Text()

		trim := strings.Trim(t, " \t")
		if strings.HasPrefix(trim, GenTerminalMarker) || strings.HasPrefix(trim, GenTerminalMarkerSh) {
			var b strings.Builder
			t := d.Text()
			b.WriteString(t)
			b.WriteString("\n")
			for d.Scan() {
				t := d.Text()
				b.WriteString(t)
				b.WriteString("\n")
			}
			terminal = b.String()

		} else if strings.HasPrefix(trim, GenBlockMarker) {
			tag := trim[len(GenBlockMarker) : len(GenBlockMarker)+1]
			var b strings.Builder
			b.WriteString(t)
			b.WriteString("\n")
			err = updateConsumeUntilMarker(d, &b, GenBlockMarker, destf)
			if err != nil {
				return nil, err
			}
			blocks[tag] = b.String()
		}
	}

	// --- Read the source and merge ----------------------------------------------------------------------------------
	s := bufio.NewScanner(srcb)
	var o bytes.Buffer

	inSection := false
	for s.Scan() {
		t := s.Text()

		trim := strings.Trim(t, " \t")
		if strings.HasPrefix(trim, GenTerminalMarker) || strings.HasPrefix(trim, GenTerminalMarkerSh) {
			// Drain first
			for s.Scan() {
			}
			o.WriteString(terminal)

		} else if strings.HasPrefix(trim, GenBlockMarker) {
			if inSection {
				inSection = false
			} else {
				inSection = true
				tag := trim[len(GenBlockMarker) : len(GenBlockMarker)+1]
				if v, ok := blocks[tag]; ok {
					o.WriteString(v)
				}
			}

		} else {
			if !inSection {
				o.WriteString(t)
				o.WriteString("\n")
			}
		}
	}

	if inSection {
		err = base.NewGinfraErrorA("Dangling section.   File is corrupt.", base.LM_FILEPATH, destf)
	}

	return &o, err
}

func procUpdateServProcessNewDir(srcf, destf string, values map[string]interface{}, flags int) error {
	var (
		b   = bytes.NewBufferString("") // Probably not the neatest way to do this.
		err error
	)

	if flags&common.FileFlag_NoTemplate > 0 {
		// Dont process template
		err = common.CopyFile(srcf, destf)

	} else {
		// Process template.
		var st os.FileInfo
		st, err = os.Stat(srcf)
		if err != nil {
			return nil
		}

		// Process template
		in, err := os.ReadFile(srcf)
		if err != nil {
			panic(fmt.Sprintf("Could not read file %s : %v", srcf, err))
		}

		var ous string
		ins := string(in)
		ous, err = common.SimpleReplacer(ins, values, false)

		// Is there a destination file already?  If so, merge.  If not, just write it.
		var dm os.FileInfo
		if dm, err = os.Stat(destf); err == nil {
			st = dm
			b, err = updateServProcessFile(ous, destf)
		} else {
			// It's not there so don't worry about the error.
			err = nil
		}
		if err != nil {
			return err
		}

		err = os.WriteFile(destf, b.Bytes(), st.Mode())
		if err != nil {
			err = base.NewGinfraErrorChildA("Could not open or write to destination file.", err, base.LM_FILEPATH, destf)
		}
	}

	return err
}

func updateServProcessNewDir(gctx *app.GContext, destRoot, srcRoot string, values map[string]interface{}) error {

	var (
		cfile string
	)

	err := filepath.Walk(srcRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			df, flags := common.GetDestFilepath(destRoot, srcRoot, path, true, true,
				true)

			newfile := false
			if _, err = os.Stat(df); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					newfile = true
					err = nil
				} else {
					return base.NewGinfraErrorChildA("Strange file error.", err, base.LM_PATH, df)
				}
			}

			if flags&common.FileFlag_NoUpdate == 0 {

				if info.IsDir() {
					if path != srcRoot && newfile {
						err = common.MakeDir(df)
					}

				} else {
					err = procUpdateServProcessNewDir(path, df, values, flags)

					if err == nil {
						gctx.GetLogger().Debugw("Update file successful.", base.LM_FILEPATH, path, base.LM_PATH, df)
					} else {
						gctx.GetLogger().Debugw("Update file failed.", base.LM_FILEPATH, path, base.LM_ERROR, err)
					}
				}
			}

			return err
		})

	if err != nil {
		err = base.NewGinfraErrorChildA("Failed to update file.", err, base.LM_FILEPATH, cfile)
	}
	return err
}

func updateServiceRoot(gctx *app.GContext, dest, assets string, values map[string]interface{}) error {
	var (
		err error = nil
	)

	defer func() {
		if r := recover(); r != nil {
			err = base.NewGinfraErrorA("Giving up due to excessive errors or panic.", base.LM_CAUSE, r)
		}
	}()

	if err == nil {
		err = updateServProcessNewDir(gctx, dest, assets, values)
	}
	return err
}
