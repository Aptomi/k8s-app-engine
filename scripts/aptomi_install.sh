#!/bin/bash

# This script downloads the latest Aptomi release from GitHub for your platform,
# installs server and client locally and creates default configs for them
# The list of releases is retrieved from: https://github.com/Aptomi/aptomi/releases

APTOMI_INSTALL_DIR="/usr/local/bin"
APTOMI_SERVER_CONFIG_DIR="/etc/aptomi"
APTOMI_CLIENT_CONFIG_DIR="$HOME/.aptomi"
APTOMI_INSTALL_CACHE="$HOME/.aptomi-install-cache"
APTOMI_DB_DIR="/var/lib/aptomi"
REPO_NAME='Aptomi/aptomi'
SCRIPT_NAME=`basename "$0"`
UPLOAD_EXAMPLE=0
CLIENT_ONLY=0
SERVER_PID=""

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

trap "script_done" INT TERM EXIT

set -eou pipefail

function script_done() {
    local CODE=$?

    if [ ! -z "${SERVER_PID}" ]; then
        log_sub "Stopping aptomi server (PID: ${SERVER_PID})"
        kill ${SERVER_PID} >/dev/null 2>&1
        while ps -p ${SERVER_PID} >/dev/null; do sleep 1; done
    fi

    if [ ! $CODE -eq 0 ]; then
        log_err "Script failed"
    fi

    exit $CODE
}

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

    local FILE_BINARY="$APTOMI_INSTALL_CACHE/$FILENAMEBINARY"
    local FILE_CHECKSUMS="$APTOMI_INSTALL_CACHE/$FILENAMECHECKSUMS"

    mkdir -p $APTOMI_INSTALL_CACHE
    if [ ! -f $FILE_BINARY ]; then
        log_sub "Downloading: $URL_BINARY"
        curl -SsL "$URL_BINARY" -o "$FILE_BINARY"
    else
        log_sub "Already downloaded. Using from cache: $FILE_BINARY"
    fi

    # Never cache checksum, it'll allow us to verify cached binary
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
    CMD="$*"

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

    # Install server & create config (only if we are not in CLIENT_ONLY mode)
    if [ $CLIENT_ONLY -eq 0 ]; then
        if [ ! -f $UNPACKED_PATH/aptomi ]; then
            log_err "Binary 'aptomi' not found inside the release"
        fi

        log_sub "Installing Aptomi server: $COLOR_GREEN${APTOMI_INSTALL_DIR}/aptomi"
        run_as_root cp "$UNPACKED_PATH/aptomi" "$APTOMI_INSTALL_DIR"
    fi

    if [ ! -f $UNPACKED_PATH/aptomictl ]; then
        log_err "Binary 'aptomictl' not found inside the release"
    fi

    # Install client & create config
    log_sub "Installing Aptomi client: $COLOR_GREEN${APTOMI_INSTALL_DIR}/aptomictl$COLOR_RESET"
    run_as_root cp "$UNPACKED_PATH/aptomictl" "$APTOMI_INSTALL_DIR"
}

function create_server_config() {
    # Skip server installation if we are in CLIENT_ONLY mode
    if [ $CLIENT_ONLY -eq 1 ]; then
        return 0
    fi

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
  noop: false
  interval: 2s

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

        # If we are in example mode, disable enforcer
        if [ $UPLOAD_EXAMPLE -eq 1 ]; then
            log_sub "Disabling enforcer (as we are running in example mode)"

            if [ $EUID -ne 0 ]; then
                sudo sed -i.bak -e "s/noop: false/noop: true/" ${APTOMI_SERVER_CONFIG_DIR}/config.yaml
            else
                sed -i.bak -e "s/noop: false/noop: true/" ${APTOMI_SERVER_CONFIG_DIR}/config.yaml
            fi
        fi
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
        mkdir -p ${APTOMI_CLIENT_CONFIG_DIR}
        cp ${TMP_DIR}/config.yaml ${APTOMI_CLIENT_CONFIG_DIR}/config.yaml
    fi
}

function copy_examples() {
    log_sub "Copying examples into $COLOR_GREEN${APTOMI_CLIENT_CONFIG_DIR}/examples$COLOR_RESET"

    mkdir -p ${APTOMI_CLIENT_CONFIG_DIR}
    cp -R ${UNPACKED_PATH}/examples ${APTOMI_CLIENT_CONFIG_DIR}/
}

function test_aptomi_server_in_path() {
    # Verify that Aptomi server is in path
    local APTOMI=`which aptomi`
    if [ "$APTOMI" == "$APTOMI_INSTALL_DIR/aptomi" ]; then
        log_sub "Aptomi server: ${COLOR_GREEN}OK${COLOR_RESET} (which aptomi -> $APTOMI)"
    else
        log_warn "Aptomi server: 'which aptomi' returned '$APTOMI', but expected '$APTOMI_INSTALL_DIR/aptomi'"
        exit 1
    fi
}

