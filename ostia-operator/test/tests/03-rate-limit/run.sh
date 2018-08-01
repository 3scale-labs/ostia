#!/usr/bin/env bash
set -e

source ../common/common.sh

ns="$(generate_project_name)"
host="${ns}.${OPENSHIFT_PUBLIC_HOSTNAME:=127.0.0.1}"

fail() {
    echo "Error. Unexpected status code response for rate limit test "${1}
    label_namespace ${ns} "ostia-test-result=failed"
    exit 1
}

echo "Testing Fixed Rate Limiting in namespace ${ns}"
setup_project ${ns}

echo "Creating Test Endpoint in namespace ${ns} "
oc process -f ./fixed_rate_global.yaml --param HOSTNAME=${host}  | oc create -f - -n ${ns} > /dev/null
wait_for_pod_ready app apicast ${ns}
echo "Proxy deployed successfully in ${ns}"

echo "Testing fixed rate limiting in namespace ${ns}"
fixed_rate=$(do_http_get ${host}"/" 20)
if [[ ${fixed_rate} != *"10 200"*"10 429"* ]]; then
    fail ${fixed_rate}
fi

fixed_rate_exec=$(do_http_get_in_pod "apicast-endpoints" 20 "${ns}")
if [[ ${fixed_rate_exec} != *"20 429"* ]]; then
    fail ${fixed_rate_exec}
fi
echo "Fixed rate limiting in namespace ${ns} success"

echo "Testing fixed rate IP based limiting in namespace ${ns}"
oc apply -f ./fixed_rate_ip.yaml -n ${ns} &> /dev/null
wait_for_pod_ready app apicast ${ns}
fixed_rate_ip=$(do_http_get ${host}"/" 20)
if [[ ${fixed_rate_ip} != *"10 200"*"10 429"* ]]; then
    fail ${fixed_rate_ip}
fi

fixed_rate_exec_ip=$(do_http_get_in_pod "apicast-endpoints" 20 ${ns})
if [[ ${fixed_rate_exec_ip} != *"10 200"*"10 429"* ]]; then
    fail ${fixed_rate_exec_ip}
fi
echo "IP based rate limiting in namespace ${ns} success"

label_namespace ${ns} "ostia-test-result=success"
echo -e "Endpoint test successful in ${ns}\n"