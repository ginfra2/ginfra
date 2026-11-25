/*
Package base
Copyright (c) 2023, 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Base names for logging and whatnot.
*/
package base

const (

	// Log configuration constants

	LMCONFIG_LOGDIR = "log"

	// Catalog of log and metric labels.

	LM_ERROR       = "error"
	LM_CAUSE       = "cause"
	LM_CAUSE_OTHER = "cause.other"
	LM_MARKER      = "marker"
	LM_TARGET      = "target"
	LM_TYPE        = "type"
	LM_NAME        = "name"
	LM_TIME        = "time"
	LM_PATH        = "path"
	LM_CODE        = "code"
	LM_TEXT        = "text"
	LM_ID          = "id"
	LM_ID_RUN      = "id.run"
	LM_UID         = "uid"
	LM_EVENT       = "event"

	LM_CONFIG_NAME = "config.name"
	LM_CONFIG_PATH = "config.path"

	LM_FILEPATH   = "filepath"
	LM_WORKINGDIR = "wd"

	LM_IMAGE_PATH = "image.path"

	LM_SERVICE_NAME  = "service.name"
	LM_SERVICE_PORT  = "port"
	LM_SERVICE_IP    = "ip"
	LM_SERVICE_CLASS = "service.class"
	LM_URL           = "url"
	LM_SERVICE_TYPE  = "service.type"

	LM_WORKER_ID = "worker.id"

	LM_NETWORK_ID   = "network.id"
	LM_NETWORK_NAME = "network.name"
	LM_HOSTNAME     = "hostname"

	LM_DEPLOYMENT_NAME = "deployment.name"

	LM_COMMAND = "command"
	LM_STDIN   = "stdin"
	LM_STDOUT  = "stdout"
	LM_STDERR  = "stderr"
	LM_OUTPUT  = "output"
	LM_SCAN    = "scan"
	LM_PID     = "pid"
	LM_CID     = "cid"

	LM_VALUE = "value"
	LM_NUMBER = "number"

	LM_TOKEN = "token"
	LM_KEY   = "key"

	LM_PACKAGE = "package"

	LM_EXPECTED = "expected"
	LM_ACTUAL   = "actual"

	LM_STATUS = "status"
	LM_RESULT = "result"

	LM_SPECIFICS = "specifics"

	LM_LOG = "log"

	LM_TIMEOUT = "timeout"
	
	LM_LINENUMBER  = "line.number"
)
