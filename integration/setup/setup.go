// +build k8srequired

package setup

import (
	"log"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/apprclient/integration/teardown"
)

func WrapTestMain(f *framework.Host, m *testing.M) {
	var v int
	var err error

	err = resources(f)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	err = teardown.Teardown(f)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	os.Exit(v)
}

func resources(f *framework.Host) error {
	err := framework.HelmCmd("registry install --wait quay.io/giantswarm/cnr-server-chart:stable -- -n cnr-server")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
