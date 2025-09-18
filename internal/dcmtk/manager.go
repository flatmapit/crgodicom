package dcmtk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Manager handles DCMTK installation and availability
type Manager struct {
	installDir string
	bundled    bool
}

// NewManager creates a new DCMTK manager
func NewManager() *Manager {
	return &Manager{
		installDir: getInstallDir(),
		bundled:    false,
	}
}

// CheckAvailability checks if DCMTK tools are available
func (m *Manager) CheckAvailability() (bool, error) {
	// First check system PATH
	if m.checkSystemDCMTK() {
		return true, nil
	}

	// Then check bundled DCMTK
	if m.checkBundledDCMTK() {
		m.bundled = true
		return true, nil
	}

	return false, fmt.Errorf("DCMTK not found in system PATH or bundled installation")
}

// GetDCMTKPath returns the path to the DCMTK executable
func (m *Manager) GetDCMTKPath(tool string) (string, error) {
	// Check system PATH first
	if path, err := exec.LookPath(tool); err == nil {
		return path, nil
	}

	// Check bundled installation
	bundledPath := filepath.Join(m.installDir, "dcmtk", tool)
	if runtime.GOOS == "windows" {
		bundledPath += ".exe"
	}

	if _, err := os.Stat(bundledPath); err == nil {
		return bundledPath, nil
	}

	return "", fmt.Errorf("DCMTK tool %s not found", tool)
}

// GetInstallationInfo returns information about DCMTK installation
func (m *Manager) GetInstallationInfo() *InstallationInfo {
	info := &InstallationInfo{
		Available: false,
		Bundled:   false,
		Path:      "",
		Version:   "",
		Tools:     make(map[string]string),
	}

	// Check system installation
	if systemPath, err := exec.LookPath("storescu"); err == nil {
		info.Available = true
		info.Path = filepath.Dir(systemPath)
		info.Version = m.getVersion("storescu")
		info.Tools = m.getAvailableTools(info.Path)
		return info
	}

	// Check bundled installation
	bundledDir := filepath.Join(m.installDir, "dcmtk")
	if _, err := os.Stat(bundledDir); err == nil {
		info.Available = true
		info.Bundled = true
		info.Path = bundledDir
		info.Version = m.getBundledVersion()
		info.Tools = m.getAvailableTools(bundledDir)
		return info
	}

	return info
}

// GetInstallationInstructions returns platform-specific installation instructions
func (m *Manager) GetInstallationInstructions() string {
	switch runtime.GOOS {
	case "windows":
		return m.getWindowsInstructions()
	case "darwin":
		return m.getMacOSInstructions()
	case "linux":
		return m.getLinuxInstructions()
	default:
		return m.getGenericInstructions()
	}
}

// InstallationInfo contains information about DCMTK installation
type InstallationInfo struct {
	Available bool              `json:"available"`
	Bundled   bool              `json:"bundled"`
	Path      string            `json:"path"`
	Version   string            `json:"version"`
	Tools     map[string]string `json:"tools"`
}

// checkSystemDCMTK checks if DCMTK is available in system PATH
func (m *Manager) checkSystemDCMTK() bool {
	_, err := exec.LookPath("storescu")
	return err == nil
}

// checkBundledDCMTK checks if bundled DCMTK is available
func (m *Manager) checkBundledDCMTK() bool {
	bundledDir := filepath.Join(m.installDir, "dcmtk")
	storescuPath := filepath.Join(bundledDir, "storescu")
	if runtime.GOOS == "windows" {
		storescuPath += ".exe"
	}

	_, err := os.Stat(storescuPath)
	return err == nil
}

// getInstallDir returns the installation directory
func getInstallDir() string {
	// Try to find the executable directory
	if exePath, err := os.Executable(); err == nil {
		return filepath.Dir(exePath)
	}

	// Fallback to current directory
	wd, _ := os.Getwd()
	return wd
}

// getVersion gets the version of a DCMTK tool
func (m *Manager) getVersion(tool string) string {
	path, err := m.GetDCMTKPath(tool)
	if err != nil {
		return "unknown"
	}

	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return "unknown"
}

// getBundledVersion gets the version of bundled DCMTK
func (m *Manager) getBundledVersion() string {
	versionFile := filepath.Join(m.installDir, "dcmtk", "VERSION")
	if content, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(content))
	}

	// Try to get version from bundled storescu
	return m.getVersion("storescu")
}

// getAvailableTools returns a map of available DCMTK tools
func (m *Manager) getAvailableTools(dir string) map[string]string {
	tools := make(map[string]string)
	requiredTools := []string{"storescu", "echoscu", "dcmdump", "dcmodify", "findscu"}

	for _, tool := range requiredTools {
		toolPath := filepath.Join(dir, tool)
		if runtime.GOOS == "windows" {
			toolPath += ".exe"
		}

		if _, err := os.Stat(toolPath); err == nil {
			tools[tool] = toolPath
		}
	}

	return tools
}

// getWindowsInstructions returns Windows installation instructions
func (m *Manager) getWindowsInstructions() string {
	return `DCMTK Installation Instructions for Windows:

Option 1 - Using Chocolatey (Recommended):
  choco install dcmtk

Option 2 - Using vcpkg:
  vcpkg install dcmtk

Option 3 - Manual Installation:
  1. Download DCMTK from https://dicom.offis.de/download/dcmtk/
  2. Extract to C:\dcmtk
  3. Add C:\dcmtk\bin to your PATH environment variable

Option 4 - Use Bundled DCMTK (if available):
  The installer may include a bundled version of DCMTK.`
}

// getMacOSInstructions returns macOS installation instructions
func (m *Manager) getMacOSInstructions() string {
	return `DCMTK Installation Instructions for macOS:

Option 1 - Using Homebrew (Recommended):
  brew install dcmtk

Option 2 - Using MacPorts:
  sudo port install dcmtk

Option 3 - Manual Installation:
  1. Download DCMTK from https://dicom.offis.de/download/dcmtk/
  2. Follow the build instructions for macOS
  3. Install to /usr/local/bin

Option 4 - Use Bundled DCMTK (if available):
  The installer may include a bundled version of DCMTK.`
}

// getLinuxInstructions returns Linux installation instructions
func (m *Manager) getLinuxInstructions() string {
	return `DCMTK Installation Instructions for Linux:

Ubuntu/Debian:
  sudo apt-get update
  sudo apt-get install dcmtk

CentOS/RHEL/Fedora:
  sudo yum install dcmtk
  # or for newer versions:
  sudo dnf install dcmtk

Arch Linux:
  sudo pacman -S dcmtk

Option 4 - Use Bundled DCMTK (if available):
  The installer may include a bundled version of DCMTK.`
}

// getGenericInstructions returns generic installation instructions
func (m *Manager) getGenericInstructions() string {
	return `DCMTK Installation Instructions:

1. Visit https://dicom.offis.de/download/dcmtk/
2. Download the appropriate version for your platform
3. Follow the installation instructions for your operating system
4. Ensure DCMTK binaries are in your system PATH

Required DCMTK tools:
- storescu (for C-STORE operations)
- echoscu (for C-ECHO operations)
- dcmdump (for DICOM file inspection)
- dcmodify (for DICOM file modification)`
}
