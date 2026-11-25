/*
Package common
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Nothing here yet, but keeping consistent with the pattern.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/

package base

// #####################################################################################################################
// # SIMPLE

type NV struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// #####################################################################################################################
// # METAFILE

const META__SERVICE_CLASS = "serviceclass"
