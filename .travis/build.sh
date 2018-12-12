#!/bin/sh

set -eu

echo "Building...\n"

make distclean bin/pullcord

echo "\nBuild complete"

