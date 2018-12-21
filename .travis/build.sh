#!/bin/sh

set -eu

echo "Building...\n"

./configure
make distclean bin/pullcord

echo "\nBuild complete"

