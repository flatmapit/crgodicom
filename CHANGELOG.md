# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **CT Study Generation**: Successfully created CT study with two series for patient CTTEST
- **PACS Integration**: Successfully sent CT study to PACS1 (Orthanc) using DCMTK storescu
- **Study Export**: Generated PDF report for CT study export
- **US Study Generation**: Created 3-series Ultrasound study for patient SMITH^GRANNY
- **Multi-format Export**: Exported US study to PNG (24 images) and PDF report formats
- **PACS Integration**: Successfully sent US study to PACS1 (Orthanc) using DCMTK storescu
- **JPEG Export Support**: Added JPEG export functionality with high-quality compression (95% quality)
- **Enhanced UID Generation**: Implemented cryptographically secure UID generation
- **Metadata Generation**: Added comprehensive DICOM metadata generation system

### Changed
- **Study Cleanup**: Cleaned up old and incomplete studies, keeping only the complete CTTEST study
- **Test Cleanup**: Removed test data files, generated binaries, and temporary files
- **Comprehensive DICOM Metadata Coverage**: Complete implementation of Type 1, Type 2, and Type 3 DICOM attributes
- **Enterprise-Grade UID Management**: Cryptographically secure UID generation with guaranteed uniqueness
- **DICOM Conformance Validation Framework**: Built-in conformance checking for all DICOM modules
- **Extended Modality Support**: Added NM, PT, RT, SR modalities with modality-specific image patterns
- **Enhanced MR Sequences**: T1, T2, FLAIR, DWI sequence patterns for MRI studies
- **Comprehensive Testing Framework**: Unit, integration, conformance, and performance testing
- **Visual Verification**: Burned-in metadata for integration testing and debugging
- **Advanced Image Generation**: Modality-specific patterns (hot spots for NM, metabolic activity for PT, dose distributions for RT)
- **Comprehensive Documentation**: Updated README.md and new TESTING.md with detailed testing procedures
- Initial CLI framework implementation
- Configuration system with YAML support
- Study templates (built-in and user-defined)
- All CLI commands: create, list, send, verify, export, create-template
- Makefile for cross-platform builds
- Basic project structure
- JPEG export capability for DICOM images
- Enhanced DICOM metadata reading and pixel data extraction
- Support for multiple export formats: PNG, JPEG, and PDF
- DCMTK CGO bindings for DICOM operations
- pkg-config integration for DCMTK compilation
- C wrapper functions for DCMTK reader, writer, and network operations

### Changed
- **Enhanced DICOM Generation**: Upgraded from basic metadata to comprehensive DICOM 3.0 compliance
- **Improved UID Generation**: Replaced simple timestamp-based UIDs with enterprise-grade cryptographic generation
- **Extended Modality Support**: Expanded from 6 to 10 supported modalities with specialized patterns
- **Enhanced CLI**: Added conformance checking flags and comprehensive validation options
- Migrated from suyashkumar/dicom library to DCMTK for all DICOM operations
- Updated DICOM reader to use DCMTK CGO bindings
- Updated DICOM writer to use DCMTK CGO bindings
- Updated PACS network operations to use DCMTK CGO bindings

### Fixed
- Resolved CGO compilation issues with DCMTK headers
- Fixed C++ compilation requirements for DCMTK integration
- Eliminated duplicate symbol errors in CGO bindings

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
