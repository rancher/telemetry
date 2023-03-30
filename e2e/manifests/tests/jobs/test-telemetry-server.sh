#!/usr/bin/env bash

TS_PORT=${TS_PORT:-8115}
TS_HOST=${TS_HOST:-"http://telemetry-server"}
TS_PATH=${TS_PATH:-"/admin/active?hours=48"}

# Check for existence of necessary credentials.
if [[ -z "$access_key" || -z "$secret_key" ]]; then
    echo >&2 "error: need \$access_key and \$secret_key environment variables set"
    exit 1
fi

# If run inside Kubernetes, assume Alpine and install dependencies.
if [ -n "${KUBERNETES_SERVICE_HOST}" ]; then
    apk add curl jq
fi

# Just try.
while true; do
    json=$(
        curl -fsSL -u "${access_key}":"${secret_key}" "${TS_HOST}":"${TS_PORT}${TS_PATH}"
    )
    if [[ "$?" -ne 0 || -z "$json" ]]; then
        echo >&2 "error: failed to fetch data"
        sleep 2
        continue
    fi

    result=$(
        echo $json | jq -r '.data[0].record.cluster.active'
    )
    if [[ "$?" -ne 0 || -z "$result" ]]; then
        echo >&2 "error: couldn't get active cluster count"
        exit 1
    fi

    if [ "$result" -ne 1 ]; then
        echo >&2 "error: wrong cluster count"
        exit 1
    fi

    echo "success"
    exit 0
done
