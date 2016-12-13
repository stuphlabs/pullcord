#!/bin/sh

echo "Installing dependencies...\n"

go get -v \
	github.com/dustin/randbo \
	github.com/fitstar/falcore \
	github.com/proidiot/gone \
	github.com/stretchr/testify/assert \
	golang.org/x/crypto \
	golang.org/x/net/html

echo "\nInstallation complete"

