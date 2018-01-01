#!/bin/sh

set -eu

echo "Running tests...\n"

make clean test

echo "\n\nAll tests passed\n"

