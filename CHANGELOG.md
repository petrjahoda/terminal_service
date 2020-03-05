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

## [2020.1.2.29] - 2020-02-29

### Added
- when saving order, order_id is set to internal


## [2020.1.2.29] - 2020-02-29

### Changed
- name of database changed to zapsi3
- proper testing for mariadb, postgres and mssql
- added logging for all important methods and functions
- code refactoring for better readability

### Fixed
- proper closing downtimes and orders
- proper updating data for order and downtime