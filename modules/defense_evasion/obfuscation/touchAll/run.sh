#!/bin/bash

# === Obfuscation via Touch ===
# Touches every file the running user has access to in order to hide modifications

echo "[INFO] Touching all writable files under: $TDIR"

# Recursively touch all writable files
find "$TDIR" -type f 2>/dev/null | while read -r file; do
    if [ -w "$file" ]; then
        touch "$file"
        echo "[TOUCHED] $file"
    fi
done

echo "[DONE] File timestamps updated."