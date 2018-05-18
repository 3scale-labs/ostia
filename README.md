# "Ostia" Project main repository

Ostia aims to be an openshift native API Management software.

This is a not supported/official redhat product.

## Deployment

* To learn how to deploy openshift locally:

<https://github.com/openshift/origin/blob/master/docs/cluster_up_down.md>

* Clone the repo:

```
git clone https://github.com/3scale/ostia.git
cd ostia
```

* Create the required objects, first the Custom Resource Definition:

```
oc create -f ostia-operator/deploy/crd.yaml
```

* Add cluster-admin perms to the ostia default user: (needs fix)

```
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:ostia:default
```

* Deploy the operator, you can create a new project for if:

```
oc new-project ostia
oc create -f ostia-operator/deploy/operator.yml
```

* Deploy the example Custom Resource in any namespace:

```
oc new-project myhelloapi
oc create -f ostia-operator/deploy/cr.yaml
```

## Build

* Prerequisites:
  * Go >= 1.10
  * Go Dep (<https://github.com/golang/dep>)
  * Operator-SDK v0.0.5 (<https://github.com/operator-framework/operator-sdk>)
  * Docker

* Installing Go:

Please refer to the official docs: <https://golang.org/doc/install>

* Installing Go Dep:

Mac:

```
brew install dep
```

Other Platforms:

```
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```

* Get the Ostia code

```
go get github.com/3scale/ostia
```

* Run dep ensure

```
cd ${GOPATH}/src/github.com/3scale/ostia/ostia-operator
dep ensure -v
```

* Install operator-framework from vendored sources

```
go install ./vendor/github.com/operator-framework/operator-sdk/commands/operator-sdk
```

Make sure ${GOPATH}/bin is added to your PATH

* Build

```
make build
```

* Push

```
make push
```
