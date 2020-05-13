#! /usr/bin/env bash
######### Bash Boilerplate ##########
set -euo pipefail # strict mode
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
cd "$SCRIPT_DIR" # cd to this script's dir
######### End Bash Boilerplate ##########

# Add GOOS and GOARCH to our environment. (and other GO vars we don't need)
eval "$(go env)"

# Create /product/ dir to hold the output of this script.
readonly BUILD_PRODUCT_DIR="$SCRIPT_DIR/products/$GOOS/$GOARCH/"
mkdir -p "$BUILD_PRODUCT_DIR"

# build_dirs is an array of all the folders which contain a main.go
readarray -t build_dirs < <(find . -name 'main.go' -print0 | xargs --null dirname)
for dir in "${build_dirs[@]}"; do
  (
    cd "$dir"
    go build -o "$BUILD_PRODUCT_DIR"
  ) 
done

# compress the build products into an archive
cd "$BUILD_PRODUCT_DIR"
if [ "$GOOS" = "windows" ]; then
  zip note-go.$GOOS.$GOARCH.zip ./*
else
  tar -czvf note-go.$GOOS.$GOARCH.tar.gz ./*
fi;
