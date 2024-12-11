#!/bin/zsh

# The tool goal is to read the directories from app/commands
# And generate the build script in makefile between BUILD_COMMANDS markers

LIST_DIR=./app/commands
MAKEFILE="Makefile"
START_MARKER="# -- BUILD_COMMANDS START --"
END_MARKER="# -- BUILD_COMMANDS END --"

# Check if required directory exists
if [[ ! -d "$LIST_DIR" ]]; then
    echo "Error: Directory $LIST_DIR does not exist"
    exit 1
fi

# Function to generate build commands for a directory
generate_build_command() {
    local dir=$1
    local cmd_name=$(basename $dir)

    echo "build_$cmd_name:"
    echo "\t@echo \"\$(BLUE)Building \$(GREEN)${cmd_name}\$(RESET)\""
    echo "\t@go build -o bin/$cmd_name app/commands/$cmd_name/main.go"
    echo ""
}

# Function to generate all commands
generate_all_commands() {
    local dirs=($LIST_DIR/*)
    local commands=""

    echo "# Auto-generated build commands"
    echo "build-all: $(for d in $dirs; do [[ -d "$d" ]] && echo -n "build_$(basename $d) "; done)"
    echo "\n"

    for dir in $dirs; do
        if [[ -d "$dir" ]]; then
            generate_build_command "$dir"
        fi
    done
}

# Check if Makefile exists
if [[ ! -f "$MAKEFILE" ]]; then
    echo "Error: $MAKEFILE not found"
    exit 1
fi

# Create temporary file
TMP_FILE=$(mktemp)

# Process the Makefile
in_build_section=false
while IFS= read -r line; do
    if [[ "$line" == "$START_MARKER" ]]; then
        echo "$line" >> "$TMP_FILE"
        generate_all_commands >> "$TMP_FILE"
        in_build_section=true
    elif [[ "$line" == "$END_MARKER" ]]; then
        echo "$line" >> "$TMP_FILE"
        in_build_section=false
    elif [[ "$in_build_section" == "false" ]]; then
        echo "$line" >> "$TMP_FILE"
    fi
done < "$MAKEFILE"

# Check if both markers were found
if ! grep -q "$START_MARKER" "$TMP_FILE" || ! grep -q "$END_MARKER" "$TMP_FILE"; then
    echo "Error: Markers $START_MARKER and $END_MARKER must both exist in $MAKEFILE"
    rm "$TMP_FILE"
    exit 1
fi

# Replace original file
mv "$TMP_FILE" "$MAKEFILE"

echo "Successfully updated $MAKEFILE"
