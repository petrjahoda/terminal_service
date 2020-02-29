#!/usr/bin/env bash
cd linux
upx terminal_service_linux
cd ..
cd mac
upx terminal_service_mac
cd ..
cd windows
upx terminal_service_windows.exe
cd ..
docker rmi -f petrjahoda/terminal_service:"$1"
docker build -t petrjahoda/terminal_service:"$1" .
docker push petrjahoda/terminal_service:"$1"