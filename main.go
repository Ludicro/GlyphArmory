package main

import (
	// Read input line by line
	"fmt" // Printing to terminal
	"io"
	"os" // Access to stdin and exit
	"os/exec"
	"sort"
	"strings" // Parsing input

	"io/fs"
	"path/filepath"

	"github.com/chzyer/readline"
)

const (
	Green  = "\033[32m"
	Cyan   = "\033[94m"
	Red    = "\033[31m"
	Yellow = "\033[93m"
	Reset  = "\033[0m"
)

// Global variables
var currentModule string
var moduleConfigs = make(map[string]map[string]string)

type ConfigEntry struct {
	Default     string
	Description string
}

func main() {

	fmt.Println(Green + "                  zzzzzzzzzzzzzz                   ")
	fmt.Println("              zzzzzzz        zzzzzzz              ")
	fmt.Println("          zzzzz        zzzz        zzzzz          ")
	fmt.Println("        zzzz         zzzzzzzz         zzzz        ")
	fmt.Println("      zzzz           zz    zz           zzzz      ")
	fmt.Println("     zzz     zzzzzzz zzzzzzzz zzzzzzz     zzz     ")
	fmt.Println("    zzz    zzzz   zzz  zzzz  zzz   zzzz    zzz    ")
	fmt.Println("   zz    zzz                          zzz    zz   ")
	fmt.Println("  zz             zzzzzzzzzzzzzzzzz            zz  ")
	fmt.Println(" zzz         zzzzzzz           zzzzzzz        zzz ")
	fmt.Println(" zz        zzzz   zz    zz    zz   zzzz        zz ")
	fmt.Println("zzz      zzzz     zz    zz    zz     zzz       zzz")
	fmt.Println("zzz               zz    zz    zz               zzz")
	fmt.Println("zz               zzz    zz    zzz               zz")
	fmt.Println("zzz             zzz     zz     zzz             zzz")
	fmt.Println("zzz           zzz       zz       zzz           zzz")
	fmt.Println(" zz                     zz                     zz ")
	fmt.Println(" zzz                   zzz                    zzz ")
	fmt.Println("  zz                   zz                     zz  ")
	fmt.Println("   zz                 zzz                    zz   ")
	fmt.Println("    zzz              zzz                   zzz    ")
	fmt.Println("     zzz            zzz                   zzz     ")
	fmt.Println("      zzzz        zzz                   zzzz      ")
	fmt.Println("        zzzz     zzz                  zzzz        ")
	fmt.Println("          zzzzz                    zzzzz          ")
	fmt.Println("              zzzzzzz        zzzzzzz              ")
	fmt.Println("                  zzzzzzzzzzzzzz                  " + Reset)

	fmt.Println(Yellow + "NOTICE: Usage of this tool is for ethical and permitted operations only.\n" +
		"        Do not use this on systems you do not own!" + Reset)

	fmt.Println("Welcome to Glyph Armory. Type 'help' to get started or 'exit' to quit.")

	// Build autocompleter from known modules
	modules, err := getAvailableModules()
	if err != nil {
		fmt.Println(Red + "[Error] Failed to load modules:" + err.Error() + Reset)
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
		// Module Commands
		readline.PcItem("use", moduleSuggestions...),
		readline.PcItem("return"),
		readline.PcItem("info"),
		readline.PcItem("set"),
		readline.PcItem("show"),
		readline.PcItem("run"),

		//General commands
		readline.PcItem("modules"),
		readline.PcItem("tree"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	// Initialize the readline instance with completer
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "[GlyphArmory] > ",
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
				fmt.Println(Red + "[Error]" + err.Error() + Reset)
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
		case "tree":
			handleTree()
		case "help":
			handleHelp()
		case "exit":
			fmt.Println(Yellow + "[!] Exiting GlyphArmory..." + Reset)
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
	fmt.Println("  return         		 Clear the selected module")
	fmt.Println("  info           		 Displays information on currently selected module")
	fmt.Println("  set <key> <value>     Sets a config option")
	fmt.Println("  show           		 Displays current config")
	fmt.Println("  run            		 Deploys the selected script")
	fmt.Println("  modules        		 Displays available modules")
	fmt.Println("  tree                  Displays a tree view of all available modules")
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
		fmt.Println(Red + "[Error] Error reading modules:" + err.Error() + Reset)
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
		fmt.Printf(Yellow+"[!] Module '%s' not found.\n"+Reset, requested)
		fmt.Println(Yellow + "[!] Use the 'modules' command to list available modules." + Reset)
		return
	}

	// Save selected module
	currentModule = requested
}

// Displays info from module's info file
func handleInfo() {
	if currentModule == "" {
		fmt.Println(Yellow + "[!] No module selected." + Reset)
		return
	}

	infoPath := filepath.Join("modules", currentModule, "info")

	infoContents, err := os.ReadFile(infoPath)
	if err != nil {
		fmt.Println(Red + "[Error] Error reading info file:" + err.Error() + Reset)
	}

	fmt.Println(string(infoContents))

}

// Executes the payload from the module
func handleRun() {
	if currentModule == "" {
		fmt.Println(Yellow + "[!] No module selected." + Reset)
		return
	}

	payloadPath := filepath.Join("modules", currentModule, "run.sh")

	// Make sure file exists
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		fmt.Printf(Red+"[Error] No 'run.sh' script found for module: %s\n"+Reset, currentModule)
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
	cmd := exec.Command("bash", payloadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = env // inject the final env

	// Execute it
	err := cmd.Run()
	if err != nil {
		fmt.Println(Red + "[Error] Error running module:" + err.Error() + Reset)
	}

}

// Sets a config option
func handleSet(args []string) {

	// Ensure a module is selected
	if currentModule == "" {
		fmt.Println(Yellow + "[!] No module selected." + Reset)
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
		fmt.Println(Yellow + "[!] No module selected." + Reset)
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
		fmt.Println(Yellow + "[!] No config information found for this module." + Reset)
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

// Displays available modules as a tree
func handleTree() {
	basePath := "modules"
	fmt.Println("Module tree:")
	printTree(basePath, "")
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
		return fmt.Sprintf(Green+"[GlyphArmory] "+Cyan+"("+"%s"+") "+Green+"> "+Reset, currentModule)
	}
	return Green + "[GlyphArmory] > " + Reset
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

// Prints a tree of directory recursively
func printTree(path string, prefix string) {

	// Get the contents of the current directory path
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(prefix + "└── (error reading)")
		return
	}

	// Only get the directories
	dirs := []fs.DirEntry{}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		}
	}

	// Sort all entries
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Name() < dirs[j].Name()
	})

	for i, entry := range dirs {
		isFinal := i == len(dirs)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if isFinal {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		fmt.Println(prefix + connector + entry.Name())

		// Build full path to check what's inside
		fullPath := filepath.Join(path, entry.Name())

		// Check if this is a module folder
		if isModuleFolder(fullPath) {
			continue //Is a module folder so don't go into it
		}

		// Recursive
		printTree(fullPath, nextPrefix)

	}

}

func isModuleFolder(path string) bool {
	files := []string{"run.sh", "run", "info", "config"}
	// Checks if path has module files
	for _, f := range files {
		if fileExists(filepath.Join(path, f)) {
			return true
		}
	}
	return false
}
