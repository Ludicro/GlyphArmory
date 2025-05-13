package main

import (
	"bufio"   // Read input line by line
	"fmt"     // Printing to terminal
	"os"      // Access to stdin and exit
	"strings" // Parsing input

	"io/fs"
	"path/filepath"
)

func main() {

	fmt.Println("Welcome to Ludicro_Armory. Type 'help' to get started.")

	// Create a reader to get input
	reader := bufio.NewReader(os.Stdin)

	// Input loop to run until user enters 'exit'
	for {
		fmt.Print("[LudicroArmory] > ") // Prompt

		// Read input until user hits Enter
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim input
		input = strings.TrimSpace(input)

		// If user types nothing, skip
		if input == "" {
			continue
		}

		// Split input
		// 1st word is command
		// nth words are arguments
		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]

		// Handle known commands
		switch command {
		case "help":
			handleHelp()
		case "use":
			handleUse(args)
		case "modules":
			modules, err := getAvailableModules()
			if err != nil {
				fmt.Println("Error:", err)
				break
			}
			fmt.Println("Available modules:")
			for _, m := range modules {
				fmt.Println("   ", m)
			}
		case "exit":
			fmt.Println("Exiting LudicroArmory...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command:", command)
		}
	}

}

func handleHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  use <module>   Load a module")
	fmt.Println("  help           Show this help message")
	fmt.Println("  exit           Exit the shell")
}

func handleUse(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: use <module>")
		return
	}
	moduleName := args[0]
	fmt.Printf("Using module: %s\n", moduleName)
	// Later we'll load the actual module here
}

// Recursively finds paths for available modules
func getAvailableModules() ([]string, error) {
	var modules []string

	// Recursively get modules and directories
	err := filepath.WalkDir("modules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Skip unreadable
		}

		// Currently only look for go files
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			// Strip "modules/" and ".go" from the path
			trimmed := strings.TrimPrefix(path, "modules/")
			trimmed = strings.TrimSuffix(trimmed, ".go")
			modules = append(modules, trimmed)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil

}
