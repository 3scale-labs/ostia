#!/usr/bin/env bash

DIRECTORY=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Creates Ostia Operator
# Arg: Namespace
deploy_operator() {
 ${DIRECTORY}/deploy_operator.sh ${1}
}

# Creates Echo API
# Arg: Namespace
deploy_echo_api() {
 ${DIRECTORY}/deploy_echo_api.sh ${1}
}

# Adds labels to namespace for filtering
# Expects two args: Namespace, and label in format "x=y"
label_namespace() {
 oc label --overwrite=true namespace ${1} ${2}  > /dev/null
}

# Creates a project and default requirements
# Arg: Namespace
setup_project() {
 oc new-project ${1} > /dev/null
 label_namespace ${1} "ostia-test=true ostia-test-result=unknown"
 deploy_operator ${1}
 deploy_echo_api ${1}
}

# Check if namespace exists
# Arg: Project name
project_exists () {
  oc get project $1 &> /dev/null
}

# Delete namespace if it exists
# Arg: Project name
delete_project () {
  if project_exists $1; then
    echo "Cleaning up project" ${1}
    oc delete project ${1} --now=true
  fi
}

# Returns a Pod name given labels to filter on
# Expects three args: Key and value filter for Pod metadata labels, Pod namespace
get_pod_name() {
 oc get pod -o jsonpath='{.items[?(@.metadata.labels.'${1}'=="'${2}'")].metadata.name}' -n ${3}
}

# Give Pods two minutes to hit ready status before error
# Expects three args: Key and value filter for Pod metadata labels, Pod namespace
wait_for_pod_ready () {
 sleep 5 # Wait for Pod to be created before checking its readiness
 retries="12"
 while [[ "retries" -gt "0" ]]
 do
    if [ "$(oc get pod -o jsonpath='{.items[?(@.metadata.labels.'${1}'=="'${2}'")].status.containerStatuses[0].ready}' -n ${3} )" == "true" ];
    then
        return
    else
        let "retries--"
        sleep 10
    fi
 done
 echo "Pod was not ready in time"
 exit 1
}

# Prints to stdout the number of each unique status codes
# Expects two args: Endpoint to make HTTP request against, number of requests to make
do_http_get() {
 for i in $(seq 1 ${2}); do
   curl -k -s -o /dev/null -w "%{http_code}\n" ${1}
 done | uniq -c

}

# Executes a provided command within a specific Pod
# Expects three args: Command to run, Pod name, Pod namespace
run_cmd_in_pod() {
  oc exec -n ${3} ${2} -- bash -c "${1}"
}

# Functionally equal to do_http_get but calling will execute within an apicast Pod in provided namespace
# This function should be used when needed to simulate different source ip for testing
# Relies on setting the host header
# Expects three args: Host to make HTTP request against, number of requests to make, namespace
do_http_get_in_pod() {
 pod_to_exec=$(get_pod_name app apicast ${3})
 cmd="for i in {1..${2}}; do curl -k -s -o /dev/null -w '%{http_code}\n' http://localhost:8080 -H 'HOST: ${1}'; done | uniq -c"
 run_cmd_in_pod "${cmd}" ${pod_to_exec} ${3}
}

# Returns a string containing calling test directory appended with a random string
generate_project_name() {
 ns=$(basename $(pwd))
 ns+="-"$(hexdump -n 8 -e '4/4 "%08X" 1 "\n"' /dev/random |  tr -dc '[:alnum:]\n\r' | tr '[:upper:]' '[:lower:]')
 echo ${ns}
}
