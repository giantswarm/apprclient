// +build k8srequired

package basic

import (
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/apprclient/integration/setup"
)

var (
	h *framework.Host
)

func init() {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			panic(err.Error())
		}
	}

	{
		c := framework.HostConfig{
			Logger: logger,

			ClusterID:  "someval",
			VaultToken: "someval",
		}

		h, err = framework.NewHost(c)
		if err != nil {
			panic(err.Error())
		}
	}

}

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	setup.WrapTestMain(h, m)
}
