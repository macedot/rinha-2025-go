#!/bin/bash
docker run --rm \
    --name k6 \
    -p 5665:5665 \
    -e MAX_REQUESTS=550 \
    -e K6_WEB_DASHBOARD=true \
    -e K6_WEB_DASHBOARD_PORT=5665 \
    -e K6_WEB_DASHBOARD_PERIOD=1s \
    -e K6_WEB_DASHBOARD_OPEN=true \
    -e K6_WEB_DASHBOARD_EXPORT='report.html' \
    -v ./scripts:/scripts \
    -i grafana/k6 \
    run - --log-output=file=/scripts/k6.logs <rinha.js
