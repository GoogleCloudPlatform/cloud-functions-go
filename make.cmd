@echo off

setlocal

set GOBIN=main
set OUT=function.zip

if "%1"=="" goto :BUILD_ALL
if "%1"=="all" goto :BUILD_ALL
if "%1"=="godev" goto :GO_DEV
if "%1"=="clean" goto :CLEAN

echo Invalid make target "%1"
exit /b 2

:BUILD_ALL
    call :GCF_JS
    if %errorlevel% neq 0 exit /b %errorlevel%

    call :GCF_GO
    if %errorlevel% neq 0 exit /b %errorlevel%

    powershell -command "& {& 'Compress-Archive' -DestinationPath %OUT% -Path %GOBIN%,node_modules,index.js,package.json -Force}"

    goto :EOF

:GCF_JS
    npm install --ignore-scripts --save local_modules/execer 2>nul || (
        rem Fallback in case npm is unavailable.
        if exist node_modules\execer (
            set errorlevel=0
            goto :EOF
        )

        if not exist node_modules mkdir node_modules
        mklink /d node_modules\execer ..\local_modules\execer
    )

    goto :EOF

:GCF_GO
    setlocal
    set GOARCH=amd64
    set GOOS=linux
    go build -tags netgo -tags node %GOBIN%.go
    endlocal

    goto :EOF

:GO_DEV
    go run %GOBIN%.go

    goto :EOF

:CLEAN
    rmdir /s /q node_modules
    del /s %GOBIN% %OUT%

    goto :EOF
