/*
Package poke
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Poke an REST endpoint or the like.
*/
package poke

import (
	"bytes"
	"encoding/json"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"io"
	"net/http"
)

// #####################################################################################################################
// #

func Post(gctx *app.GContext, url, auth string, args []string) (string, error) {

	data := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return "", base.NewGinfraErrorA("Poke POST failed.", base.LM_CODE, 400, base.LM_CAUSE, "ODD number of arguments.")
		}
		data[args[i]] = args[i+1]
	}
	data["auth"] = auth

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", base.NewGinfraErrorChild("Failed to marshal args to JSON.", err)

	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", base.NewGinfraErrorChild("Failed to create POST request.", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", base.NewGinfraErrorChild("failed to execute POST request.", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", base.NewGinfraErrorA("Poke POST failed.", base.LM_CODE, resp.StatusCode, base.LM_CAUSE,
			string(body))
	}

	return string(body), nil
}
