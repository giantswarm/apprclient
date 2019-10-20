// +build k8srequired

package setup

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/apprclient/integration/env"
)

const (
	tillerNamespace = "giantswarm"
)

type Config struct {
	CPK8sClients *k8sclient.Clients
	K8s          *k8sclient.Setup
	Logger       micrologger.Logger
}

func NewConfig() (Config, error) {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var cpK8sClients *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger: logger,

			KubeConfigPath: env.KubeConfigPath(),
		}

		cpK8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var k8sSetup *k8s.Setup
	{
		c := k8s.SetupConfig{
			Clients: config.CPK8sClients,
			Logger:  config.Logger,
		}

		k8sSetup, err = k8s.NewSetup(c)
		if err != nil {
			return 1, microerror.Mask(err)
		}
	}

	c := Config{
		CPK8sClients: cpK8sClients,
		Logger:       logger,
		K8s:          k8sSetup,
	}

	return c, nil
}
