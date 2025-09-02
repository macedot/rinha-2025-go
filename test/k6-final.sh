#!/bin/bash
K6_WEB_DASHBOARD_OPEN=false \
K6_WEB_DASHBOARD=true \
K6_WEB_DASHBOARD_PORT=5665 \
K6_WEB_DASHBOARD_PERIOD=1s \
K6_WEB_DASHBOARD_EXPORT='report.html' \
k6 run ./rinha-final.js

if [[ ! -f "./summary/final-results.json" ]]; then
    exit 0
fi

T=$(date +%Y%m%dT%H%M%S)
mv -f ./summary/final-results.json ./summary/final-results.json.$T
