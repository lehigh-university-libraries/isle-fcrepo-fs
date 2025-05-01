#!/usr/bin/env bash

set -eou pipefail

if [ -f /app/ca.pem ]; then
    echo "Found /app/ca.pem, adding it to the trusted certificates..."
    cp /app/ca.pem /usr/local/share/ca-certificates/ca.pem
    update-ca-certificates
fi

exec gosu isle-fcrepo-fs /app/isle-fcrepo-fs
