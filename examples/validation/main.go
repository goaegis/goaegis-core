package main

import (
	"fmt"
	"log"

	aegis "github.com/goaegis/goaegis-core/aegis/core"
)

func main() {
	fmt.Println("🔍 Demonstrating Configuration Validation")
	fmt.Println()

	// Example 1: Valid configuration
	fmt.Println("Example 1: Loading VALID configuration...")
	a1 := aegis.New()
	if err := a1.LoadConfig("../simple/config.yaml"); err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("Valid configuration loaded successfully!")
	}
	fmt.Println()

	// Example 2: Invalid configuration with multiple errors
	fmt.Println("Example 2: Loading INVALID configuration...")
	a2 := aegis.New()
	if err := a2.LoadConfig("./invalid.yaml"); err != nil {
		fmt.Println("Validation caught errors:")
		fmt.Printf("%v", err)
	} else {
		fmt.Println("Configuration loaded (unexpected)")
	}
	fmt.Println()

	fmt.Println("\nValidation Features:")
	fmt.Println("  • Duplicate detection (resources, roles, subjects)")
	fmt.Println("  • Unknown resource references in permissions")
	fmt.Println("  • Unknown role references in subjects")
	fmt.Println("  • Unknown role references in inheritance")
	fmt.Println("  • Circular role inheritance detection")
	fmt.Println("  • Invalid effect values (must be 'allow' or 'deny')")
	fmt.Println("  • Empty/missing required fields")
	fmt.Println("  • Name/key consistency checks")
}
