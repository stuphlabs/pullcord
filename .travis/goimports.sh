#!/bin/sh

set -eu

go get golang.org/x/tools/cmd/goimports

find . -name '*.go' -exec goimports -l {} \; | awk '
/./ {
	if (!found) {
		found = 1
		print "The following files have gofmt issues:"
	}
	print
}
END {
	if (found) {
		exit 1
	} else {
		print "No gofmt issues."
	}
}'

