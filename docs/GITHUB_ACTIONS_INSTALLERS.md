# GitHub Actions Automated Installer Builds

This document explains how GitHub Actions automatically builds installers for CRGoDICOM across different platforms and branches.

## Overview

The project includes three GitHub Actions workflows that automatically build installers:

1. **Release Build** (`release-build.yml`) - Builds production installers for tagged releases
2. **Development Build** (`develop-build.yml`) - Builds development installers for the develop branch
3. **Feature Build** (`feature-build.yml`) - Builds feature installers for feature branches

## Workflow Triggers

### Release Build Workflow
- **Triggers**: 
  - Push tags matching `v*` (e.g., `v1.0.0`)
  - Manual dispatch via GitHub UI
- **Purpose**: Creates production-ready installers for releases
- **Artifacts**: MSI, DMG, AppImage installers
- **Retention**: 30 days

### Development Build Workflow
- **Triggers**:
  - Push to `develop` branch
  - Pull requests targeting `develop`
  - Manual dispatch
- **Purpose**: Creates development installers for testing
- **Artifacts**: Development installers with "dev" prefix
- **Retention**: 7 days

### Feature Build Workflow
- **Triggers**:
  - Push to feature branches (`feature/*`, `feat/*`, `hotfix/*`)
  - Pull requests targeting `develop` or `main`
  - Manual dispatch
- **Purpose**: Creates feature installers for testing new functionality
- **Artifacts**: Feature installers with "feat" prefix
- **Retention**: 3 days

## Supported Platforms

All workflows build for the following platforms:

| Platform | OS | Architecture | Installer Type |
|----------|----|--------------|--------------|
| Windows | `windows-latest` | x64, ARM64 | MSI, ZIP |
| macOS | `macos-latest` | x64, ARM64 | DMG |
| Linux | `ubuntu-latest` | x64, ARM64 | AppImage, TAR.GZ |

## Installer Features

### Windows Installers
- **MSI Package**: Professional Windows installer with registry integration
- **File Associations**: `.dcm` files associated with CRGoDICOM
- **Start Menu**: Application shortcuts in Start Menu
- **Desktop Shortcut**: Optional desktop shortcut creation
- **Uninstaller**: Complete removal capability
- **Registry Keys**: Installation tracking and configuration

### macOS Installers
- **DMG Package**: Standard macOS disk image installer
- **App Bundle**: Proper macOS application bundle structure
- **Code Signing**: Ready for Apple Developer code signing
- **Notarization**: Ready for Apple notarization process
- **File Associations**: DICOM file type associations
- **Launch Services**: Integration with macOS file handling

### Linux Installers
- **AppImage**: Portable Linux application format
- **Desktop Integration**: Application menu entries
- **MIME Types**: DICOM file type associations
- **System Integration**: Command-line tool installation
- **Package Formats**: Multiple distribution formats

## Workflow Configuration

### Build Matrix
Each workflow uses a matrix strategy to build for multiple platforms simultaneously:

```yaml
strategy:
  matrix:
    include:
      - os: ubuntu-latest
        platform: linux-amd64
        installer_type: appimage
      - os: windows-latest
        platform: windows-amd64
        installer_type: msi
      - os: macos-latest
        platform: darwin-amd64
        installer_type: dmg
```

### Version Handling
- **Release builds**: Use git tag as version (e.g., `v1.0.0`)
- **Development builds**: Use `dev-<commit-hash>` format
- **Feature builds**: Use `feat-<branch-name>-<commit-hash>` format

### Artifact Management
- **Upload**: All installers uploaded as GitHub Actions artifacts
- **Release**: Production installers attached to GitHub releases
- **Retention**: Different retention periods based on build type

## Installation Scripts

Post-installation scripts are included for each platform:

### Windows (`windows-post-install.bat`)
- Adds application to system PATH
- Creates desktop and Start Menu shortcuts
- Sets up file associations for `.dcm` files
- Creates configuration and studies directories
- Updates Windows registry

### macOS (`macos-post-install.sh`)
- Creates command-line symlink in `/usr/local/bin`
- Sets up desktop shortcut
- Configures file associations via Launch Services
- Creates user configuration and studies directories
- Sets proper permissions