function test_aptomi_client_in_path() {
    # Verify that Aptomi client is in path
    local APTOMICTL=`which aptomictl`
    if [ "$APTOMICTL" == "$APTOMI_INSTALL_DIR/aptomictl" ]; then
        log_sub "Aptomi client: ${COLOR_GREEN}OK${COLOR_RESET} (which aptomictl -> $APTOMICTL)"
    else
        log_warn "Aptomi client: 'which aptomictl' returned '$APTOMICTL', but expected '$APTOMI_INSTALL_DIR/aptomictl'"
        exit 1
    fi
}

function test_aptomi_server_version_success() {
    # Run 'aptomi version' and remove leading whitespaces
    local SERVER_VERSION_OUTPUT
    SERVER_VERSION_OUTPUT=$(aptomi version 2>/dev/null)
    if [ $? -eq 0 ]; then
        log_sub "Running 'aptomi version': ${COLOR_GREEN}OK${COLOR_RESET}"
    else
        log_err "Failed to execute 'aptomi version'"
        log_err $SERVER_VERSION_OUTPUT
        exit 1
    fi
}

function start_aptomi_server() {
    # Start Aptomi server
    local SERVER_RUNNING_PRIOR=`ps | grep aptomi | grep server`
    if [ ! -z "${SERVER_RUNNING_PRIOR}" ]; then
        log_err "Aptomi server already running. Can't run another instance for testing (may want to use 'killall aptomi')"
        exit 1
    fi

    aptomi server &>/dev/null &
    SERVER_PID=$!
    log_sub "Starting 'aptomi server' for testing (PID: ${SERVER_PID})"
    sleep 2
    local SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}"`
    if [ -z "${SERVER_RUNNING}" ]; then
        log_err "Aptomi server failed to start"
        exit 1
    fi
}

function test_aptomi_client_version_success() {
    # Run client to show the version
    local CLIENT_VERSION_OUTPUT
    if [ $CLIENT_ONLY -eq 0 ]; then
        CLIENT_VERSION_OUTPUT=$(aptomictl version 2>/dev/null)
    else
        CLIENT_VERSION_OUTPUT=$(aptomictl version --client 2>/dev/null)
    fi
    if [ $? -eq 0 ]; then
        log_sub "Running 'aptomictl version': ${COLOR_GREEN}OK${COLOR_RESET}"
    else
        log_err "Failed to execute 'aptomictl version'"
        log_err $CLIENT_VERSION_OUTPUT
        exit 1
    fi
}

function test_aptomi_client_show_policy_success() {
    # Run client to show the policy
    local CLIENT_POLICY_SHOW_OUTPUT
    CLIENT_POLICY_SHOW_OUTPUT=$(aptomictl policy show 2>/dev/null)
    if [ $? -eq 0 ]; then
        log_sub "Running 'aptomictl policy show': ${COLOR_GREEN}OK${COLOR_RESET}"
    else
        log_err "Failed to execute 'aptomictl policy show'"
        log_err $CLIENT_POLICY_SHOW_OUTPUT
        exit 1
    fi
}

function test_aptomi() {
    # Test that installed aptomi binaries are in PATH
    if [ $CLIENT_ONLY -eq 0 ]; then
        test_aptomi_server_in_path
        test_aptomi_server_version_success
        start_aptomi_server
    fi

    test_aptomi_client_in_path
    test_aptomi_client_version_success

    if [ $CLIENT_ONLY -eq 0 ]; then
        test_aptomi_client_show_policy_success

        # Upload example, if needed
        if [ $UPLOAD_EXAMPLE -eq 1 ]; then
            upload_example
        fi
    fi
}

function example_run_line() {
    local CMD="$*"

    # Run command
    log_sub "${CMD}"
    ($CMD 1>/dev/null 2>&1)
}

function upload_example() {
    log_sub "Uploading example"
    example_run_line "aptomictl policy apply --wait --username Sam -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/Sam"
    example_run_line "aptomictl policy apply --wait --username Sam -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/Sam/clusters.yaml.template"
    example_run_line "aptomictl policy apply --wait --username Frank -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/Frank"
    example_run_line "aptomictl policy apply --wait --username John -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/John"
    example_run_line "aptomictl policy apply --wait --username John -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/john-prod-ts.yaml"
    example_run_line "aptomictl policy apply --wait --username Alice -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/alice-stage-ts.yaml"
    example_run_line "aptomictl policy apply --wait --username Bob -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/bob-stage-ts.yaml"
    example_run_line "aptomictl policy apply --wait --username Carol -f ${APTOMI_CLIENT_CONFIG_DIR}/examples/twitter-analytics/policy/carol-stage-ts.yaml"
}

function help() {
    echo "This script installs Aptomi. Accepted CLI arguments are:"
    echo -e "\t--help: prints this help"
    echo -e "\t--with-example: imports example after installing and disables enforcer"
}

# Parsing input arguments (if any)
export INPUT_ARGUMENTS="$@"
while [[ $# -gt 0 ]]; do
  case $1 in
    '--client-only')
        CLIENT_ONLY=1
        ;;
    '--with-example')
        UPLOAD_EXAMPLE=1
        ;;
    '--help')
        help
        exit 0
        ;;
  esac
  shift
done

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
