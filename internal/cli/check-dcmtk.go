package cli

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/dcmtk"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CreateCheckDCMTKCommand returns the check-dcmtk command
func CreateCheckDCMTKCommand() *cli.Command {
	return &cli.Command{
		Name:    "check-dcmtk",
		Usage:   "Check DCMTK installation and provide setup instructions",
		Aliases: []string{"check", "dcmtk-status"},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Show detailed DCMTK information",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "install-help",
				Usage: "Show installation instructions",
				Value: false,
			},
		},
		Action: checkDCMTKAction,
	}
}

func checkDCMTKAction(c *cli.Context) error {
	verbose := c.Bool("verbose")
	showInstallHelp := c.Bool("install-help")

	manager := dcmtk.NewManager()
	
	fmt.Println("🔍 Checking DCMTK Installation...")
	fmt.Println()

	// Check availability
	available, err := manager.CheckAvailability()
	if err != nil {
		fmt.Printf("❌ DCMTK Status: %v\n", err)
		fmt.Println()
		
		if showInstallHelp {
			fmt.Println("📋 Installation Instructions:")
			fmt.Println(manager.GetInstallationInstructions())
			return nil
		}
		
		fmt.Println("💡 To see installation instructions, run:")
		fmt.Println("   crgodicom check-dcmtk --install-help")
		return nil
	}

	// Get installation info
	info := manager.GetInstallationInfo()
	
	if available {
		fmt.Printf("✅ DCMTK Status: Available\n")
		if info.Bundled {
			fmt.Printf("📦 Installation: Bundled with CRGoDICOM\n")
		} else {
			fmt.Printf("💻 Installation: System installation\n")
		}
		fmt.Printf("📍 Path: %s\n", info.Path)
		fmt.Printf("🔢 Version: %s\n", info.Version)
		
		if verbose {
			fmt.Println()
			fmt.Println("🛠️  Available Tools:")
			for tool, path := range info.Tools {
				fmt.Printf("   • %s: %s\n", tool, path)
			}
		}
		
		// Test key tools
		fmt.Println()
		fmt.Println("🧪 Testing Key Tools:")
		
		keyTools := []string{"storescu", "echoscu"}
		for _, tool := range keyTools {
			if path, err := manager.GetDCMTKPath(tool); err == nil {
				fmt.Printf("   ✅ %s: Available at %s\n", tool, path)
			} else {
				fmt.Printf("   ❌ %s: %v\n", tool, err)
			}
		}
		
		fmt.Println()
		fmt.Println("🎉 DCMTK is ready for use with CRGoDICOM!")
		
	} else {
		fmt.Printf("❌ DCMTK Status: Not Available\n")
		fmt.Println()
		
		if showInstallHelp {
			fmt.Println("📋 Installation Instructions:")
			fmt.Println(manager.GetInstallationInstructions())
		} else {
			fmt.Println("💡 To see installation instructions, run:")
			fmt.Println("   crgodicom check-dcmtk --install-help")
		}
		
		return fmt.Errorf("DCMTK is required for PACS operations")
	}

	return nil
}

// CheckDCMTKAvailability is a helper function for other commands
func CheckDCMTKAvailability() error {
	manager := dcmtk.NewManager()
	
	_, err := manager.CheckAvailability()
	if err != nil {
		logrus.Warnf("DCMTK not available: %v", err)
		logrus.Info("Run 'crgodicom check-dcmtk --install-help' for installation instructions")
		return err
	}
	
	info := manager.GetInstallationInfo()
	if info.Bundled {
		logrus.Info("Using bundled DCMTK installation")
	} else {
		logrus.Info("Using system DCMTK installation")
	}
	
	return nil
}

// GetDCMTKPath is a helper function to get DCMTK tool path
func GetDCMTKPath(tool string) (string, error) {
	manager := dcmtk.NewManager()
	return manager.GetDCMTKPath(tool)
}

// ShowDCMTKInstallationHelp displays installation help
func ShowDCMTKInstallationHelp() {
	manager := dcmtk.NewManager()
	fmt.Println("📋 DCMTK Installation Instructions:")
	fmt.Println(manager.GetInstallationInstructions())
}
