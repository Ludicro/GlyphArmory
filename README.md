# Glyph Armory

**Glyph Armory** is a modular, embedded red team shell inspired by tools like `msfconsole`. It provides an interactive CLI for managing, configuring, and executing post-exploitation and offensive security scripts.

> ⚠️ **Disclaimer:** This tool is intended for authorized security assessments only. Do **not** use it on systems you do not own or explicitly have permission to test.

---

## 🧠 Features

- 🔒 Fully self-contained single binary (no external scripts required)
- 🧠 Modular structure for OSINT, persistence, shells, and more
- 🖥️ In-memory script execution — no disk artifacts
- 🧩 Configurable module options with default values
- 🌲 Module tree exploration with `tree` command
- 🎨 Terminal color coding and dynamic prompts

---

## 📁 Directory Structure

```
modules/
├── directory1/
│   └── directory1a/
│       └── module1/
│           ├── run.sh
│           ├── config
│           └── info
└── directory2/
    └── directory2a/
        └── module2/
            ├── run.sh
            ├── config
            └── info
```

Each module contains:
- `run.sh` — the executable payload (bash or other)
- `config` — `KEY:DEFAULT:DESCRIPTION` style config file
- `info` — plain-text info about the module

---

## 🚀 Usage

Start the shell:

```bash
./glyph_armory
```

### Example Commands

```
use persistence/simple_shell
show
set RHOST 10.0.0.5
run
```

---

## 📚 Command Summary

| Command              | Description                                 |
|----------------------|---------------------------------------------|
| `use <module>`       | Select a module to use                      |
| `return`             | Clear the selected module                   |
| `info`               | Show module description                     |
| `set <key> <value>`  | Set a module option                         |
| `show`               | Display current config + default values     |
| `run`                | Execute the module                          |
| `modules`            | List all available modules                  |
| `tree`               | Display the module folder structure         |
| `clear`              | Clear the terminal screen                   |
| `help`               | Display help menu                           |
| `exit`               | Exit the shell                              |

---

## ✨ In-Memory Execution

All scripts are executed via `bash` using `stdin` only — no files are written to disk.

---

## 📜 License

This tool is for educational and authorized testing purposes only. Use responsibly.
