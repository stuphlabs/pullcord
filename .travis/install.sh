#!/bin/sh

echo "Installing dependencies...\n"

go get -v github.com/stretchr/testify/assert \
	github.com/fitstar/falcore \
	github.com/dustin/randbo \
	golang.org/x/crypto

echo "\nInstallation complete"

