#!/bin/bash

# This script downloads the latest Aptomi release from GitHub for your platform,
# installs server and client locally and creates default configs for them
# The list of releases is retrieved from: https://github.com/Aptomi/aptomi/releases

APTOMI_INSTALL_DIR="/usr/local/bin"
APTOMI_SERVER_CONFIG_DIR="/etc/aptomi"
APTOMI_CLIENT_CONFIG_DIR="$HOME/.aptomi"
APTOMI_DB_DIR="/var/lib/aptomi"

SCRIPT_NAME=`basename "$0"`
REPO_NAME='Aptomi/aptomi'

COLOR_GRAY='\033[0;37m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;92m'
COLOR_YELLOW='\033[0;93m'
COLOR_RESET='\033[0m'

DEBUG=${DEBUG:-no}
if [ "yes" == "$DEBUG" ]; then
    set -x
fi

function check_installed() {
    if ! [ -x "$(command -v $1)" ]; then
        log_err "$1 is not installed" >&2
        exit 1
    fi
}

function log() {
    echo -e "$COLOR_BLUE[$(date +"%F %T")] $SCRIPT_NAME $COLOR_RED|$COLOR_RESET" $@$COLOR_GRAY
}

function log_sub() {
    echo -e "$COLOR_BLUE[$(date +"%F %T")] $SCRIPT_NAME $COLOR_RED|$COLOR_RESET - " $@$COLOR_GRAY
}

function log_warn() {
    echo -e "$COLOR_BLUE[$(date +"%F %T")] $SCRIPT_NAME $COLOR_RED|$COLOR_RESET $COLOR_YELLOW - WARNING:" $@$COLOR_GRAY
}

function log_err() {
    echo -e "$COLOR_BLUE[$(date +"%F %T")] $SCRIPT_NAME $COLOR_RED| ERROR:" $@$COLOR_GRAY
}

function get_arch() {
    local ARCH=$(uname -m)
    case $ARCH in
        armv5*) ARCH="armv5";;
        armv6*) ARCH="armv6";;
        armv7*) ARCH="armv7";;
        aarch64) ARCH="arm64";;
        x86) ARCH="386";;
        x86_64) ARCH="amd64";;
        i686) ARCH="386";;
        i386) ARCH="386";;
    esac
    echo $ARCH
}

function get_os() {
    local OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')
    case "$OS" in
        mingw*) OS='windows';;
    esac
    echo $OS
}

function verify_supported_platform() {
    local ARCH=$1
    local OS=$2

    if [ -z "${ARCH}" ] || [ -z "${OS}" ]; then
        log_err "Unable to detect platform: architecture=$ARCH, os=$OS"
        exit 1
    fi
    log_sub "Detected: architecture=$COLOR_GREEN$ARCH$COLOR_RESET, os=$COLOR_GREEN$OS$COLOR_RESET"

    local supported="darwin_amd64\nlinux_386\nlinux_amd64"
    if ! echo "${supported}" | grep -q "${OS}_${ARCH}"; then
        log_err "No binaries available for ${OS}_${ARCH}"
        log_err "To build from source, go to https://github.com/$REPO_NAME#building-from-source"
        exit 1
    fi
}

function get_latest_release() {
    curl --silent "https://api.github.com/repos/$REPO_NAME/releases/latest" | # Get latest release from GitHub API
    grep '"tag_name":' |                                              # Filter out tag_name line
    sed -E 's/.*"([^"]+)".*/\1/'                                      # Parse out JSON value
}

