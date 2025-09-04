#!/bin/bash
K6_WEB_DASHBOARD_OPEN=false \
K6_WEB_DASHBOARD=true \
K6_WEB_DASHBOARD_PORT=5665 \
K6_WEB_DASHBOARD_PERIOD=1s \
K6_WEB_DASHBOARD_EXPORT='report.html' \
k6 run ./rinha.js

if [[ ! -f "./summary/partial-results.json" ]]; then
    exit 0
fi

T=$(date +%Y%m%dT%H%M%S)
mv -f ./summary/partial-results.json ./summary/partial-results.json.$T
