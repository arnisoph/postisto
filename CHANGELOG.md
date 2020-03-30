# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Calendar Versioning](https://calver.org/) (YYYY.MM.DD-<SHORT-COMMIT-HASH>).

## [Unreleased]
### Added
- Added initial changelog

## [v2020.03.30-5625bf2] - 2020-03-30
### Added
- Added onetime-filtering CLI flag

### Changed
- Improved IMAP reconnection logic. Thanks again to @hikhvar 
- Updated Golang to 1.14
- Updated Golang dependencies
- Increased test coverage

## [v2020.02.08-6723c60] - 2020-02-08
### Added
- Added initial proper README file

### Changed
- Changed Semantic to automated Calendar Versioning
- Updated Golang dependencies
- Changed tests to use github.com/testcontainers/testcontainers-go for integration testing

## [v0.2.0] - 2020-01-20
### Added
- Added more logging
- Added re-connect mechanism if connection or user session is lost

### Changed
- Improved build, test & release process
- Improved code quality
- Improved charset handling on message parsing

## [v0.1.0] - 2020-01-12
### Added
- This is the initial release!

[Unreleased]: https://github.com/arnisoph/postisto/compare/v2020.03.30-5625bf2...HEAD
[v2020.03.30-5625bf2]: https://github.com/arnisoph/postisto/compare/v2020.02.08-6723c60...v2020.03.30-5625bf2
[v2020.02.08-6723c60]: https://github.com/arnisoph/postisto/compare/v0.2.0...v2020.02.08-6723c60
[v0.2.0]: https://github.com/arnisoph/postisto/compare/v0.1.0...v0.2.0 
[v0.1.0]: https://github.com/arnisoph/postisto/tree/v0.1.0