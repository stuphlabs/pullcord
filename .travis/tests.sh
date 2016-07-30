#!/bin/sh

COVERAGE_MINIMUM_PERCENTAGE="80"

echo "Running tests...\n"

UNIT_TEST_OUTPUT="`go test -v -cover ./...`"
UNIT_TEST_STATUS=$?

echo "$UNIT_TEST_OUTPUT"

if [ $UNIT_TEST_STATUS -ne 0 ]
then
	echo "\nUnit tests failed"
	exit 1
else
	echo "\nUnit tests passed"
fi

echo "$UNIT_TEST_OUTPUT" | mawk '\
$1 == "coverage:" {
	if ($2 < '$COVERAGE_MINIMUM_PERCENTAGE') {
		exit(1);
	}
}'
COVERAGE_TEST_STATUS=$?

if [ $COVERAGE_TEST_STATUS -ne 0 ]
then
	echo -n "\nCoverage test failed"
	echo " (less than $COVERAGE_MINIMUM_PERCENTAGE% coverage)"
	exit 1
else
	echo "\nCoverage test passed"
fi

echo "\n\nAll tests passed\n"

