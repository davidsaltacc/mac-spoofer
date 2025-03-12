@echo off
go install github.com/akavel/rsrc@latest
rsrc -manifest main.manifest -o rsrc.syso -arch amd64 -ico icon.ico
go build -ldflags="-H windowsgui"
rm rsrc.syso