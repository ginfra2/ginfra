@ECHO OFF

IF NOT "%~1"=="" GOTO START
ECHO This script requires three arguments.
ECHO ARG1 - SERVICE ROOT PATH
GOTO EOF

:START
ECHO ginfra_stub local %1 %2 %3 %4 %5 %6 %7 %8 %9
ginfra_stub local %1 %2 %3 %4 %5 %6 %7 %8 %9

:EOF
