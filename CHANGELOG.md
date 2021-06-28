# Alarm Service Changelog

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/).

Please note, that this project, while following numbering syntax, it DOES NOT
adhere to [Semantic Versioning](http://semver.org/spec/v2.0.0.html) rules.

## Types of changes

* ```Added``` for new features.
* ```Changed``` for changes in existing functionality.
* ```Deprecated``` for soon-to-be removed features.
* ```Removed``` for now removed features.
* ```Fixed``` for any bug fixes.
* ```Security``` in case of vulnerabilities.

## [2021.2.3.28] - 2021-06-28

### Added
- when creating downtime, adding userId and orderId, if valid

## [2021.2.3.14] - 2021-06-14

### Added
- copyright
- updated libraries

## [2021.2.3.5] - 2021-06-05

### Changed

- proper datetime when closing and opening order, downtime and user records

## [2021.2.2.13] - 2021-05-13

### Changed

- updated to latest go 1.16.4
- updated to latest go libraries

## [2021.2.2.3] - 2021-05-03

### Changed
- updated to latest go
- updated to latest go libraries

## [2021.2.1.8] - 2021-04-08

### Added
- application name to sql connection string

## [2021.1.3.30] - 2021-03-30

### Changed
- updated to latest go
- updated to latest libraries

### Fixed
- proper loading shift when saving order record

## [2021.1.3.18] - 2021-03-18

### Changed
- updated to latest go
- updated to latest libraries

## [2021.1.2.22] - 2021-02-22

### Changed
- opening one database connection for workplace loop

## [2021.1.2.21] - 2021-02-21

### Changed
- updated to latest go
- updated to latest libraries

## [2021.1.2.20] - 2021-01-20

### Changed
- reduced database calls to minimum

## [2021.1.1.11] - 2021-01-11

### Fixed
- saving proper workplace_id when saving user_record

## [2020.4.3.14] - 2020-12-14

### Changed
- updated to latest go
- updated to latest libraries

## [2020.4.2.17] - 2020-11-17

### Changed
- updated all go libraries 
- updated Dockerfile
- updated create.sh

## [2020.4.1.26] - 2020-10-26

### Fixed
- fixed leaking goroutine bug when opening sql connections, the right way is this way

## [2020.4.1.3] - 2020-10-03

### Changed
- updated readme.md

## [2020.3.2.22] - 2020-08-29

### Changed
- functions naming changed to idiomatic go (ThisFunction -> thisFunction)

## [2020.3.2.22] - 2020-08-22

### Added
- automatic go get -u all when creating docker image

## [2020.3.2.13] - 2020-08-13

### Changed
- updated to latest libraries
- updated to go 1.15 -> final program size reduced to 90%

## [2020.3.2.5] - 2020-08-05

### Added
- MIT license

### Changed
- code moved to more files
- logging updated
- name of functions updated

## [2020.3.2.4] - 2020-08-04

### Fixed
- proper creating new order and downtime and user records
- proper checkign for open order and downtime records

### Changed
- updated to latest libraries


## [2020.3.1.30] - 2020-07-30

### Fixed
- proper behavior with new database structure
- proper closing database connections with sqlDB, err := db.DB() and defer sqlDB.Close()

### Changed
- added tzdata to docker image

## [2020.3.1.27] - 2020-07-27

### Changed
- changed to gorm v2
- postgres only

### Removed
- logging to file
- all about config file

## [2020.2.2.18] - 2020-05-18

### Added
- init for actual service directory
- db.logmode(false)

## [2020.1.3.31] - 2020-03-31

### Added
- updated create.sh for uploading proper docker version automatically

## [2020.1.2.29] - 2020-02-29

### Added
- when saving order, order_id is set to internal
- when saving downtime, workplace_id added 

#### Fixed
- when saving downtime_records, adding user only when record in user_records


## [2020.1.2.29] - 2020-02-29

### Changed
- name of database changed to zapsi3
- proper testing for mariadb, postgres and mssql
- added logging for all important methods and functions
- code refactoring for better readability

### Fixed
- proper closing downtimes and orders
- proper updating data for order and downtime