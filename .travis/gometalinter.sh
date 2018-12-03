#!/bin/sh

set -eu

IGNORE_ERRORS='TRUE'

go get -t -u -v github.com/alecthomas/gometalinter

gometalinter -i -u

go get -t -u -v ./...

echo "Running gometalinter (this could take some time)..."
if [ "x${IGNORE_ERRORS}" = 'xTRUE' ]; then
	set +e
fi
gometalinter \
	--enable-all \
	--deadline=600s \
	--enable-gc \
	--cyclo-over=15 \
	./...
LINT_EXIT=$?

if [ ${LINT_EXIT} -eq 0 ]; then
	echo "No gometalinter issues found."
fi