function download_and_install_release() {
    local ARCH=$1
    local OS=$2
    local VERSION=$3

    if [ -z "${VERSION}" ]; then
        log_err "Unable to get the latest release from GitHub (https://api.github.com/repos/$REPO_NAME/releases/latest)"
        exit 1
    fi
    log_sub "Version: $COLOR_GREEN$VERSION$COLOR_RESET"

    local VERSIONWITHOUTV=${VERSION:1:${#VERSION}}

    local FILENAMEBINARY="aptomi_${VERSIONWITHOUTV}_${OS}_${ARCH}.tar.gz"

    local FILENAMECHECKSUMS="aptomi_${VERSIONWITHOUTV}_checksums.txt"
    local URL_BINARY="https://github.com/$REPO_NAME/releases/download/$VERSION/$FILENAMEBINARY"
    local URL_CHECKSUMS="https://github.com/$REPO_NAME/releases/download/$VERSION/$FILENAMECHECKSUMS"

    local TMP_DIR="$(mktemp -dt aptomi-install-files-XXXXXX)"
    local FILE_BINARY="$TMP_DIR/$FILENAMEBINARY"
    local FILE_CHECKSUMS="$TMP_DIR/$FILENAMECHECKSUMS"

    log_sub "Downloading: $URL_BINARY"
    curl -SsL "$URL_BINARY" -o "$FILE_BINARY"
    log_sub "Downloading: $URL_CHECKSUMS"
    curl -SsL "$URL_CHECKSUMS" -o "$FILE_CHECKSUMS"

    local sum=$(openssl sha1 -sha256 ${FILE_BINARY} | awk '{print $2 xxx}')
    local expected_line=$(cat ${FILE_CHECKSUMS} | grep ${FILENAMEBINARY})
    if [ "$sum  ${FILENAMEBINARY}" != "$expected_line" ]; then
        log_err "Failed to download '${FILENAMEBINARY}' or SHA sum does not match. Aborting install"
        exit 1
    fi

    install_binaries_from_archive $FILE_BINARY $FILENAMEBINARY
}

function run_as_root() {
  local CMD="$*"

  if [ $EUID -ne 0 ]; then
    CMD="sudo $CMD"
  fi

  $CMD
}

function install_binaries_from_archive() {
    local FILE_BINARY=$1
    local FILENAMEBINARY=$2
    local TMP_DIR="$(mktemp -dt aptomi-install-unpacked-XXXXXX)"

    # Unpack the archive
    log "Installing Aptomi"
    log_sub "Unpacking $FILENAMEBINARY"
    tar xf "$FILE_BINARY" -C "$TMP_DIR"

    # Cut .tar.gz to get the name of the directory inside the archive
    local DIRNAME="${FILENAMEBINARY%.*}"
    DIRNAME="${DIRNAME%.*}"
    UNPACKED_PATH="$TMP_DIR/$DIRNAME"

    if [ ! -f $UNPACKED_PATH/aptomi ]; then
        log_err "Binary 'aptomi' not found inside the release"
    fi

    if [ ! -f $UNPACKED_PATH/aptomictl ]; then
        log_err "Binary 'aptomictl' not found inside the release"
    fi

    # Install server & create config
    log_sub "Installing Aptomi server: $COLOR_GREEN${APTOMI_INSTALL_DIR}/aptomi"
    run_as_root cp "$UNPACKED_PATH/aptomi" "$APTOMI_INSTALL_DIR"

    # Install client & create config
    log_sub "Installing Aptomi client: $COLOR_GREEN${APTOMI_INSTALL_DIR}/aptomictl$COLOR_RESET"
    run_as_root cp "$UNPACKED_PATH/aptomictl" "$APTOMI_INSTALL_DIR"
}

function create_server_config() {
    local TMP_DIR="$(mktemp -dt aptomi-install-server-config-XXXXXX)"

    log_sub "Creating config for Aptomi server: $COLOR_GREEN${APTOMI_SERVER_CONFIG_DIR}/config.yaml$COLOR_RESET"
    if [ -f ${APTOMI_SERVER_CONFIG_DIR}/config.yaml ]; then
        log_warn "Config for Aptomi server already exists. Keeping existing config"
    else
        cat >${TMP_DIR}/config.yaml <<EOL
debug: true

api:
  host: 0.0.0.0
  port: 27866

db:
  connection: ${APTOMI_DB_DIR}/db.bolt

enforcer:
  interval: 5s

users:
  file:
    - ${APTOMI_SERVER_CONFIG_DIR}/users_builtin.yaml
    - ${APTOMI_SERVER_CONFIG_DIR}/users_example.yaml
  ldap-disabled:
    - host: localhost
      port: 10389
      basedn: "o=aptomiOrg"
      filter: "(&(objectClass=organizationalPerson))"
      filterbyname: "(&(objectClass=organizationalPerson)(cn=%s))"
      labeltoattributes:
        name: cn
        description: description
        global_ops: isglobalops
        is_operator: isoperator
        mail: mail
        team: team
        org: o
        short-description: role
        deactivated: deactivated
EOL
        run_as_root mkdir -p ${APTOMI_SERVER_CONFIG_DIR}
        run_as_root cp ${TMP_DIR}/config.yaml ${APTOMI_SERVER_CONFIG_DIR}/config.yaml
    fi

    log_sub "Creating built-in admin users for Aptomi server: $COLOR_GREEN${APTOMI_SERVER_CONFIG_DIR}/users_builtin.yaml$COLOR_RESET"
    if [ -f ${APTOMI_SERVER_CONFIG_DIR}/users_builtin.yaml ]; then
        log_warn "Built-in admin users for Aptomi server already exist. Keeping it"
    else
        cat >${TMP_DIR}/users_builtin.yaml <<EOL
- name: admin
  passwordhash: "\$2a\$10\$2eh0YI/gzj2UdxN8j52NseQW54BsZ5cUGhFstblR1D8UOGMUCwuMm"
  domainadmin: true
EOL
        run_as_root mkdir -p ${APTOMI_SERVER_CONFIG_DIR}
        run_as_root cp ${TMP_DIR}/users_builtin.yaml ${APTOMI_SERVER_CONFIG_DIR}/users_builtin.yaml
    fi

    log_sub "Creating example users for Aptomi server: $COLOR_GREEN${APTOMI_SERVER_CONFIG_DIR}/users_example.yaml$COLOR_RESET"
    if [ -f ${APTOMI_SERVER_CONFIG_DIR}/users_example.yaml ]; then
        log_warn "Example users for Aptomi server already exist. Keeping it"
    else
        run_as_root mkdir -p ${APTOMI_SERVER_CONFIG_DIR}
        run_as_root cp ${UNPACKED_PATH}/examples/twitter-analytics/_external/users.yaml ${APTOMI_SERVER_CONFIG_DIR}/users_example.yaml
    fi

    log_sub "Creating directory for Aptomi server database: $COLOR_GREEN${APTOMI_DB_DIR}$COLOR_RESET"
    if [ -d ${APTOMI_DB_DIR} ]; then
        log_warn "Directory for Aptomi server database already exists. Keeping it"
    else
        run_as_root mkdir -p ${APTOMI_DB_DIR}
    fi

    run_as_root chown ${USER:=$(/usr/bin/id -run)} ${APTOMI_DB_DIR}
}

function create_client_config() {
    local TMP_DIR="$(mktemp -dt aptomi-install-client-config-XXXXXX)"

    log_sub "Creating config for Aptomi client: $COLOR_GREEN${APTOMI_CLIENT_CONFIG_DIR}/config.yaml$COLOR_RESET"

    if [ -f ${APTOMI_CLIENT_CONFIG_DIR}/config.yaml ]; then
        log_warn "Config for Aptomi client already exists. Keeping existing config"
    else
        cat >${TMP_DIR}/config.yaml <<EOL
debug: true

api:
  host: 127.0.0.1
  port: 27866

auth:
  username: admin
EOL
        run_as_root mkdir -p ${APTOMI_CLIENT_CONFIG_DIR}
        run_as_root cp ${TMP_DIR}/config.yaml ${APTOMI_CLIENT_CONFIG_DIR}/config.yaml
    fi
}

function copy_examples() {
    log_sub "Copying examples into $COLOR_GREEN${APTOMI_CLIENT_CONFIG_DIR}/examples$COLOR_RESET"

    run_as_root mkdir -p ${APTOMI_CLIENT_CONFIG_DIR}
    run_as_root cp -R ${UNPACKED_PATH}/examples ${APTOMI_CLIENT_CONFIG_DIR}/
}

function test_aptomi() {
    # Verify that Aptomi server is in path
    local APTOMI=`which aptomi`
    if [ "$APTOMI" == "$APTOMI_INSTALL_DIR/aptomi" ]; then
        log_sub "Aptomi server: ${COLOR_GREEN}OK${COLOR_RESET} (which aptomi -> $APTOMI)"
    else
        log_warn "Aptomi server: 'which aptomi' returned '$APTOMI', but expected '$APTOMI_INSTALL_DIR/aptomi'"
        exit 1
    fi

    # Verify that Aptomi client is in path
    local APTOMICTL=`which aptomictl`
    if [ "$APTOMICTL" == "$APTOMI_INSTALL_DIR/aptomictl" ]; then
        log_sub "Aptomi client: ${COLOR_GREEN}OK${COLOR_RESET} (which aptomictl -> $APTOMICTL)"
    else
        log_warn "Aptomi client: 'which aptomictl' returned '$APTOMICTL', but expected '$APTOMI_INSTALL_DIR/aptomictl'"
        exit 1
    fi

    # Run 'aptomi version' and remove leading whitespaces
    local SERVER_VERSION_OUTPUT=$(aptomi version 2>/dev/null | grep 'Git Version')
    SERVER_VERSION_OUTPUT="$(echo -e "${SERVER_VERSION_OUTPUT}" | sed -e 's/^[[:space:]]*//')"
    if [ ! -z "${SERVER_VERSION_OUTPUT}" ]; then
        log_sub "Running 'aptomi version': ${COLOR_GREEN}OK${COLOR_RESET}"
    else
        log_err "Failed to parse output of 'aptomi version'"
        exit 1
    fi

    local TMP_DIR="$(mktemp -dt aptomi-install-server-runtime-XXXXXX)"

    # Start Aptomi server
    local SERVER_RUNNING_PRIOR=`ps | grep aptomi | grep server`
    if [ ! -z "${SERVER_RUNNING_PRIOR}" ]; then
        log_err "Aptomi server already running. Can't run another instance for testing (may want to use 'killall aptomi')"
        exit 1
    fi

    aptomi server &>${TMP_DIR}/server.log &
    local SERVER_PID=$!
    log_sub "Starting 'aptomi server' for testing (PID: ${SERVER_PID})"
    sleep 1
    local SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}"`
    if [ -z "${SERVER_RUNNING}" ]; then
        log_err "Aptomi server failed to start"
        exit 1
    fi

    # Run client to show the version
    local CLIENT_VERSION_OUTPUT=$(aptomictl version 2>/dev/null | grep 'Git Version' | wc -l)
    CLIENT_VERSION_OUTPUT="$(echo -e "${CLIENT_VERSION_OUTPUT}" | sed -e 's/^[[:space:]]*//')"
    if [ $CLIENT_VERSION_OUTPUT -eq 2 ]; then
        log_sub "Running 'aptomictl version': ${COLOR_GREEN}OK${COLOR_RESET}"
    else
        log_err "Failed to parse output of 'aptomictl version'"
        exit 1
    fi

    # Run client to show the policy
    local CLIENT_POLICY_SHOW_OUTPUT=$(aptomictl policy show 2>/dev/null | grep 'Policy Version')
    if [ ! -z "${CLIENT_POLICY_SHOW_OUTPUT}" ]; then
        log_sub "Running 'aptomictl policy show': ${COLOR_GREEN}OK${COLOR_RESET}"
    else
        log_err "Failed to parse output of 'aptomictl policy show'"
        exit 1
    fi

    log_sub "Stopping aptomi server (PID: ${SERVER_PID})"
    kill ${SERVER_PID} 2>/dev/null
    wait ${SERVER_PID} 2>/dev/null
}

# Initial checks
log "Starting Aptomi install"
check_installed 'curl'
check_installed 'grep'
check_installed 'sed'
check_installed 'awk'
check_installed 'openssl'
check_installed 'tar'
check_installed 'cp'
check_installed 'mkdir'
check_installed 'cat'
check_installed 'ps'

# Detect platform and verify that it's supported
log "Detecting platform"
ARCH=$(get_arch)
OS=$(get_os)
verify_supported_platform $ARCH $OS

# Download the latest release from GitHub and install it
log "Installing the latest release from GitHub"
VERSION=$(get_latest_release)
download_and_install_release $ARCH $OS $VERSION

# Set up server and client locally on the same host
create_server_config
create_client_config
copy_examples

# Test Aptomi
log "Testing Aptomi"
test_aptomi

# Done
log "Done"
