//go:build unit_test

/*
Package manage
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Testing tools specific to management.
*/
package manage

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
)

func Test_dataToTar(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		data        []byte
		wantErr     bool
		checkErrors func(error) bool
	}{
		{
			name:     "normal case",
			fileName: "goFile.go",
			data:     []byte("Hello"),
			wantErr:  false,
		},
		{
			name:     "empty data",
			fileName: "goFile.go",
			data:     []byte(""),
			wantErr:  false,
		},
		{
			name:     "bad filename",
			fileName: "/",
			data:     []byte(""),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := DataToTar(tt.fileName, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("DataToTar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkErrors != nil && !tt.checkErrors(err) {
				t.Errorf("DataToTar() error = %v, wanted a different error", err)
			}

			// If we want to check the output of the DataToTar function, we can do it here.
			// In this case we check the tar output only in the normal case
			if !tt.wantErr && tt.name == "normal case" {
				br := bytes.NewReader([]byte(got))
				tr := tar.NewReader(br)

				hdr, err := tr.Next()
				if err != nil {
					t.Errorf("error reading tar header: %v", err)
					return
				}

				data := make([]byte, hdr.Size)
				_, err = io.ReadFull(tr, data)
				if err != nil {
					t.Errorf("error reading tar data: %v", err)
					return
				}

				if hdr.Name != tt.fileName || string(data) != string(tt.data) {
					t.Errorf("DataToTar() = %v, wanted %v", got, tt)
					return
				}
			}
		})
	}
}

func Test_stringToTar(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		data        string
		wantErr     bool
		checkErrors func(error) bool
	}{
		{
			name:     "normal case",
			fileName: "goFile.go",
			data:     "Hello",
			wantErr:  false,
		},
		{
			name:     "empty data",
			fileName: "goFile.go",
			data:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := StringToTar(tt.fileName, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("DataToTar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.checkErrors != nil && !tt.checkErrors(err) {
				t.Errorf("DataToTar() error = %v, wanted a different error", err)
			}

			// If we want to check the output of the DataToTar function, we can do it here.
			// In this case we check the tar output only in the normal case
			if !tt.wantErr && tt.name == "normal case" {
				br := bytes.NewReader([]byte(got))
				tr := tar.NewReader(br)

				hdr, err := tr.Next()
				if err != nil {
					t.Errorf("error reading tar header: %v", err)
					return
				}

				data := make([]byte, hdr.Size)
				_, err = io.ReadFull(tr, data)
				if err != nil {
					t.Errorf("error reading tar data: %v", err)
					return
				}

				if hdr.Name != tt.fileName || string(data) != string(tt.data) {
					t.Errorf("DataToTar() = %v, wanted %v", got, tt)
					return
				}
			}
		})
	}
}
