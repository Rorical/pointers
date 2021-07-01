@echo off
protoc --go_out=. *.proto
pause