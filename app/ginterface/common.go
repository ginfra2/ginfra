/*
Package ginterfaces
Copyright (c) 2023, 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Common code.
*/
package ginterfaces

import (
	"errors"
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"path/filepath"
	"strings"
)

func checkApiDir(target string) error {

	apid := filepath.Join(target, "api")
	f, err := os.Stat(apid)
	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(apid, os.ModePerm)
	} else if err == nil && f != nil {
		if !f.IsDir() {
			err = base.NewGinfraErrorA("File exists where the api/ directory should be.", base.LM_FILEPATH,
				apid)
		}
	}
	return err
}

func seekRef(object interface{}, rSet map[string]struct{}) {

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
					rSet[rStr[:strings.Index(rStr, "#")]] = struct{}{}
				}

			} else if mapObj, ok := value.(map[string]interface{}); ok {
				seekRef(mapObj, rSet)

			} else if arrObj, ok := value.([]interface{}); ok {
				for _, item := range arrObj {
					seekRef(item, rSet)
				}
			}
		}

	}
}
