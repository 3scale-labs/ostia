#!/usr/bin/env bash

set -m
EXIT_CODE=0
dir=$(pwd)

# Check that the user running the tests can create resources passed as params. Exit if fails
# Expects two args: Verb (CRUD) and resource type.
test_perms() {
    [[ "$(oc auth can-i ${1} ${2})" == "yes" ]] && return
    echo "Current user cannot ${1} ${2}"
    exit 1
}

# Iterates over the background jobs and reports any failed test
process_jobs() {
    for job in $(jobs -p); do
        if ! wait ${job} ; then
            EXIT_CODE=1;
        fi
    done
}

# Setup function. Determines the server IP if not provided and creates CRD
init() {
    if $(oc get pods > /dev/null 2>&1); then
        HOSTNAME_FROM_OC=$(basename $(oc whoami --show-server=true) | cut -f1 -d":")
        [[ ${HOSTNAME_FROM_OC} == *.nip.io ]] || HOSTNAME_FROM_OC+=.nip.io
    fi

    if [ -z "$OPENSHIFT_PUBLIC_HOSTNAME" ]; then
        export OPENSHIFT_PUBLIC_HOSTNAME=${HOSTNAME_FROM_OC}
    fi

    test_perms create CustomResourceDefinition
    oc create -f ../deploy/crd.yaml &> /dev/null || true
    test_perms create RoleBinding
}


init

# Run the tests in parallel in background
for f in ./tests/*-*/*.sh; do
  echo "Launching: $f"
  cd ${f%/*}; ./$(basename $f) &
  cd $dir
done

process_jobs


# Prints a list of failed tests
if [[ ${EXIT_CODE} != 0 ]]; then
    echo -e "The following tests failed:"
    oc get projects -o custom-columns=:.metadata.name -l  'ostia-test-result in (failed, unknown)' --no-headers=true
fi

exit "$EXIT_CODE"