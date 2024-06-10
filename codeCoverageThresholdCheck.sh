#!/usr/bin/env bash
set -e

pushd ./src/ || exit

THRESHOLD=15
COVERAGE=$(go tool cover -func ./coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')

echo "Code coverage is $COVERAGE%, and the threshold is $THRESHOLD%"

if (( $(echo "$COVERAGE $THRESHOLD" | awk '{print ($1 < $2)}') )); then
  exit 1
fi

exit 0
