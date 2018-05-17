#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

DOCKER_REPO_ROOT="/go/src/github.com/3scale/ostia"
IMAGE=${IMAGE:-"gcr.io/coreos-k8s-scale-testing/codegen:1.9.3"}

source_volume() {
    volume=$(docker volume create)
    container=$(docker create --rm --volume "${volume}:${DOCKER_REPO_ROOT}" "${IMAGE}")
    docker cp "$PWD" "${container}:${DOCKER_REPO_ROOT}"
    docker rm "${container}" >/dev/null
    echo "$volume"
}

volume=$(source_volume)

container=$(
  docker create \
  -v "${volume}:${DOCKER_REPO_ROOT}" \
  -w "${DOCKER_REPO_ROOT}/ostia-operator" \
  "${IMAGE}" \
  "/go/src/k8s.io/code-generator/generate-groups.sh"  \
  "deepcopy" \
  "github.com/3scale/ostia/ostia-operator/pkg/generated" \
  "github.com/3scale/ostia/ostia-operator/pkg/apis" \
  "ostia:v1alpha1" \
  --go-header-file "./tmp/codegen/boilerplate.go.txt" \
  $@
)

docker start "${container}" > /dev/null
docker wait "${container}" > /dev/null

docker cp  "${container}:${DOCKER_REPO_ROOT}/ostia-operator/pkg" .

docker rm --volumes "${container}" > /dev/null
docker volume rm "${volume}" > /dev/null