#!/usr/bin/env bash
set -e

source ../common/common.sh

ns="$(generate_project_name)"
endpoint="${ns}.${OPENSHIFT_PUBLIC_HOSTNAME:=127.0.0.1}"

echo "Testing reconciliation in namespace ${ns}"
setup_project ${ns}

oc new-app -f ./endpoint.yaml --param HOSTNAME=${endpoint} -n ${ns} >/dev/null
wait_for_pod_ready app apicast ${ns}
echo "Proxy deployed successfully in ${ns}"

# Verifying expected HTTPS status
result=$(do_http_get ${endpoint}"/hello" 10)

if [[ ${result} != *"10 200"* ]]; then
    echo "Error. Unexpected status code response for reconciliation test "${result}
    exit 1
fi

oc patch api/endpoint -p '{"spec":{"endpoints":[{"name":"endpoint","path":"/bye"}]}}' --type merge -n ${ns} >/dev/null
wait_for_pod_ready app apicast ${ns}

result=$(do_http_get ${endpoint}"/bye" 10)

if [[ ${result} != *"10 200"* ]]; then
    echo "Error. Unexpected status code response for reconciliation test "${result}
    label_namespace ${ns} "ostia-test-result=failed"
    exit 1
fi

label_namespace ${ns} "ostia-test-result=success"
echo -e "Reconciliation test successful in ${ns}\n"
