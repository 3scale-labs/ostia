#!/usr/bin/env bash

NAMESPACE=${1}

source "${BASH_SOURCE[0]%/*}/common.sh"

echo "Deploying Operator in ${NAMESPACE}"
oc create -f ../../../deploy/role.yaml -n ${NAMESPACE} > /dev/null
oc create -f ../../../deploy/service_account.yaml -n ${NAMESPACE} > /dev/null
oc create -f ../../../deploy/role_binding.yaml -n ${NAMESPACE} > /dev/null
cat ../../../deploy/operator.yaml | sed "s|REPLACE_IMAGE|${IMAGE:=}|g" | oc create -f- -n ${NAMESPACE} > /dev/null
wait_for_pod_ready name ostia-operator ${NAMESPACE}
echo -e "Operator deployed successfully in ${NAMESPACE}\n"
