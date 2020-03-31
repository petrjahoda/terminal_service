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

docker rmi -f petrjahoda/terminal_service:latest
docker build -t petrjahoda/terminal_service:latest .
docker push petrjahoda/terminal_service:latest

docker rmi -f petrjahoda/terminal_service:2020.1.3
docker build -t petrjahoda/terminal_service:2020.1.3 .
docker push petrjahoda/terminal_service:2020.1.3
