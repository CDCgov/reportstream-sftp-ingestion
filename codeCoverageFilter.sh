#!/usr/bin/env bash
set -e

pushd ./src/ || exit

EXCLUDE_FILES=(
  github.com/CDCgov/reportstream-sftp-ingestion/cmd/main.go
  github.com/CDCgov/reportstream-sftp-ingestion/storage/azure.go
)

for exclusion in "${EXCLUDE_FILES[@]}"
do
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "/${exclusion//\//\\/}/d" ./coverage.out
  else
    sed -i "/${exclusion//\//\\/}/d" ./coverage.out
  fi

done