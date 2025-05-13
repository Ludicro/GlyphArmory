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

	fmt.Println("Welcome to Ludicro_Armory. Type 'help' to get started or 'exit' to quit.")

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
	// Uses readline to autocomplete available commands across the entire tool
	completer := readline.NewPrefixCompleter(
		readline.PcItem("use", moduleSuggestions...),
		readline.PcItem("modules"),
		readline.PcItem("info"),
		readline.PcItem("return"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
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
		rl.SetPrompt(buildPrompt()) //Dynamically set the prompt to include the module

		// Read the line in the terminal and save it
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

		// Handle the command used
		switch command {
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
		case "info":
			handleInfo()
		case "return":
			handleReturn()
		case "help":
			handleHelp()
		case "exit":
			fmt.Println("Exiting LudicroArmory...")
			os.Exit(0)
		default:
			fmt.Println("Unknown command:", command)
		}
	}
}

// === Console Command Functions ===

// Prints the help statements
func handleHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  use <module>   Load a module")
	fmt.Println("  modules        Display available modules")
	fmt.Println("  info           Displays information on currently selected module")
	fmt.Println("  return         Clear the selected module")
	fmt.Println("  help           Show this help message")
	fmt.Println("  exit           Exit the shell")
}

// Sets the module to be used
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
	// Cycle through all the available modules to make sure the selection is valid
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
}

// Displays info from module's info file
func handleInfo() {
	if currentModule == "" {
		fmt.Println("No module selected.")
		return
	}

	infoPath := filepath.Join("modules/", currentModule, "info")

	infoContents, err := os.ReadFile(infoPath)
	if err != nil {
		fmt.Println("Error reading info file:", err)
	}

	fmt.Println("Contents of info file:\n", string(infoContents))

}

// Resets the currentModule
func handleReturn() {
	currentModule = ""
}

// === Utility Functions ===

// Recursively finds paths for available modules
func getAvailableModules() ([]string, error) {
	var modules []string

	// Recursively get modules and directories
	err := filepath.WalkDir("modules", func(path string, directory fs.DirEntry, err error) error {
		if err != nil {
			return err // Skip unreadable
		}

		// Currently directories
		if directory.IsDir() {

			// Check if it contains an info file
			infoPath := filepath.Join(path, "info")

			if fileExists(infoPath) {
				// Strip "modules/"
				trimmed := strings.TrimPrefix(path, "modules/")
				modules = append(modules, trimmed)
			}

		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return modules, nil

}

// Sets the terminal prompt depending on the current module
// Returns the prompt
func buildPrompt() string {
	if currentModule != "" {
		return fmt.Sprintf("[LudicroArmory] (%s) > ", currentModule)
	}
	return "[LudicroArmory] > "
}

// Verifies a file exists when given the path
// Returns true if file is found
// Returns false if file is NOT found
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
