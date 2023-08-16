#! /usr/bin/env bash
#
# Copyright 2020 Blues Inc.  All rights reserved.
# Use of this source code is governed by licenses granted by the
# copyright holder including that found in the LICENSE file.
#
######### Bash Boilerplate ##########
set -euo pipefail # strict mode
readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$SCRIPT_DIR" # cd to this script's dir
######### End Bash Boilerplate ##########

#
# note-go build.sh
#
# This script builds all the note-go executables (note, notecard, notehub) by
# looking for any folder containing a main.go and running `go build`.
#
# Parameters: Set $GOOS and $GOARCH to cross compile for different platforms.
#
# Output: Executables are saved in "./build/$GOOS/$GOARCH/"
#

# Add GOOS and GOARCH to our environment. (and other GO vars we don't need)
eval "$(go env)"

readonly BUILD_EXE_DIR="$SCRIPT_DIR/build/$GOOS/$GOARCH/"
mkdir -p "$BUILD_EXE_DIR"

# Build each executable binary
# build_dirs is an array of all the folders which contain a main.go
IFS=$'\r\n' GLOBIGNORE='*' command eval  'build_dirs=($(find . -name main.go -print0 | xargs -0n1 dirname))'
for dir in "${build_dirs[@]}"; do
  (
    cd "$dir"
    go build -o "$BUILD_EXE_DIR"
  ) 
done
