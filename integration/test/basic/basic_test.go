// +build k8srequired

package basic

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	"github.com/giantswarm/apprclient"
)

func Test_Client_GetReleaseVersion(t *testing.T) {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	restCfg := f.RestConfig()
	fwc := k8sportforward.Config{
		RestConfig: restCfg,
	}

	fw, err := k8sportforward.New(fwc)
	if err != nil {
		t.Fatalf("could not create forwarder %v", err)
	}

	podName, err := f.GetPodName("default", "app=cnr-server")
	if err != nil {
		t.Fatalf("could not get cnr-server pod name %v", err)
	}
	tc := k8sportforward.TunnelConfig{
		Remote:    5000,
		Namespace: "default",
		PodName:   podName,
	}
	tunnel, err := fw.ForwardPort(tc)
	if err != nil {
		t.Fatalf("could not create tunnel %v", err)
	}

	serverAddress := "http://localhost:" + strconv.Itoa(tunnel.Local)
	err = waitForServer(f, serverAddress+"/cnr/api/v1/packages")
	if err != nil {
		t.Fatalf("server didn't come up on time")
	}

	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      serverAddress,
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		t.Fatalf("could not create appr %v", err)
	}

	err = a.PushChartTarball("tb-chart", "5.5.5", "/e2e/fixtures/tb-chart.tar.gz")
	if err != nil {
		t.Fatalf("could not push tarball %v", err)
	}

	err = a.PromoteChart("tb-chart", "5.5.5", "5-5-beta")
	if err != nil {
		t.Fatalf("could not promote chart %v", err)
	}

	expected := "5.5.5"
	actual, err := a.GetReleaseVersion("tb-chart", "5-5-beta")
	if err != nil {
		t.Fatalf("could not get release %v", err)
	}

	if expected != actual {
		t.Fatalf("release didn't match expected, want %q, got %q", expected, actual)
	}
}

func waitForServer(f *framework.Host, url string) error {
	var err error

	operation := func() error {
		_, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("could not retrieve %s: %v", url, err)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for server at %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
