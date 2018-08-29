#!/usr/bin/env bash
NAMESPACE=${1}

source "${BASH_SOURCE[0]%/*}/common.sh"

echo "Deploying echo-api in ${NAMESPACE}"
oc new-app -f ../common/echo_api.yaml -n ${NAMESPACE} >/dev/null
wait_for_pod_ready app echo-api ${NAMESPACE}
echo "echo-api deployed in ${NAMESPACE}"