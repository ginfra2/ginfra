/*
Package ginterfaces
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Merge api spec json files.  Broader merge functionality.  Newest entries overwrite older entries.
*/
package ginterfaces

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

func openapiFindReadYamlFile(interfaceDir, name string) (map[string]interface{}, error) {
	y := make(map[string]interface{})
	var (
		fp string
	)

	// If interfaceDir is given, check first.
	if interfaceDir != "" {
		cfp := filepath.Join(interfaceDir, name)
		info, err := os.Stat(cfp)
		if err == nil && !info.IsDir() {
			fp = cfp
		}
	}

	// Now try cwd
	if fp == "" {
		info, err := os.Stat(name)
		if err == nil && !info.IsDir() {
			fp = name
		}
	}

	// Last try ginfra interface dir
	if fp == "" {
		cfp := filepath.Join(base.GetGinfraGinterfacesDir(), name)
		info, err := os.Stat(cfp)
		if err == nil && !info.IsDir() {
			fp = cfp
		}
	}

	if fp == "" {
		return y, base.NewGinfraErrorA("Does not exist.", base.LM_NAME, name)
	}

	raw, err := os.ReadFile(fp)
	if err == nil {
		err = yaml.Unmarshal(raw, &y)
	}
	if err != nil {
		err = base.NewGinfraErrorChild("Bad yaml.", err)
	}

	return y, err
}

func openApiSeekRef(object interface{}, resolve map[string]struct{}) {

	rObj, ok := object.(map[string]interface{})
	if ok {

		for key, value := range rObj {
			//$ref: 'gi_common.yaml#/components/schemas/giusername'
			if rStr, ok := value.(string); ok {
				if key == "$ref" && strings.Contains(rStr, ".yaml#") {
					// Fix
					newValue := rStr[strings.Index(rStr, "#"):]
					rObj[key] = newValue

					// Add to things we have to resolve.
					resolve[rStr[:strings.Index(rStr, "#")]] = struct{}{}
				}

			} else if mapObj, ok := value.(map[string]interface{}); ok {
				seekRef(mapObj, resolve)

			} else if arrObj, ok := value.([]interface{}); ok {
				for _, item := range arrObj {
					seekRef(item, resolve)
				}
			}
		}

	}
}

func openapiMergeObjectSection(name string, in any, target map[string]interface{}, overwrite []string) ([]string, error) {

	if m, ok := in.(map[string]interface{}); ok {

		for k, v := range m {
			if _, ok = target[k]; ok {
				overwrite = append(overwrite, fmt.Sprintf("Overwrite: %s in %s", k, name))
			}
			target[k] = v
		}

	} else {
		return overwrite, base.NewGinfraErrorA("Source section is not an object.", base.LM_NAME, name)
	}

	return overwrite, nil
}

func openapiMergeArraySection(name string, in any, target []interface{}) ([]interface{}, error) {

	if m, ok := in.([]interface{}); ok {

		for _, v := range m {
			target = append(target, v)
		}

	} else {
		return nil, base.NewGinfraErrorA("Source section is not an array.", base.LM_NAME, name)
	}

	return target, nil
}

func openapiGetOrMakeObject(name string, in map[string]interface{}) (map[string]interface{}, error) {
	var (
		out map[string]interface{}
		err error
	)

	if i, ok := in[name]; ok {
		if im, okk := i.(map[string]interface{}); okk {
			out = im
		} else {
			err = base.NewGinfraErrorA("Section is not an object.", base.LM_NAME, name)
		}
	} else {
		out = make(map[string]interface{})
	}

	return out, err
}

func openapiGetOrMakeArray(name string, in map[string]interface{}) ([]interface{}, error) {
	var (
		out []interface{}
		err error
	)

	if i, ok := in[name]; ok {
		if im, okk := i.([]interface{}); okk {
			out = im
		} else {
			err = base.NewGinfraErrorA("Section is not an array.", base.LM_NAME, name)
		}
	} else {
		out = make([]interface{}, 0)
	}

	return out, err
}

