# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com), and this
project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]
### Added
* GitHub website - https://matt1484.github.io/bl3_auto_vip/
* Go modules ([#1])
* GitHub Actions CI pipeline ([#6])
* Change log

### Changed
* Improve README

## v2.0.0 - 2019-09-11
### Changed
* Rewrote all code in go to add future mobile support (also more maintainable
and smaller executable)

## v1.2.1 - 2019-08-28
### Fixed
* Fixed bug where tables in comments would count as codes

### Added
* Password masking

## v1.2.0 - 2019-08-27
### Added
* Timer so it does not immediately close when done
* Support for codes with multiple types

### Fixed
* Bad logging around/error handling involving code type setup

## v1.1.0 - 2019-08-25
### Added
* Support for command line args (email and
  password)

### Changed
* Now uses REST endpoints and JSON parsing rather than headless browser
* Utilize .net core 3.0

### Fixed
* Timeout issues and added 

## v1.0.0 - 2019-08-22
* Initial release