@echo off
REM Windows Post-Installation Script for CRGoDICOM
REM This script runs after installation to set up the application

echo Setting up CRGoDICOM...

REM Add to PATH if not already present
set "INSTALL_DIR=%~dp0"
set "PATH_TO_ADD=%INSTALL_DIR%"

REM Check if CRGoDICOM is already in PATH
echo %PATH% | find /i "%PATH_TO_ADD%" >nul
if %errorlevel% neq 0 (
    echo Adding CRGoDICOM to system PATH...
    setx PATH "%PATH%;%PATH_TO_ADD%" /M
    echo CRGoDICOM added to system PATH
) else (
    echo CRGoDICOM is already in system PATH
)

REM Create desktop shortcut
set "DESKTOP=%USERPROFILE%\Desktop"
set "START_MENU=%APPDATA%\Microsoft\Windows\Start Menu\Programs"

echo Creating shortcuts...

REM Desktop shortcut
echo [InternetShortcut] > "%DESKTOP%\CRGoDICOM.url"
echo URL=file:///%INSTALL_DIR%crgodicom.exe >> "%DESKTOP%\CRGoDICOM.url"
echo IconFile=%INSTALL_DIR%crgodicom.exe >> "%DESKTOP%\CRGoDICOM.url"
echo IconIndex=0 >> "%DESKTOP%\CRGoDICOM.url"

REM Start Menu shortcut
if not exist "%START_MENU%\CRGoDICOM" mkdir "%START_MENU%\CRGoDICOM"
echo [InternetShortcut] > "%START_MENU%\CRGoDICOM\CRGoDICOM.url"
echo URL=file:///%INSTALL_DIR%crgodicom.exe >> "%START_MENU%\CRGoDICOM\CRGoDICOM.url"
echo IconFile=%INSTALL_DIR%crgodicom.exe >> "%START_MENU%\CRGoDICOM\CRGoDICOM.url"
echo IconIndex=0 >> "%START_MENU%\CRGoDICOM\CRGoDICOM.url"

REM Create file association for .dcm files
echo Setting up file associations...
reg add "HKEY_CLASSES_ROOT\.dcm" /ve /d "CRGoDICOM.Document" /f >nul 2>&1
reg add "HKEY_CLASSES_ROOT\CRGoDICOM.Document" /ve /d "DICOM Medical Image" /f >nul 2>&1
reg add "HKEY_CLASSES_ROOT\CRGoDICOM.Document\DefaultIcon" /ve /d "%INSTALL_DIR%crgodicom.exe,0" /f >nul 2>&1
reg add "HKEY_CLASSES_ROOT\CRGoDICOM.Document\shell\open\command" /ve /d "\"%INSTALL_DIR%crgodicom.exe\" \"%%1\"" /f >nul 2>&1

REM Create default configuration directory
set "CONFIG_DIR=%APPDATA%\CRGoDICOM"
if not exist "%CONFIG_DIR%" mkdir "%CONFIG_DIR%"

REM Copy default configuration if it doesn't exist
if not exist "%CONFIG_DIR%\crgodicom.yaml" (
    copy "%INSTALL_DIR%crgodicom.yaml" "%CONFIG_DIR%\" >nul 2>&1
    echo Default configuration copied to %CONFIG_DIR%
)

REM Create studies directory
set "STUDIES_DIR=%USERPROFILE%\Documents\CRGoDICOM\studies"
if not exist "%STUDIES_DIR%" (
    mkdir "%STUDIES_DIR%" >nul 2>&1
    echo Studies directory created at %STUDIES_DIR%
)

echo.
echo CRGoDICOM installation completed successfully!
echo.
echo Installation directory: %INSTALL_DIR%
echo Configuration directory: %CONFIG_DIR%
echo Studies directory: %STUDIES_DIR%
echo.
echo You can now run CRGoDICOM from the command line or use the shortcuts.
echo.
pause
