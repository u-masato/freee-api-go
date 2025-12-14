#!/bin/bash
# update-openapi.sh - Download the latest freee API OpenAPI specification
#
# Usage:
#   ./tools/update-openapi.sh
#
# This script downloads the OpenAPI specification from the official freee API schema repository
# and saves it to api/openapi.json

set -e  # Exit on error
set -u  # Exit on undefined variable

# Configuration
OPENAPI_URL="https://raw.githubusercontent.com/freee/freee-api-schema/master/v2020_06_15/open-api-3/api-schema.json"
OUTPUT_FILE="api/openapi.json"
TEMP_FILE="api/openapi.json.tmp"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the project root directory (parent of tools/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

echo -e "${YELLOW}Downloading freee API OpenAPI specification...${NC}"
echo "Source: ${OPENAPI_URL}"
echo "Destination: ${OUTPUT_FILE}"
echo ""

# Create api directory if it doesn't exist
mkdir -p api

# Download to temporary file
if curl -f -L -o "${TEMP_FILE}" "${OPENAPI_URL}"; then
    echo -e "${GREEN}✓ Download successful${NC}"

    # Validate that it's a valid JSON file
    if jq empty "${TEMP_FILE}" 2>/dev/null; then
        echo -e "${GREEN}✓ JSON validation successful${NC}"

        # Extract version information
        OPENAPI_VERSION=$(jq -r '.openapi' "${TEMP_FILE}")
        API_VERSION=$(jq -r '.info.version' "${TEMP_FILE}")
        API_TITLE=$(jq -r '.info.title' "${TEMP_FILE}")

        echo ""
        echo "OpenAPI Specification Details:"
        echo "  OpenAPI Version: ${OPENAPI_VERSION}"
        echo "  API Title: ${API_TITLE}"
        echo "  API Version: ${API_VERSION}"
        echo ""

        # Check if file has changed
        if [ -f "${OUTPUT_FILE}" ]; then
            if cmp -s "${TEMP_FILE}" "${OUTPUT_FILE}"; then
                echo -e "${YELLOW}⚠ No changes detected. File is already up to date.${NC}"
                rm "${TEMP_FILE}"
                exit 0
            else
                echo -e "${YELLOW}⚠ File has changed. Updating...${NC}"
            fi
        fi

        # Move temporary file to final location
        mv "${TEMP_FILE}" "${OUTPUT_FILE}"
        echo -e "${GREEN}✓ OpenAPI specification updated successfully${NC}"
        echo ""
        echo "File saved to: ${OUTPUT_FILE}"

        # Show file size
        FILE_SIZE=$(du -h "${OUTPUT_FILE}" | cut -f1)
        echo "File size: ${FILE_SIZE}"

    else
        echo -e "${RED}✗ Error: Downloaded file is not valid JSON${NC}"
        rm -f "${TEMP_FILE}"
        exit 1
    fi
else
    echo -e "${RED}✗ Error: Failed to download OpenAPI specification${NC}"
    rm -f "${TEMP_FILE}"
    exit 1
fi

echo ""
echo -e "${GREEN}Done!${NC}"
