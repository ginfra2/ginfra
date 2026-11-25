@ECHO OFF

REM This file will not be updated, so any additions or changes are safe.

@ECHO OFF

IF NOT "%~1"=="" GOTO START
ECHO This script requires one argument.
ECHO ARG1 - SERVICE ROOT PATH
GOTO EOF

:START

bin\start.cmd $1

ECHO build\ginfra_service.exe %1 %2 %3 %4 %5 %6 %7 %8 %9
build\ginfra_service.exe %1 %2 %3 %4 %5 %6 %7 %8 %9

:EOF
