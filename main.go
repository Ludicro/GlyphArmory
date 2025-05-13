package main

import (
	// Read input line by line
	"fmt" // Printing to terminal
	"io"
	"os" // Access to stdin and exit
	"os/exec"
	"strings" // Parsing input

	"io/fs"
	"path/filepath"

	"github.com/chzyer/readline"
)

// Global variables
var currentModule string
var moduleConfigs = make(map[string]map[string]string)

type ConfigEntry struct {
	Default     string
	Description string
}

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
		readline.PcItem("run"),
		readline.PcItem("set"),
		readline.PcItem("show"),
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
		case "run":
			handleRun()
		case "set":
			handleSet(args)
		case "show":
			handleShow()
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
	fmt.Println("  use <module>   		 Load a module")
	fmt.Println("  modules        		 Displays available modules")
	fmt.Println("  info           		 Displays information on currently selected module")
	fmt.Println("  run            		 Deploys the selected script")
	fmt.Println("  set <key> <value>     Sets a config option")
	fmt.Println("  show           		 Displays current config")
	fmt.Println("  return         		 Clear the selected module")
	fmt.Println("  help           		 Show this help message")
	fmt.Println("  exit           		 Exit the shell")
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

	infoPath := filepath.Join("modules", currentModule, "info")

	infoContents, err := os.ReadFile(infoPath)
	if err != nil {
		fmt.Println("Error reading info file:", err)
	}

	fmt.Println(string(infoContents))

}

// Executes the payload from the module
func handleRun() {
	if currentModule == "" {
		fmt.Println("No module selected.")
		return
	}

	payloadPath := filepath.Join("modules", currentModule, "run.sh")

	// Make sure file exists
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		fmt.Printf("No 'run.sh' script found for module: %s\n", currentModule)
		return
	}

	// Load config
	configPath := filepath.Join("modules", currentModule, "config")
	expected, _ := parseModuleConfig(configPath)

	// Build env
	env := os.Environ()

	// Get current config values for module
	current := moduleConfigs[currentModule]

	// Inject each expected variable
	// Inject each expected variable
	for key, entry := range expected {
		val := current[key]
		if val == "" {
			val = entry.Default
		}
		env = append(env, fmt.Sprintf("%s=%s", key, val))
	}

	// Add any custom (unexpected) keys
	for key, val := range current {
		if _, found := expected[key]; !found {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	// Set up the command to run
	cmd := exec.Command(payloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env // inject the final env

	// Execute it
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running module:", err)
	}

}

// Sets a config option
func handleSet(args []string) {

	// Ensure a module is selected
	if currentModule == "" {
		fmt.Println("No module selected.")
		return
	}

	// Must provide a key and value
	if len(args) != 2 {
		fmt.Println("Usage: set <key> <value>")
		return
	}

	key := args[0]                       // First argument is the key
	value := strings.Join(args[1:], " ") // Joins the rest in case of spaces and sets them as the value

	// Init config map for the module if it doesn't exist
	if _, exists := moduleConfigs[currentModule]; !exists {
		moduleConfigs[currentModule] = make(map[string]string)
	}

	// Store the key-value in the current module's config map
	moduleConfigs[currentModule][key] = value
	fmt.Printf("[*] Set %s = %s for module %s\n", key, value, currentModule)
}

// Displays config options
func handleShow() {
	if currentModule == "" {
		fmt.Println("No module selected.")
		return
	}

	// Determine path to config file
	configPath := filepath.Join("modules", currentModule, "config")

	// Parse the config file to get expected keys, defaults, and descriptions
	expected, orderedKeys := parseModuleConfig(configPath)

	// Get any current values set by the user for this module
	current := moduleConfigs[currentModule]

	// Exit early if there's nothing to show
	if len(expected) == 0 && len(current) == 0 {
		fmt.Println("No config information found for this module.")
		return
	}

	// Get column widths
	maxKeyLen := len("Name")
	maxValLen := len("Current Value")
	maxDefLen := len("Default")

	for _, key := range orderedKeys {
		entry := expected[key]
		val := current[key]
		if val == "" {
			val = "(default)"
		}
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
		if len(val) > maxValLen {
			maxValLen = len(val)
		}
		if len(entry.Default) > maxDefLen {
			maxDefLen = len(entry.Default)
		}
	}
	// Also check for custom keys (not in config file)
	for key, val := range current {
		if _, found := expected[key]; !found {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}
			if len(val) > maxValLen {
				maxValLen = len(val)
			}
		}
	}

	// Build format string
	format := fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%s\n", maxKeyLen, maxValLen, maxDefLen)

	// Print table header
	fmt.Printf("Config for module: %s\n", currentModule)
	fmt.Printf(format, "Name", "Current Value", "Default", "Description")
	fmt.Printf(format, strings.Repeat("-", maxKeyLen), strings.Repeat("-", maxValLen), strings.Repeat("-", maxDefLen), strings.Repeat("-", 20))

	// Show expected keys and values in order
	for _, key := range orderedKeys {
		entry := expected[key]
		val := current[key]

		// Show default value if no value was set
		if val == "" {
			val = "(default)"
		}
		fmt.Printf(format, key, val, entry.Default, entry.Description)
	}

	// Show any user-set keys that aren't defined in the config file
	for key, val := range current {
		if _, found := expected[key]; !found {
			fmt.Printf("%-*s  %-*s  %s\n", maxKeyLen, key, maxValLen, val, "(custom key)")
		}
	}
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

// Gets the config information from config file in module
func parseModuleConfig(configPath string) (map[string]ConfigEntry, []string) {
	configMap := make(map[string]ConfigEntry)
	orderedKeys := []string{}

	// Skip if file doesn't exist
	if !fileExists(configPath) {
		return configMap, orderedKeys
	}

	// Read raw information
	rawContent, err := os.ReadFile(configPath)
	if err != nil {
		return configMap, orderedKeys
	}

	// Split each line
	lines := strings.Split(string(rawContent), "\n")
	// For each line
	for _, raw := range lines {
		line := strings.TrimSpace(raw) // Remove whitespace
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 3) // Split into key:value:desc

		// Fill default value
		key := strings.TrimSpace(parts[0])
		defaultVal := ""
		if len(parts) > 1 {
			defaultVal = strings.TrimSpace(parts[1])
		}

		// Fill description
		desc := ""
		if len(parts) > 2 {
			desc = strings.TrimSpace(parts[2])
		}

		// Set config map
		configMap[key] = ConfigEntry{
			Default:     defaultVal,
			Description: desc,
		}
		orderedKeys = append(orderedKeys, key)
	}

	return configMap, orderedKeys
}
