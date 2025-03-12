@echo off
del mac-spoofer.exe >nul 2>&1
go install github.com/akavel/rsrc@latest
rsrc -manifest main.manifest -o rsrc.syso -arch amd64 -ico icon.ico
go build -ldflags="-H windowsgui"
del rsrc.syso 