func openapiMergeSpecs(name string, in map[string]interface{}, merged map[string]interface{}, overwrite []string) ([]string, error) {
	var (
		err error
		ms  map[string]interface{}
		as  []interface{}
	)

	for k, v := range in {
		switch k {

		case "openapi", "jsonSchemaDialect", "externalDocs":
			merged[k] = v

		case "info", "webhooks":
			// Ignore any overwrites for this section.
			if ms, err = openapiGetOrMakeObject(k, merged); err != nil {
				return overwrite, err
			}
			if _, err = openapiMergeObjectSection(k, v, ms, overwrite); err != nil {
				return overwrite, err
			}
			merged[k] = ms

		case "paths":
			if ms, err = openapiGetOrMakeObject(k, merged); err != nil {
				return overwrite, err
			}
			if overwrite, err = openapiMergeObjectSection(k, v, ms, overwrite); err != nil {
				return overwrite, err
			}
			merged[k] = ms

		case "security", "tags", "servers":
			if as, err = openapiGetOrMakeArray(k, merged); err != nil {
				return overwrite, err
			}
			if as, err = openapiMergeArraySection(k, v, as); err != nil {
				return overwrite, err
			}
			merged[k] = as

		case "components":
			if vv, ok := v.(map[string]interface{}); ok {

				// Component level
				mm, errr := openapiGetOrMakeObject(k, merged)
				if errr != nil {
					return overwrite, errr
				}
				for kkk, vvv := range vv {

					// Inner item level
					if vvvv, okkk := vvv.(map[string]interface{}); okkk {

						if ms, err = openapiGetOrMakeObject(kkk, mm); err != nil {
							return overwrite, err
						}
						if overwrite, err = openapiMergeObjectSection(k+":"+kkk, vvvv, ms, overwrite); err != nil {
							// This probably can't happen.
							return overwrite, err
						}
						mm[kkk] = ms

					} else {
						return overwrite, base.NewGinfraErrorA("Section in Components is not an Object.", base.LM_NAME, name)
					}
				}

				merged[k] = mm
			} else {
				return overwrite, base.NewGinfraErrorA("Components not an object.", base.LM_NAME, name)
			}

		} // end switch
	}

	return overwrite, err
}

type openapiSpec struct {
	Name     string
	Loaded   bool
	Resolved bool
	Spec     map[string]interface{}
}

// OpenApiLoadSpecs loads and merges OpenAPI specifications, resolving references and applying service name prefixes.
func OpenApiLoadSpecs(interfaceDir, serviceName string, noControl bool, names []string) (map[string]interface{}, []string, error) {

	var (
		m         map[string]interface{}
		load      = make([]openapiSpec, 0)
		specs     = make(map[string]openapiSpec)
		merged    = make(map[string]interface{})
		err       error
		overwrite []string
	)

	if !noControl {
		names = append(names, "ginfra.yaml") // Always merge the control interface.
	}

	// Specified loads
	for _, n := range names {
		if n != "" {
			specs[n] = openapiSpec{Name: n}
		}
	}

	seeking := true
	for seeking {
		seeking = false

		// Loads
		for k, v := range specs {
			if !v.Loaded {
				seeking = true
				if m, err = openapiFindReadYamlFile(interfaceDir, v.Name); err != nil {
					return nil, overwrite, err
				}
				v.Spec = m
				v.Loaded = true
				specs[k] = v
				load = append([]openapiSpec{v}, load...) // will merge deepest to shallowest.

			}
		}

		// Find resolves
		r := make(map[string]struct{})
		for k, v := range specs {
			// Find new resolves.
			if !v.Resolved {
				seeking = true
				openApiSeekRef(v.Spec, r)
				v.Resolved = true
				specs[k] = v
			}
		}

		// Load any new resolves.
		for k, _ := range r {
			if _, ok := specs[k]; !ok {
				specs[k] = openapiSpec{Name: k}
			}
		}
	}

	// Everything is loaded, so now merge
	for _, s := range load {

		if overwrite, err = openapiMergeSpecs(s.Name, s.Spec, merged, overwrite); err != nil {
			return merged, overwrite, err
		}
	}

	// Prefix everything with the service name if it isn't already.
	/*
		if paths, ok := merged["paths"].(map[string]interface{}); ok {
			for k, v := range paths {
				if !strings.HasPrefix(k, "/"+serviceName+"/") {
					nk := strings.Replace("/"+serviceName+"/"+k, "//", "/", -1)

					delete(paths, k)
					if _, ok := paths[nk]; ok {
						return nil, overwrite,
							base.NewGinfraError(fmt.Sprintf("Collision while prefixing '%s' to become '%s'", k, serviceName+"/"+k))
					} else {
						paths[nk] = v
					}
				}
			}
		}
	*/

	return merged, overwrite, err
}

// OpenApiLoadSaveSpecs loads and merges OpenAPI specifications, then saves the result to the specified destination file.
func OpenApiLoadSaveSpecs(interfaceDir, serviceName, dest string, noControl bool, names []string) ([]string, error) {
	var (
		b         []byte
		overwrite []string
	)
	i, overwrite, err := OpenApiLoadSpecs(interfaceDir, serviceName, noControl, names)
	if err == nil {
		b, err = yaml.Marshal(i)
		if err == nil {
			err = os.WriteFile(dest, b, 0666)
		}
	}
	return overwrite, err
}
