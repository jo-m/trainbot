#!/bin/bash

REPO_URL=https://github.com/phoboslab/qoi
REPO_CLONE_DIR=qoi_upstream

# cd to current directory
cd "$(dirname "$0")"

# clean up and clone
rm -rf "$REPO_CLONE_DIR"
git clone "$REPO_URL" "$REPO_CLONE_DIR"

cp "$REPO_CLONE_DIR/LICENSE" ../../LICENSE_QOI
cp "$REPO_CLONE_DIR/qoi.h" .

echo "// Package cqoi is a CGo wrapper for the QOI image format from ${REPO_URL}.
// The current files are from upstream commit $(cd ${REPO_CLONE_DIR} && git rev-parse HEAD).
package cqoi" > pkg.go
