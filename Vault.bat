@echo off
setlocal EnableExtensions
cd /d "%~dp0"
set "ROOT=%cd%"
title Vault
echo.
echo Vault
echo Double-click this file to install what is missing, start the site, and import a collection.
echo Put kitchen-collection.zip in this folder before the first launch.
echo.

if /i "%~1"=="stop" goto stop

set "REIMPORT="
set "ARG1=%~1"
if /i "%ARG1%"=="--reimport" (
  set "REIMPORT=1"
  set "ARG1=%~2"
)

call :already_running
if not errorlevel 1 goto already

call :ensure_source
if errorlevel 1 goto failed
call :ensure_go
if errorlevel 1 goto failed
call :build
if errorlevel 1 goto failed
call :find_archive
if errorlevel 1 goto failed
call :launch
if errorlevel 1 goto failed
goto ready

:already
echo Vault is already running.
echo Run "Vault.bat stop" before importing a new collection.
goto ready

:already_running
netstat -ano | findstr /R /C:":8787 .*LISTENING" >nul
exit /b %errorlevel%

:ensure_source
if exist "%ROOT%\go.mod" (
  set "APP=%ROOT%"
  exit /b 0
)
if exist "%ROOT%\src\mtg-main\go.mod" (
  set "APP=%ROOT%\src\mtg-main"
  exit /b 0
)
echo Downloading Vault source...
if not exist "%ROOT%\src" mkdir "%ROOT%\src"
"%SystemRoot%\System32\curl.exe" -fL --retry 3 -o "%ROOT%\src\vault-source.zip" "https://github.com/MelQ29/mtg/archive/refs/heads/main.zip"
if errorlevel 1 exit /b 1
"%SystemRoot%\System32\tar.exe" -xf "%ROOT%\src\vault-source.zip" -C "%ROOT%\src"
if errorlevel 1 exit /b 1
if not exist "%ROOT%\src\mtg-main\go.mod" exit /b 1
set "APP=%ROOT%\src\mtg-main"
exit /b 0

:ensure_go
set "GOARCH_NAME=amd64"
if /i "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "GOARCH_NAME=arm64"
if /i "%PROCESSOR_ARCHITEW6432%"=="ARM64" set "GOARCH_NAME=arm64"
if /i "%PROCESSOR_ARCHITECTURE%"=="x86" if "%PROCESSOR_ARCHITEW6432%"=="" (
  echo Vault needs 64-bit Windows.
  exit /b 1
)
set "GOROOT=%ROOT%\.tools\go"
if exist "%GOROOT%\bin\go.exe" exit /b 0
echo Downloading Go 1.25.0 for %GOARCH_NAME%...
if not exist "%ROOT%\.tools" mkdir "%ROOT%\.tools"
set "GOZIP=%ROOT%\.tools\go.zip"
"%SystemRoot%\System32\curl.exe" -fL --retry 3 -o "%GOZIP%" "https://go.dev/dl/go1.25.0.windows-%GOARCH_NAME%.zip"
if errorlevel 1 exit /b 1
if exist "%GOROOT%" rmdir /s /q "%GOROOT%"
"%SystemRoot%\System32\tar.exe" -xf "%GOZIP%" -C "%ROOT%\.tools"
if errorlevel 1 exit /b 1
if not exist "%GOROOT%\bin\go.exe" exit /b 1
exit /b 0

:build
echo Building Vault...
if not exist "%ROOT%\bin" mkdir "%ROOT%\bin"
set "PATH=%GOROOT%\bin;%PATH%"
set "GOROOT=%GOROOT%"
set "CGO_ENABLED=0"
set "GOTOOLCHAIN=local"
set "GOTELEMETRY=off"
pushd "%APP%"
"%GOROOT%\bin\go.exe" build -o "%ROOT%\bin\vault.exe" ./cmd/kitchen
set "BUILD_ERR=%errorlevel%"
popd
exit /b %BUILD_ERR%

:find_archive
set "ARCHIVE="
if defined ARG1 (
  if not exist "%ARG1%" (
    echo File not found: %ARG1%
    exit /b 1
  )
  set "ARCHIVE=%ARG1%"
)
if not defined ARCHIVE if exist "%ROOT%\kitchen-collection.zip" set "ARCHIVE=%ROOT%\kitchen-collection.zip"
if not defined ARCHIVE if exist "%ROOT%\collection.zip" set "ARCHIVE=%ROOT%\collection.zip"
if not defined ARCHIVE for %%F in ("%ROOT%\*.kitchen.zip") do (
  if not defined ARCHIVE set "ARCHIVE=%%~fF"
)
if defined REIMPORT if not defined ARCHIVE (
  echo --reimport needs the zip path. Example: Vault.bat --reimport kitchen-collection.zip
  exit /b 1
)
exit /b 0

:launch
if not exist "%ROOT%\logs" mkdir "%ROOT%\logs"
if not exist "%ROOT%\data" mkdir "%ROOT%\data"
if defined ARCHIVE echo Importing %ARCHIVE%
if not defined ARCHIVE echo No collection zip found. The site will open empty. Put kitchen-collection.zip in this folder and run Vault.bat again.
> "%ROOT%\logs\run.cmd" echo @echo off
>> "%ROOT%\logs\run.cmd" echo cd /d "%APP%"
if defined REIMPORT (
  >> "%ROOT%\logs\run.cmd" echo "%ROOT%\bin\vault.exe" -db "%ROOT%\data\kitchen.db" -images "%ROOT%\data\images" -web "%APP%\web" -archive "%ARCHIVE%" -reimport
) else if defined ARCHIVE (
  >> "%ROOT%\logs\run.cmd" echo "%ROOT%\bin\vault.exe" -db "%ROOT%\data\kitchen.db" -images "%ROOT%\data\images" -web "%APP%\web" -archive "%ARCHIVE%"
) else (
  >> "%ROOT%\logs\run.cmd" echo "%ROOT%\bin\vault.exe" -db "%ROOT%\data\kitchen.db" -images "%ROOT%\data\images" -web "%APP%\web"
)
>> "%ROOT%\logs\run.cmd" echo if errorlevel 1 pause
start "Vault" /MIN cmd /c "%ROOT%\logs\run.cmd"
echo Waiting for the site. A large collection can take a few minutes.
set /a N=0
:wait_loop
set /a N+=1
if %N% GTR 300 (
  echo Vault did not start. Open the minimized Vault window for the error.
  exit /b 1
)
if %N% GEQ 5 (
  tasklist /FI "IMAGENAME eq vault.exe" | find /I "vault.exe" >nul
  if errorlevel 1 (
    echo Vault stopped. Open the minimized Vault window for the error.
    exit /b 1
  )
)
"%SystemRoot%\System32\curl.exe" -fsS -o nul http://127.0.0.1:8787/ >nul 2>&1 && exit /b 0
timeout /t 1 /nobreak >nul
goto wait_loop

:ready
start "" "http://127.0.0.1:8787/"
echo.
echo Vault is open at http://127.0.0.1:8787/
echo You can close this window. The site keeps running.
echo To stop it, double-click Vault.bat stop
echo.
pause
exit /b 0

:stop
echo Stopping Vault...
taskkill /IM vault.exe /F >nul 2>&1
echo Stopped.
pause
exit /b 0

:failed
echo.
echo Vault could not start.
pause
exit /b 1
