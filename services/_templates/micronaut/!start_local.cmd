@ECHO OFF

REM This file will not be updated, so any additions or changes are safe.

IF NOT "%~1"=="" GOTO START
ECHO This script requires one argument.
ECHO ARG1 - SERVICE ROOT PATH
GOTO EOF

REM Set env
java -jar assets/ginfraapp-1.0.0.jar config props

:START
if not exist "ENV" GOTO STARTJAVA
    FOR /F "tokens=*" %%i in ('type Settings.txt') do SET %%i

:STARTJAVA
java -javaagent:assets/opentelemetry-javaagent.jar -jar build/libs/<<<ServiceName>>>-0.1-all.jar %1 %2 %3 %4 %5 %6 %7 %8 %9

:EOF
