/*
Package selenium
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Shared values for services.
*/
package selenium

const CurrentStatusIdle = "IDLE"
const CurrentStatusRunning = "RUNNING"
const CurrentStatusDone = "DONE" // Done is the same as idle, but indicates there is a result available for the last run.
const CurrentStatusError = "ERROR"
