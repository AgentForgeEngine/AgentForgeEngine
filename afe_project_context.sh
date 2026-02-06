#!/bin/bash

# Script to list all files in project with headers, showing first 50 lines of each file
echo "=== PROJECT FILE STRUCTURE (excluding ./afe) ==="
echo

# Find all files, excluding ./afe and hidden directories
find . -type f ! -path "*/.*" ! -path "./afe" | sort | while read -r file; do
    if [ -f "$file" ] && [ "$file" != "./afe" ]; then
        echo "=== $file ==="
        # Show first 50 lines of the file
        head -n 50 "$file" 2>/dev/null
        echo
    fi
done

echo "=== END OF PROJECT FILES ==="

