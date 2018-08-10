package main

import (
	"context"
	"runtime"

	stub "github.com/3scale/ostia/ostia-operator/pkg/stub"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/sirupsen/logrus"
	"os"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func printInfo(namespace string) {
	if namespace == "" {
		logrus.Infof("Listening on all the namespaces")
	} else {
		logrus.Infof("Listening to namespace: %s", namespace)
	}
}

func main() {
	namespace := os.Getenv("WATCH_NAMESPACE")
	printVersion()
	printInfo(namespace)
	sdk.Watch("ostia.3scale.net/v1alpha1", "API", namespace, 5)
	sdk.Handle(stub.NewHandler())
	sdk.Run(context.TODO())
}
