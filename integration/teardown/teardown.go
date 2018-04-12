// +build k8srequired

package teardown

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
)

func Teardown(f *framework.Host) error {
	err := framework.HelmCmd("delete cnr-server --purge")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
