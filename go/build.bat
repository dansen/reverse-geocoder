@echo off
REM Build script for reverse-geocoder-go
SETLOCAL ENABLEDELAYEDEXPANSION

REM Determine script directory
set SCRIPT_DIR=%~dp0
pushd %SCRIPT_DIR%

ECHO === 1. Go env check ===
go version || (echo Go not installed & exit /b 1)

ECHO === 2. Tidy modules ===
go mod tidy || (echo go mod tidy failed & exit /b 1)

ECHO === 3. Run tests ===
FOR /F "tokens=*" %%G IN ('go test ./...') DO (
  echo %%G
)
IF ERRORLEVEL 1 (
  echo Tests failed.
  exit /b 1
)

ECHO === 4. Build CLI ===
cd cmd\rgeocoder || (echo missing cmd\rgeocoder & exit /b 1)
go build -o ..\..\rgeocoder.exe || (echo build failed & exit /b 1)
cd ..\..

ECHO === Done ===
popd
ENDLOCAL
