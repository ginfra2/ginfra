/*
Package shared
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

TLS/crypto bundle support.
*/
package shared

import (
	"archive/zip"
	"bytes"
	"gitlab.com/ginfra/ginfra/base"
	"io"
	"os"
	"path/filepath"
)

// #####################################################################################################################
// # VALUES AND TYPES

const BundleFileName = "bundle.zip"
const BundleServerCert = "server.cert"
const BundleServerKey = "server.key"

// #####################################################################################################################
// # IMPLEMENTATION

func BundleDetect(root string) string {
	stat, err := os.Stat(filepath.Join(root, BundleFileName))
	if err == nil && !stat.IsDir() {
		return filepath.Join(root, BundleFileName)
	}
	if stat, err = os.Stat(BundleFileName); err == nil && !stat.IsDir() {
		return BundleFileName
	}
	if stat, err = os.Stat(filepath.Join("/", BundleFileName)); err == nil && !stat.IsDir() {
		return filepath.Join("/", BundleFileName)
	}
	return ""
}

func BundleExtractFiles(bundle, destpath string, items []string) error {

	if len(items) == 0 {
		return nil
	}

	a, err := zip.OpenReader(bundle)
	if err != nil {
		return base.NewGinfraErrorChild("Bundle not accessible", err)
	}
	defer func(a *zip.ReadCloser) {
		_ = a.Close()
	}(a)

	var (
		inr  io.ReadCloser
		outr io.WriteCloser
	)
	found := 0
	for _, f := range a.File {
		for _, i := range items {
			if f.Name == i {
				if inr, err = f.Open(); err != nil {
					return base.NewGinfraErrorChild("Could not access file in bundle", err)
				}
				defer func(inr io.ReadCloser) {
					_ = inr.Close()
				}(inr)

				if outr, err = os.OpenFile(filepath.Join(destpath, i), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()); err != nil {
					return base.NewGinfraErrorChildA("Could open destination file", err, base.LM_FILEPATH, filepath.Join(destpath, i))
				}
				defer func(outr io.WriteCloser) {
					_ = outr.Close()
				}(outr)

				if _, err = io.Copy(outr, inr); err != nil {
					return base.NewGinfraErrorChildA("Could write to destination file", err, base.LM_FILEPATH, filepath.Join(destpath, i))
				}

				found++
			}
		}
	}

	if found < len(items) {
		return base.NewGinfraError("Not all items found.")
	}
	return nil
}

type ExtractData []byte

func BundleExtract(bundle string, items []string) ([]ExtractData, error) {

	r := make([]ExtractData, 2)

	if len(items) == 0 {
		return r, nil
	}

	a, err := zip.OpenReader(bundle)
	if err != nil {
		return r, base.NewGinfraErrorChild("Bundle not accessible", err)
	}
	defer func(a *zip.ReadCloser) {
		_ = a.Close()
	}(a)

	var (
		inr io.ReadCloser
	)
	found := 0
	for _, f := range a.File {
		for idx, i := range items {
			if f.Name == i {
				if inr, err = f.Open(); err != nil {
					return r, base.NewGinfraErrorChild("Could not access file in bundle", err)
				}
				defer func(inr io.ReadCloser) {
					_ = inr.Close()
				}(inr)

				var outb bytes.Buffer
				if _, err = io.Copy(&outb, inr); err != nil {
					return r, base.NewGinfraErrorChild("Could copy data.", err)
				}

				r[idx] = outb.Bytes()
				found++
			}
		}
	}

	if found < len(items) {
		return r, base.NewGinfraError("Not all items found.")
	}
	return r, nil
}

// #####################################################################################################################
// # CONVENIENCE

func BundleExtractServerTLS(root string) (bool, string, string, error) {
	bp := BundleDetect(root)
	if bp == "" {
		return false, "", "", nil
	}

	data, err := BundleExtract(bp, []string{BundleServerCert, BundleServerKey})
	if err != nil {
		return false, "", "", err
	}

	return true, string(data[0]), string(data[1]), nil
}
