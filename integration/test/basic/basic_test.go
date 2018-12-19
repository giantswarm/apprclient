// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

func testSetup(ctx context.Context, t *testing.T) (*apprclient.Client, *k8sportforward.Tunnel) {
	var err error

	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	var fw *k8sportforward.Forwarder
	{
		restCfg := h.RestConfig()

		c := k8sportforward.ForwarderConfig{
			RestConfig: restCfg,
		}

		fw, err = k8sportforward.NewForwarder(c)
		if err != nil {
			t.Fatalf("could not create forwarder %v", err)
		}
	}

	podName, err := h.GetPodName("default", "app=cnr-server")
	if err != nil {
		t.Fatalf("could not get cnr-server pod name %v", err)
	}
	tunnel, err := fw.ForwardPort("default", podName, 5000)
	if err != nil {
		t.Fatalf("could not create tunnel %v", err)
	}

	serverAddress := "http://" + tunnel.LocalAddress()
	err = waitForServer(h, serverAddress+"/cnr/api/v1/packages")
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
	return a, tunnel
}

func testTeardown(ctx context.Context, a *apprclient.Client, tunnel *k8sportforward.Tunnel, t *testing.T) {
	err := a.DeleteRelease(ctx, "tb-chart", "5.5.5")
	if err != nil {
		t.Fatalf("could not delete release %v", err)
	}

	tunnel.Close()
}

func Test_Client_GetReleaseVersion(t *testing.T) {
	var err error

	ctx := context.Background()

	a, tunnel := testSetup(ctx, t)
	defer testTeardown(ctx, a, tunnel, t)

	err = a.PushChartTarball(ctx, "tb-chart", "5.5.5", "/e2e/fixtures/tb-chart.tar.gz")
	if err != nil {
		t.Fatalf("could not push tarball %v", err)
	}

	err = a.PromoteChart(ctx, "tb-chart", "5.5.5", "5-5-beta")
	if err != nil {
		t.Fatalf("could not promote chart %v", err)
	}

	expected := "5.5.5"
	actual, err := a.GetReleaseVersion(ctx, "tb-chart", "5-5-beta")
	if err != nil {
		t.Fatalf("could not get release %v", err)
	}

	if expected != actual {
		t.Fatalf("release didn't match expected, want %q, got %q", expected, actual)
	}
}

func waitForServer(h *framework.Host, url string) error {
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
