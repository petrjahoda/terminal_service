[![developed_using](https://img.shields.io/badge/developed%20using-Jetbrains%20Goland-lightgrey)](https://www.jetbrains.com/go/)
<br/>
![GitHub](https://img.shields.io/github/license/petrjahoda/terminal_service)
[![GitHub last commit](https://img.shields.io/github/last-commit/petrjahoda/terminal_service)](https://github.com/petrjahoda/terminal_service/commits/master)
[![GitHub issues](https://img.shields.io/github/issues/petrjahoda/terminal_service)](https://github.com/petrjahoda/terminal_service/issues)
<br/>
![GitHub language count](https://img.shields.io/github/languages/count/petrjahoda/terminal_service)
![GitHub top language](https://img.shields.io/github/languages/top/petrjahoda/terminal_service)
![GitHub repo size](https://img.shields.io/github/repo-size/petrjahoda/terminal_service)
<br/>
[![Docker Pulls](https://img.shields.io/docker/pulls/petrjahoda/terminal_service)](https://hub.docker.com/r/petrjahoda/terminal_service)
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/petrjahoda/terminal_service?sort=date)](https://hub.docker.com/r/petrjahoda/terminal_service/tags)
<br/>
[![developed_using](https://img.shields.io/badge/database-PostgreSQL-red)](https://www.postgresql.org) [![developed_using](https://img.shields.io/badge/runtime-Docker-red)](https://www.docker.com)

# Terminal Service
## Description
Go service, that creates and updates orders and downtimes for workplaces.

## Installation Information
Install under docker runtime using [this dockerfile image](https://github.com/petrjahoda/system/tree/master/latest) with this command: ```docker-compose up -d```

## Implementation Information
Check the software running with this command: ```docker stats```. <br/>
Terminal_service has to be running.

## Additional information
* working with workplace that has linked devices in  ```device_workplace``` table, device has to of type 'Zapsi Touch' in ```device``` table
* creates and updates orders in ```order_records``` table
* creates and updates downtimes in ```downtime_records``` table


Devices example
![Devices](devices.png)

Device Workplaces example
![DeviceWorkplaces](device_workplaces.png)

Downtime Records example
![DowntimeRecords](downtimes.png)

Order Records example
![OrderRecords](orders.png)


© 2020 Petr Jahoda

