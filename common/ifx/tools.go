/*
Package ifx
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Interface tools.
*/
package ifx

import (
	"context"
	control "gitlab.com/ginfra/ginfra/api/service_control/client"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"net/http"
	"time"
)

var (
	plc      *control.ClientWithResponses
	pretries = app.ServiceConnectRetries
)

func pingServiceClient(c *control.Client) error {
	hc := &http.Client{
		Timeout: time.Duration(app.ServiceConnectRetryTime) * time.Millisecond,
	}
	c.Client = hc
	return nil
}

// PingService attempts to communicate with a service by sending requests to the specified host, port, and service.
// It retries a preconfigured number of times if the service is unavailable and returns a response string or an error.
func PingService(host, port, service, name string) (string, error) {
	var err error

	if plc == nil {
		plc, err = control.NewClientWithResponses("http://"+host+":"+port+"/", pingServiceClient)
		if err != nil {
			return "", base.NewGinfraErrorA(name+" can not be contacted because client could not be created.", base.LM_CAUSE,
				err)
		}
	}
	defer func() { plc = nil }() // These should be on demand unless a test is posting it.

	var gr *control.GiGetSpecificsResponse

	for i := 0; i < pretries; i++ {

		gr, err = plc.GiGetSpecificsWithResponse(context.Background())
		if gr != nil && gr.StatusCode() == http.StatusOK {
			break
		}
		if i >= pretries-1 {
			// Tries are up.  Real problem.
			if gr != nil {
				return "", base.NewGinfraErrorA(name+" not available.", base.LM_CAUSE, string(gr.Body), base.LM_CODE,
					gr.StatusCode())
			} else {
				return "", base.NewGinfraErrorA(name+" not available.", base.LM_CAUSE, err)
			}

		} else {
			time.Sleep(time.Duration(app.ServiceConnectRetryTime*i) * time.Millisecond)
		}

	}

	r := string(gr.Body)
	if gr.JSON200 != nil {
		r = *gr.JSON200
	}
	return r, nil
}
