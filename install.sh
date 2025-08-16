#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Installing Coolify CLI...${NC}\n"

# Determine OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

# Convert architecture names
case "${ARCH}" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: ${ARCH}${NC}" && exit 1 ;;
esac

# Convert OS names
case "${OS}" in
    Linux) OS="linux" ;;
    Darwin) OS="darwin" ;;
    *) echo -e "${RED}Unsupported operating system: ${OS}${NC}" && exit 1 ;;
esac

# Set installation paths
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="${HOME}/.coolify-cli"
BINARY_NAME="coolify-cli"
BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"

# Create config directory if it doesn't exist
if [ ! -d "${CONFIG_DIR}" ]; then
    echo -e "${BLUE}Creating config directory...${NC}"
    mkdir -p "${CONFIG_DIR}"
fi

# Download and install binary
echo -e "${BLUE}Downloading Coolify CLI...${NC}"
LATEST_RELEASE_URL="https://github.com/vaarvik/coolify-cli/releases/latest/download/coolify-cli-${OS}-${ARCH}"

# Check if we have curl or wget
if command -v curl > /dev/null 2>&1; then
    curl -fsSL "${LATEST_RELEASE_URL}" -o "${BINARY_NAME}"
elif command -v wget > /dev/null 2>&1; then
    wget -q "${LATEST_RELEASE_URL}" -O "${BINARY_NAME}"
else
    echo -e "${RED}Error: Neither curl nor wget found. Please install one of them and try again.${NC}"
    exit 1
fi

# Make binary executable
chmod +x "${BINARY_NAME}"

# Move binary to installation directory (requires sudo)
echo -e "${BLUE}Installing binary to ${INSTALL_DIR}...${NC}"
if [ -w "${INSTALL_DIR}" ]; then
    mv "${BINARY_NAME}" "${BINARY_PATH}"
else
    echo -e "${YELLOW}Requesting sudo access to install binary...${NC}"
    sudo mv "${BINARY_NAME}" "${BINARY_PATH}"
fi

# Initialize config if it doesn't exist
if [ ! -f "${CONFIG_DIR}/config.json" ]; then
    echo -e "${BLUE}Initializing configuration...${NC}"
    "${BINARY_PATH}" config init
fi

# Print success message
echo -e "\n${GREEN}✅ Coolify CLI installed successfully!${NC}"
echo -e "\n${BLUE}To get started:${NC}"
echo -e "1. Get your API token from ${YELLOW}https://app.coolify.io/security/api-tokens${NC}"
echo -e "2. Set your token: ${YELLOW}coolify-cli instances set token cloud <your-token>${NC}"
echo -e "3. Test the connection: ${YELLOW}coolify-cli config test${NC}"
echo -e "\nFor more information, run: ${YELLOW}coolify-cli --help${NC}"