### Linux (`linux-post-install.sh`)
- Creates system-wide command-line symlink
- Sets up desktop application entry
- Configures MIME type associations
- Installs application icons
- Creates user configuration and studies directories
- Sets up desktop shortcuts

## Usage Examples

### Creating a Release
1. Create and push a version tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
2. GitHub Actions automatically builds installers
3. Installers are attached to the GitHub release

### Development Testing
1. Push changes to `develop` branch
2. GitHub Actions builds development installers
3. Download installers from Actions artifacts
4. Test the development build

### Feature Testing
1. Create a feature branch:
   ```bash
   git checkout -b feature/new-functionality
   git push origin feature/new-functionality
   ```
2. GitHub Actions builds feature installers
3. Download and test the feature build

## Manual Workflow Dispatch

All workflows support manual triggering:

1. Go to **Actions** tab in GitHub repository
2. Select the desired workflow
3. Click **Run workflow**
4. Choose branch and options
5. Click **Run workflow**

## Artifact Download

### From GitHub Actions
1. Go to **Actions** tab
2. Select the workflow run
3. Scroll to **Artifacts** section
4. Download the desired installer

### From GitHub Releases
1. Go to **Releases** page
2. Select the desired release
3. Download installers from **Assets** section

## Configuration Files

### Installer Configuration (`installer/config.json`)
- Application metadata
- Platform-specific installer settings
- Dependencies and requirements
- File inclusion/exclusion rules
- Post-installation scripts
- Uninstaller configuration

### Workflow Files
- `.github/workflows/release-build.yml` - Release builds
- `.github/workflows/develop-build.yml` - Development builds
- `.github/workflows/feature-build.yml` - Feature builds

## Customization

### Adding New Platforms
1. Add platform to workflow matrix
2. Create platform-specific installer logic
3. Add post-installation script
4. Update documentation

### Modifying Installer Behavior
1. Edit `installer/config.json`
2. Update post-installation scripts
3. Modify workflow installer creation steps
4. Test changes

### Adding New Workflows
1. Create new workflow file in `.github/workflows/`
2. Define triggers and build matrix
3. Implement installer creation logic
4. Configure artifact upload and release

## Troubleshooting

### Common Issues
1. **Build Failures**: Check workflow logs for specific errors
2. **Missing Artifacts**: Verify artifact upload steps completed
3. **Installer Issues**: Test installers on target platforms
4. **Permission Errors**: Check file permissions and ownership

### Debug Mode
Enable debug logging by adding `ACTIONS_STEP_DEBUG: true` to workflow secrets.

### Local Testing
Test installer creation locally using Makefile targets:
```bash
make installer-all    # Create all installers
make installer-windows # Create Windows installer
make installer-macos   # Create macOS installer
make installer-linux   # Create Linux installer
```

## Security Considerations

### Code Signing
- Windows: Configure WiX code signing
- macOS: Set up Apple Developer certificates
- Linux: Consider GPG signing for packages

### Secrets Management
- Store signing certificates in GitHub Secrets
- Use environment variables for sensitive data
- Rotate secrets regularly

### Dependency Security
- Keep build dependencies updated
- Use specific versions for reproducibility
- Scan dependencies for vulnerabilities

## Performance Optimization

### Build Caching
- Cache Go modules between builds
- Cache build dependencies
- Use build matrix for parallel builds

### Artifact Optimization
- Compress large artifacts
- Remove unnecessary files
- Use appropriate retention periods

### Workflow Efficiency
- Minimize build time with targeted builds
- Use conditional steps where possible
- Optimize dependency installation

## Monitoring and Notifications

### Build Status
- GitHub provides build status badges
- Failed builds send notifications
- PR comments show build status

### Release Notifications
- GitHub releases send notifications
- Configure webhooks for external systems
- Use GitHub API for automation

## Future Enhancements

### Planned Features
- [ ] Automated testing of installers
- [ ] Integration with package repositories
- [ ] Automated dependency updates
- [ ] Build performance monitoring
- [ ] Multi-architecture support
- [ ] Container-based builds

### Community Contributions
- [ ] Additional platform support
- [ ] Improved installer UIs
- [ ] Enhanced post-installation scripts
- [ ] Better error handling
- [ ] Documentation improvements

This automated installer system provides a robust foundation for distributing CRGoDICOM across multiple platforms while maintaining consistency and quality across all builds.
