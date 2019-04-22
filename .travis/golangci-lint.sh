#!/bin/sh

set -eu

IGNORE_ERRORS='TRUE'

go get -t -u -v github.com/golangci/golangci-lint/cmd/golangci-lint
go get -t -u -v ./...

echo "Running golangci-lint (this could take some time)..."
if [ "x${IGNORE_ERRORS}" = 'xTRUE' ]; then
	set +e
fi
golangci-lint run ./...
LINT_EXIT=$?

if [ ${LINT_EXIT} -eq 0 ]; then
	echo "No golangci-lint issues found."
fi

