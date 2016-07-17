#!/bin/sh

echo "Installing dependencies...\n"

go get github.com/stretchr/testify/assert \
	github.com/fitstar/falcore \
	github.com/dustin/randbo

