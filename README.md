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

* Add cluster-admin perms to the ostia default user, so it can watch events on all namespaces:

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
  * Operator-SDK (<https://github.com/operator-framework/operator-sdk>)

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

* Let's install the operator-sdk, make sure `$GOPATH` is set and included in PATH `$GOPATH/bin`

```
mkdir -p $GOPATH/src/github.com/operator-framework/
git clone https://github.com/operator-framework/operator-sdk.git \
          $GOPATH/src/github.com/operator-framework/operator-sdk

cd $GOPATH/src/github.com/operator-framework/operator-sdk
dep ensure
go install github.com/operator-framework/operator-sdk/commands/operator-sdk
```

* Clone the ostia repo into your GOPATH

```
mkdir -p ${GOPATH}/src/github.com/3scale/
git clone https://github.com/3scale/ostia.git ${GOPATH}/src/github.com/3scale/ostia
```

* Run dep ensure

```
cd ${GOPATH}/src/github.com/3scale/ostia/ostia-operator
dep ensure
```

* Build

```
make build
```

* Push

```
make push
```