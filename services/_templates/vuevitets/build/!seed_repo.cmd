@ECHO OFF

REM This file will not be updated and can be changed.

REM This should be run from the repo root.

if not exist postgres_url (
  echo "postgres://postgres:YOURPASSWORD@192.168.1.220:5432"> postgres_url
)

if not exist service_config.json (
    copy /Y build\service_config.json.NEW .\service_config.json
)
