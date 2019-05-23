package e2e

import (
	"os"
	"testing"

	f "github.com/operator-framework/operator-sdk/pkg/test"
)

func TestMain(m *testing.M) {
	var _, env = os.LookupEnv(f.TestNamespaceEnv)

	if env {
		f.MainEntry(m)
	}
}
