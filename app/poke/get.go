/*
Package poke
Copyright (c) 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Poke an REST endpoint or the like.
*/
package poke

import (
	"bytes"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"io"
	"net/http"
)

// #####################################################################################################################
// #

func Get(gctx *app.GContext, url, auth string, args []string) (string, error) {

	data := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		data[args[i]] = args[i+1]
	}
	data["auth"] = auth

	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return "", base.NewGinfraErrorChild("Failed to create GET request.", err)
	}
	req.Header.Set("Content-Type", "application/json")

	for k, v := range data {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", base.NewGinfraErrorChild("failed to execute GET request.", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", base.NewGinfraErrorA("Poke GET failed.", base.LM_CODE, resp.StatusCode, base.LM_CAUSE,
			string(body))
	}

	return string(body), nil
}
