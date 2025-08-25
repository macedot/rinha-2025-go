#!/bin/bash
docker run --rm \
    --name k6 \
    -p 5665:5665 \
    -e K6_WEB_DASHBOARD=true \
    -e K6_WEB_DASHBOARD_PORT=5665 \
    -e K6_WEB_DASHBOARD_PERIOD=1s \
    -e K6_WEB_DASHBOARD_OPEN=true \
    -e K6_WEB_DASHBOARD_EXPORT='report.html' \
    -v ./scripts:/scripts \
    -i grafana/k6 \
    run - --log-output=file=/scripts/k6.logs <rinha.js

if [[ ! -f "./scripts/partial-results.json" ]]; then
    exit 0
fi

T=$(date +%Y%m%dT%H%M%S)
mv -f ./scripts/k6.logs ./scripts/k6.logs.$T
mv -f ./scripts/partial-results.json ./scripts/partial-results.json.$T
