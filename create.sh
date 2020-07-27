#!/usr/bin/env bash
cd linux
upx terminal_service_linux
cd ..

docker rmi -f petrjahoda/terminal_service:latest
docker build -t petrjahoda/terminal_service:latest .
docker push petrjahoda/terminal_service:latest

docker rmi -f petrjahoda/terminal_service:2020.3.1
docker build -t petrjahoda/terminal_service:2020.3.1 .
docker push petrjahoda/terminal_service:2020.3.1
