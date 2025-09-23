# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial CLI framework implementation
- Configuration system with YAML support
- Study templates (built-in and user-defined)
- All CLI commands: create, list, send, verify, export, create-template
- Makefile for cross-platform builds
- Basic project structure
- JPEG export capability for DICOM images
- Enhanced DICOM metadata reading and pixel data extraction
- Support for multiple export formats: PNG, JPEG, and PDF

### Changed
- N/A

### Deprecated
- N/A

### Removed
- N/A

### Fixed
- DICOM parsing issues that prevented pixel data extraction
- Export system now properly reads DICOM metadata and pixel data from files

### Security
- N/A

## [0.1.0] - 2025-01-17

### Added
- Initial project setup
- Git repository initialization
- Basic specification documentation
