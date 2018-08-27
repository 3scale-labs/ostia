#!/usr/bin/env bash
NAMESPACE=${1}

source "${BASH_SOURCE[0]%/*}/common.sh"

echo "Deploying Operator in ${NAMESPACE}"
oc create -f ../../../deploy/rbac.yaml -n ${NAMESPACE} > /dev/null
oc create -f ../../../deploy/operator.yaml -n ${NAMESPACE} > /dev/null
wait_for_pod_ready name ostia-operator ${NAMESPACE}
echo -e "Operator deployed successfully in ${NAMESPACE}\n"
