@echo off
setlocal enabledelayedexpansion

set OUTPUT_DIR=.
set APP_NAME=WM
set PLATFORMS=windows/amd64 linux/amd64

rem 默认编译所有平台

where upx >nul 2>nul
if %errorlevel% neq 0 (
    echo UPX not found. Skipping compression.
    set USE_UPX=false
) else (
    set USE_UPX=true
)

echo Downloading dependencies...
call go mod tidy

for %%P in (%PLATFORMS%) do (
    for /f "tokens=1,2 delims=/" %%a in ("%%P") do (
        set CURRENT_GOOS=%%a
        set CURRENT_GOARCH=%%b
        if "!CURRENT_GOOS!"=="windows" (
            set OUTPUT_FILENAME=wm.exe
        ) else (
            set OUTPUT_FILENAME=wm
        )
        
        echo Building !CURRENT_GOOS!/!CURRENT_GOARCH!...
        set GOOS=!CURRENT_GOOS!
        set GOARCH=!CURRENT_GOARCH!
        
        go build -o %OUTPUT_DIR%\!OUTPUT_FILENAME! cmd\server\main.go
        
        if !errorlevel! neq 0 (
            echo Build failed!
            exit /b 1
        )
        
        if "!USE_UPX!"=="true" (
            echo Compressing via UPX...
            upx %OUTPUT_DIR%\!OUTPUT_FILENAME!
        )
    )
)

echo Build complete! Artifacts in project root
endlocal
