#!/usr/bin/env bash

# Bump the API version in configs/api.yaml and cmd/api/main.go.
# Usage:
#   ./scripts/bump_version.sh <new_version>
#   ./scripts/bump_version.sh patch|minor|major

set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

API_CONFIG_FILE="$BACKEND_DIR/configs/api.yaml"
API_MAIN_FILE="$BACKEND_DIR/cmd/api/main.go"
API_CONFIG_REL="configs/api.yaml"
API_MAIN_REL="cmd/api/main.go"

usage() {
    echo "Usage: $0 <new_version | patch | minor | major>"
    echo ""
    echo "Examples:"
    echo "  $0 1.0.0    Set API version to 1.0.0"
    echo "  $0 patch    Bump patch version  (x.y.Z)"
    echo "  $0 minor    Bump minor version  (x.Y.0)"
    echo "  $0 major    Bump major version  (X.0.0)"
}

if [[ $# -ne 1 ]]; then
    usage
    exit 1
fi

ensure_version_files_clean() {
    local files=("$API_CONFIG_REL" "$API_MAIN_REL")

    if ! git -C "$BACKEND_DIR" diff --quiet -- "${files[@]}" \
        || ! git -C "$BACKEND_DIR" diff --cached --quiet -- "${files[@]}"; then
        echo -e "${RED}Version files already have local changes. Commit or stash them first.${NC}"
        exit 1
    fi
}

validate_semver() {
    if ! [[ "$1" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}Invalid version: $1${NC}"
        echo "Expected format: X.Y.Z (e.g. 1.2.3)"
        exit 1
    fi
}

bump_semver() {
    local current="$1"
    local part="$2"
    IFS='.' read -r major minor patch <<< "$current"
    case "$part" in
        major) echo "$((major + 1)).0.0" ;;
        minor) echo "$major.$((minor + 1)).0" ;;
        patch) echo "$major.$minor.$((patch + 1))" ;;
        *)
            echo -e "${RED}Unknown bump type: $part${NC}"
            exit 1
            ;;
    esac
}

get_api_version() {
    grep -oP '^\s*version:\s*"\K[0-9]+\.[0-9]+\.[0-9]+' \
        "$API_CONFIG_FILE"
}

update_api_version() {
    local current="$1"
    local new="$2"

    sed -i "s/version: \"$current\"/version: \"$new\"/" \
        "$API_CONFIG_FILE"
    echo -e "  ${GREEN}✓${NC} configs/api.yaml"

    sed -i "s/^\/\/ @version.*$/\/\/ @version         $new/" \
        "$API_MAIN_FILE"
    echo -e "  ${GREEN}✓${NC} cmd/api/main.go"
}

CURRENT="$(get_api_version)"

ensure_version_files_clean

if [[ -z "$CURRENT" ]]; then
    echo -e "${RED}Could not detect current API version from configs/api.yaml${NC}"
    exit 1
fi

INPUT="$1"

case "$INPUT" in
    patch|minor|major)
        NEW="$(bump_semver "$CURRENT" "$INPUT")"
        ;;
    *)
        NEW="$INPUT"
        ;;
esac

validate_semver "$NEW"

if [[ "$NEW" = "$CURRENT" ]]; then
    echo -e "${YELLOW}API version is already $CURRENT${NC}"
    exit 0
fi

echo -e "${YELLOW}Bumping API version: $CURRENT → $NEW${NC}"
echo ""
update_api_version "$CURRENT" "$NEW"
echo ""
echo -e "${GREEN}API version updated to $NEW${NC}"
