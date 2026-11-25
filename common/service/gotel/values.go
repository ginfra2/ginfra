/*
Package gotel
Copyright (c) 2024 Erich Gatejen
[LICENSE]

OpenTelemetry.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package gotel

// ------------------------------------------------------------------------------------------------
// URL RELATED

const OtelUrlProtocol = "url.protocol"
const OtelUrl = "url"

// ------------------------------------------------------------------------------------------------
// NETWORK RELATED

const OtelHostLocal = "net.host.local"
const OtelHostPeer = "net.host.peer"
const OtelPortLocal = "net.port.local"
const OtelPortPeer = "net.port.peer"

// ------------------------------------------------------------------------------------------------
// HTTP/MINE RELATED

const OtelHttpMethod = "http.method"
const OtelHttpStatus = "http.status"
const OtelHttpUserAgent = "http.useragent"

// ------------------------------------------------------------------------------------------------
// FILE RELATED

const OtelTraceFileSuffix = "_trace.log"
const OtelMetricsFileSuffix = "_metrics.log"
const OtelLogsFileSuffix = "_logs.log"
