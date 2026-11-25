@ECHO OFF

REM  This file will not be updated and can be changed.

REM This should be run from the repo root.

if not exist service_config.json (
    copy /Y build\service_config.json.NEW .\service_config.json
)
