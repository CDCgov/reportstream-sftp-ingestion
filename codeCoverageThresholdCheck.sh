#!/usr/bin/env bash
set -e

pushd ./src/ || exit

THRESHOLD=30
COVERAGE=$(go tool cover -func ./coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')

echo "Code coverage is $COVERAGE%, and the threshold is $THRESHOLD%"

if (( $(echo "$COVERAGE $THRESHOLD" | awk '{print ($1 < $2)}') )); then
  echo "Code coverage is not high enough"
  exit 1
fi

echo "Code coverage passes the threshold"
exit 0
