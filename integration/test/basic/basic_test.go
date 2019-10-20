// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func testSetup(ctx context.Context, t *testing.T) (*apprclient.Client, *k8sportforward.Tunnel) {
	var err error

	var fw *k8sportforward.Forwarder
	{
		c := k8sportforward.ForwarderConfig{
			RestConfig: setup.K8sClients.RESTConfig(),
		}

		fw, err = k8sportforward.NewForwarder(c)
		if err != nil {
			t.Fatalf("could not create forwarder %v", err)
		}
	}

	var podName string
	{
		lo := metav1.ListOptions{
			LabelSelector: "app=cnr-server",
		}
		pods, err := setup.CPK8sClient().CoreV1().Pods(metav1.NamespaceDefault).List(lo)
		if err != nil {
			t.Fatalf("could not list pods %v", err)
		}
		if len(pods.Items) != 1 {
			t.Fatalf("expected 1 pod got %d", len(pods.Items))
		}

		podName = pods.Items[0].Name
	}

	var tunnel *k8sportforward.Tunnel
	{
		tunnel, err = fw.ForwardPort("default", podName, 5000)
		if err != nil {
			t.Fatalf("could not create tunnel %v", err)
		}
	}

	serverAddress := "http://" + tunnel.LocalAddress()
	err = waitForServer(serverAddress + "/cnr/api/v1/packages")
	if err != nil {
		t.Fatalf("server didn't come up on time")
	}

	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: config.Logger,

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
	ctx := context.Background()
	var err error

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

func waitForServer(url string) error {
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
