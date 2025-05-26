#!/bin/bash

# === Uninstall Stealth Reverse Shell Hook for 'cd' ===
# Cleans /etc/bash.bashrc, removes cd() function if present

TARGET_BASHRC="/etc/bash.bashrc"

echo "[+] Cleaning up stealth 'cd' hook..."

# Remove the 'cd hook' block (our known signature)
awk '
  BEGIN { skip=0 }
  /# cd hook/ { skip=1; next }
  skip && /^\}/ { skip=0; next }
  skip { next }
  { print }
' "$TARGET_BASHRC" > "${TARGET_BASHRC}.tmp" && mv "${TARGET_BASHRC}.tmp" "$TARGET_BASHRC"

echo "[+] Hook removed from $TARGET_BASHRC"

echo "[+] Uninstall complete and cleaned."
