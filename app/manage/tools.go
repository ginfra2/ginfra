/*
Package manage
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Tools specific to management.
*/
package manage

import (
	"archive/tar"
	"gitlab.com/ginfra/ginfra/base"
	"strings"
	"time"
)

// #####################################################################################################################
// #

func StringToTar(fileName string, data string) (string, error) {
	b := []byte(data)
	return DataToTar(fileName, b)
}

func DataToTar(fileName string, data []byte) (string, error) {

	b := new(strings.Builder)
	tw := tar.NewWriter(b)

	// File header
	h := new(tar.Header)
	h.Typeflag = tar.TypeReg
	h.Size = int64(len(data))
	h.Mode = int64(0x666)
	h.ModTime = time.Now()
	h.Name = fileName
	err := tw.WriteHeader(h)
	if err != nil {
		return "", base.NewGinfraErrorChild("Failed to create tar header.", err)
	}

	// File body
	_, err = tw.Write(data)
	if err != nil {
		return "", base.NewGinfraErrorChild("Failed to write to tar.", err)
	}

	return b.String(), nil
}
