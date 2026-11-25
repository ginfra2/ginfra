/*
Package host
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Various handlers to manage hosted deployments.
*/
package host

import (
	"context"
	dtypes "github.com/docker/docker/api/types"
	dockerc "github.com/docker/docker/client"
	"gitlab.com/ginfra/ginfra/base"
	"time"
)

// #####################################################################################################################
// # VALUES

const HandleContainerWatchInterval = 250
const HandleContainerWatchTrip = 6 // When the server should be good to go.
const HandleContainerWatchMax = 12 // TODO This is polling which kind sucks so revisit this.

const HandleContainerStartInterval = 75
const HandleContainerStartTrip = 2 // When the server should be good to go.
const HandleContainerStartMax = 12

// #####################################################################################################################
// # GENERAL

func InspectContainer(cli *dockerc.Client, cid string) dtypes.ContainerJSON {
	if cjson, err := cli.ContainerInspect(context.Background(), cid); err != nil {
		panic(base.NewGinfraErrorChild("Could not get container information.", err))
	} else {
		return cjson
	}
}

func InspectStartedContainer(cli *dockerc.Client, cid string, network string) (cjson dtypes.ContainerJSON, ip string) {

	time.Sleep(time.Duration(HandleContainerStartTrip*HandleContainerStartInterval) * time.Millisecond)
	for i := HandleContainerStartTrip; i < HandleContainerStartMax; i++ {
		cjson = InspectContainer(cli, cid)
		if cjson.State.Running == true {
			// If it is still running after this time it must be ok.
			ip = cjson.NetworkSettings.Networks[network].IPAddress
			break
		}
		time.Sleep(time.Duration(HandleContainerStartInterval) * time.Millisecond)
	}
	if ip == "" {
		panic(base.NewGinfraError("Container did not start in time."))
	}

	return cjson, ip
}

// #####################################################################################################################
// #

// handleFreshContainerWatch best effort to watch for the container to be running.
func handleFreshContainerWatch(cli *dockerc.Client, cid string) dtypes.ContainerJSON {
	var (
		cjson dtypes.ContainerJSON
		i     int
	)

	// Ugly polling which is why you should really use one of the servers that supports feedback (like echoapi).
	time.Sleep(time.Duration(HandleContainerWatchTrip*HandleContainerWatchInterval) * time.Millisecond)
	for i = HandleContainerWatchTrip; i < HandleContainerWatchMax; i++ {
		cjson = InspectContainer(cli, cid)
		if cjson.State.Running == true {
			// If it is still running after this time it must be ok.
			break
		}
		time.Sleep(time.Duration(HandleContainerWatchInterval) * time.Millisecond)
	}

	if i == HandleContainerWatchMax {
		panic(base.NewGinfraError("Container is not running."))
	}

	return cjson
}

func HandleFreshContainer(cli *dockerc.Client, cid, port, network string) dtypes.ContainerJSON {

	var (
		cjson dtypes.ContainerJSON
	)

	// Get initial inspection
	cjson, _ = InspectStartedContainer(cli, cid, network)
	handleFreshContainerWatch(cli, cid)
	return cjson
}
