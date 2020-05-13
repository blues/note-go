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
# note-go package.sh
#
# This script packages all the note-go executables (note, notecard, notehub)
# into an archive named note-go.$GOOS.$GOARCH.tar.gz or zip in the case of
# $GOOS=windows.
#
# Parameters: This script uses $GOOS and $GOARCH determine where to look for the
#             executables.
#
# Output: Archives are saved in "./build/packages/"
#

# Add GOOS and GOARCH to our environment. (and other GO vars we don't need)
eval "$(go env)"

readonly BUILD_EXE_DIR="$SCRIPT_DIR/build/$GOOS/$GOARCH/"
mkdir -p "$BUILD_EXE_DIR"
readonly BUILD_PACKAGE_DIR="$SCRIPT_DIR/build/packages/"
mkdir -p "$BUILD_PACKAGE_DIR"

# compress the build products into an archive
cd "$BUILD_EXE_DIR"
if [ "$GOOS" = "windows" ]; then
  zip "$BUILD_PACKAGE_DIR/note-go.$GOOS.$GOARCH.zip" ./*
else
  tar -czvf "$BUILD_PACKAGE_DIR/note-go.$GOOS.$GOARCH.tar.gz" ./*
fi;
