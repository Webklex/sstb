# Changelog

All notable changes to `webklex/sstb` will be documented in this file.

Updates should follow the [Keep a CHANGELOG](http://keepachangelog.com/) principles.

## [UNRELEASED]
### Fixed
- Daily statistics fixed
- Prevent idle alert spam
- Summary will use the local order backups instead of requesting them from the exchange 

### Added
- Idle alert support added
- Save fulfilled orders as json

### Breaking changes
- Job config parameter `alert.summary` is now an integer array, where each integer represents the 
hour you want to generate and send it.

## [1.0.0] - 2021-03-14
Initial release