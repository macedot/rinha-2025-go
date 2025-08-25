#!/bin/bash
for i in $(seq 1 10); do
  ./k6.sh
  if [[ ! -f "./scripts/partial-results.json" ]]; then
    exit 0
  fi
  echo $i
  mv -f ./scripts/k6.logs ./scripts/k6.logs.$i
  mv -f ./scripts/partial-results.json ./scripts/partial-results.json.$i
done
