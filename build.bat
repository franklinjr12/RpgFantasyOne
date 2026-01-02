@echo off
setlocal

set "PROJECT_ROOT=%~dp0"
set "OUTPUT_DIR=%PROJECT_ROOT%output"
set "OUTPUT_EXE=%OUTPUT_DIR%\app.exe"

echo Building SingleFantasy...

if not exist "%OUTPUT_DIR%" (
    mkdir "%OUTPUT_DIR%"
    echo Created output directory: %OUTPUT_DIR%
)

cd /d "%PROJECT_ROOT%"

go build -o "%OUTPUT_EXE%" ./app

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Executable: %OUTPUT_EXE%
) else (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

endlocal

