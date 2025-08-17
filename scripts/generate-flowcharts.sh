#! /bin/bash

ROOT_DIR="assets"

for file in $ROOT_DIR/*.mmd; do
    mmdc -i "$file" -o "${file%.mmd}.svg"
done

echo "Flowcharts generated successfully"