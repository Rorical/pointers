@echo off
protoc --go_out=. *.proto --experimental_allow_proto3_optional
pause