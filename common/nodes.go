/*
Package common
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Node information.
*/

package common

// #####################################################################################################################
// # VALUEs

const TYPICAL_SERVICE_PATH = "/"

// #####################################################################################################################
// # KNOWN SERVICE TYPES.

type ServiceType string

const (
	SRVTYPE__UNKNOWN   ServiceType = "unknown"   // Not known but still manageable if containers build properly and InvocationCmd is set.
	SRVTYPE__APIECHO   ServiceType = "apiecho"   // API service implemented with GO, openapi and echo.
	SRVTYPE__NODEJS    ServiceType = "nodejs"    // NodeJS server
	SRVTYPE__SERVER    ServiceType = "server"    // A server platform where typically 3rd party servers run.
	SRVTYPE__VUEVITETS ServiceType = "vuevitets" // A Vue.js/Vite/Typescript app.
)
