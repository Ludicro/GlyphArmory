package main

import (
	// Read input line by line
	"fmt" // Printing to terminal
	"io"
	"os"      // Access to stdin and exit
	"strings" // Parsing input

	"io/fs"
	"path/filepath"

	"github.com/chzyer/readline"
)

// Global functions
var currentModule string

func main() {

	fmt.Println("Welcome to Ludicro_Armory. Type 'help' to get started.")

	// Build autocompleter from known modules
	modules, err := getAvailableModules()
	if err != nil {
		fmt.Println("Failed to load modules:", err)
		return
	}

	// Wrap each module as 'use <module>' for autocomplete suggestion
	var moduleSuggestions []readline.PrefixCompleterInterface
	for _, m := range modules {
		moduleSuggestions = append(moduleSuggestions, readline.PcItem(m))
	}

	// Root completer
	completer := readline.NewPrefixCompleter(
		readline.PcItem("help"),
		readline.PcItem("exit"),
		readline.PcItem("modules"),
		readline.PcItem("use", moduleSuggestions...),
	)

	// Initialize the readline instance with completer
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "[LudicroArmory] > ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	// Main loop

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		}

		// Trim empty spaces
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		// Break up inputs into commands and arguments
		parts := strings.Fields(input)
		command := parts[0]
		args := parts[1:]

		switch command {
		case "help":
			handleHelp()
		case "use":
			handleUse(args)
		case "modules":
			mods, err := getAvailableModules()
			if err != nil {
				fmt.Println("Error:", err)
				break
			}
			fmt.Println("Available modules:")
			for _, m := range mods {
				fmt.Println("  ", m)
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
	fmt.Println("  modules        Display available modules")
	fmt.Println("  help           Show this help message")
	fmt.Println("  exit           Exit the shell")
}

func handleUse(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: use <module>")
		return
	}
	requested := args[0]

	// Get available modules
	available, err := getAvailableModules()
	if err != nil {
		fmt.Println("Error reading modules:", err)
		return
	}

	// Check if requested module is valid
	valid := false
	for _, mod := range available {
		if mod == requested {
			valid = true
			break
		}
	}

	if !valid {
		fmt.Printf("Module '%s' not found.\n", requested)
		fmt.Println("Use the 'modules' command to list available modules.")
		return
	}

	// Save selected module
	currentModule = requested
	fmt.Printf("Using module: %s\n", currentModule)

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
