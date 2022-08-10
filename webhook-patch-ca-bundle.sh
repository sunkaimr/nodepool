#!/bin/bash

ROOT=$(cd $(dirname $0)/../../; pwd)

set -o errexit
set -o nounset
set -o pipefail


export CA_BUNDLE=$(kubectl config view --raw --flatten -o json| jq -r '.clusters[].cluster."certificate-authority-data"' | base64 -d)

if command -v envsubst >/dev/null 2>&1; then
    envsubst
else
    sed -i "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
fi
