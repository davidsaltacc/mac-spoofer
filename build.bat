@echo off
go install github.com/akavel/rsrc@latest
rsrc -manifest main.manifest -o rsrc.syso
go build -ldflags="-H windowsgui"