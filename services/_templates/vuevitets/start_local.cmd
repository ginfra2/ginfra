@ECHO OFF

SET DEPLOYMENT=default
SET SNAME=<<<ServiceName>>>
SET PORT=8080

IF NOT "%~1"=="" GOTO START
ECHO This script requires four arguments.
ECHO ARG1 - SERVICE ROOT PATH
GOTO EOF

:START

REM Hax the json file.
FOR /F "tokens=*" %%g IN ('Powershell -Nop -C "(Get-Content .\service_config.json|ConvertFrom-Json).deployment"') do (SET DEPLOYMENT=%%g)
FOR /F "tokens=*" %%g IN ('Powershell -Nop -C "(Get-Content .\service_config.json|ConvertFrom-Json).name"') do (SET SNAME=%%g)
FOR /F "tokens=*" %%g IN ('Powershell -Nop -C "(Get-Content .\service_config.json|ConvertFrom-Json).port"') do (SET PORT=%%g)
FOR /F "tokens=*" %%g IN ('Powershell -Nop -C "(Get-Content .\service_config.json|ConvertFrom-Json).port"') do (SET CURI=%%g)
FOR /F "tokens=*" %%g IN ('Powershell -Nop -C "(Get-Content .\service_config.json|ConvertFrom-Json).port"') do (SET CTOKEN=%%g)

REM Setup .env
REM Use the primary host URI.  If there is no primary host, it will still be 'localhost'
ECHO VITE_GINFRA_CONFIG_URI=%CURI%> .env
ECHO VITE_GINFRA_ACCESS_TOKEN=%CTOKEN% >> .env
ECHO VITE_GINFRA_DEPLOYMENT=%DEPLOYMENT% >> .env
ECHO VITE_GINFRA_NAME=%SNAME% >> .env
ECHO VITE_GINFRA_PORT=%PORT% >> .env

REM Run it
npm run dev

:EOF
