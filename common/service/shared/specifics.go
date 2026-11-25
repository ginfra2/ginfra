/*
Package shared
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Service specifics.
*/
package shared

import (
	"encoding/json"
	"errors"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"os"
)

// #####################################################################################################################
// # TYPES

// Specifics represent structured data containing exposed elements and metadata as part of its configuration.
// Exposed contains the services the service wishes to expose.
// Meta encapsulates metadata attributes including version, name, type, module name, service class, and invocation command.
type Specifics struct {
	Exposed map[string]Expose `json:"exposed"`
	Meta    Meta              `json:"meta"`
}

type Meta struct {
	Version       int    `json:"version"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	ModuleName    string `json:"module_name"`
	ServiceClass  string `json:"serviceclass"`
	InvocationCmd string `json:"invocationcmd"`
}

// Expose represents a structure that defines a service and its associated service class configuration.
type Expose struct {
	Name         string       `json:"name,omitempty"`
	ServiceClass ServiceClass `json:"serviceclass,omitempty"`
}

// ServiceClass represents a service with a name and URL configuration for discovery and/or service endpoint.
type ServiceClass struct {
	Name string `json:"name,omitempty"`
	Url  string `json:"url,omitempty"`
}

// #####################################################################################################################
// # Func

// SpecificsLoad loads and parses specifics data from a JSON string, validates required fields, and replaces placeholders
// (wrapped in double parenthesis) from values.
func SpecificsLoad(s string, path string, values map[string]interface{}) (*Specifics, error) {
	var spec Specifics

	sr, err := common.SimpleReplacer(s, values, true)
	if err != nil {
		return nil, base.NewGinfraErrorChildA("Could not load spec due to replacement problem",
			err, base.LM_PATH, path)
	}

	err = json.Unmarshal([]byte(sr), &spec)
	if err != nil {
		return nil, base.NewGinfraErrorChildA("Could not load spec", err, base.LM_PATH, path)
	}

	// Validate required
	if spec.Meta.Name == "" {
		return nil, base.NewGinfraErrorA("Specifics file does not contain a name.", base.LM_PATH, path)
	}
	if spec.Meta.Type == "" {
		return nil, base.NewGinfraErrorA("Specifics file does not contain a type.", base.LM_PATH, path)
	}
	if spec.Meta.ServiceClass == "" {
		return nil, base.NewGinfraErrorA("Specifics file does not contain a serviceclass.", base.LM_PATH, path)
	}

	return &spec, nil
}

// SpecificsLoadFile reads a file from the given path, parses its content as specifics, and replaces placeholders.
// Returns a pointer to Specifics or an error if the file cannot be read, does not exist, or is invalid.
func SpecificsLoadFile(path string, values map[string]interface{}) (*Specifics, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, base.NewGinfraErrorChildAFlags("Specifics file does not exist.", err,
				base.ERRFLAG_NOT_FOUND, base.LM_PATH, path)
		} else {
			return nil, base.NewGinfraErrorChildA("Could not read specifics file.", err, base.LM_PATH, path)
		}
	}
	return SpecificsLoad(string(b), path, values)
}

// WriteFile writes the Specifics struct to a JSON file at the given path and returns an error if any operation fails.
func (s *Specifics) WriteFile(path string) error {
	d, err := json.Marshal(s)
	if err != nil {
		// When marshalling a struct, it is very, very difficult to get an error.
		return base.NewGinfraErrorChildA("Could not get data from specifics", err)
	}

	if err = os.WriteFile(path, d, 0644); err != nil {
		return base.NewGinfraErrorChildA("Could not write specifics file", err, base.LM_PATH, path)
	}

	return nil
}

// SpecificsProcess is a bit hackish now as it just finds and replaces at <<<port>>> entries.
func SpecificsProcess(s *Specifics, values map[string]interface{}) (*Specifics, error) {
	var r Specifics
	r.Meta = s.Meta
	r.Exposed = make(map[string]Expose)

	for n, v := range s.Exposed {
		n, _ = common.SimpleReplacer(n, values, false) // Only keep exposed for variables that are present.
		if n != "" {
			v.ServiceClass.Url, _ = common.SimpleReplacer(v.ServiceClass.Url, values, true)
			r.Exposed[n] = v
		}
	}

	return &r, nil
}
