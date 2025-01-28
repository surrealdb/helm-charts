package tests

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

var (
	SurrealDBChartPath              = "../charts/surrealdb"
	KubectlTimeout                  = 1 * time.Second
	DeploymentReplicasUpdateTimeout = 20 * time.Second
)

func TestMain(m *testing.M) {
	var err error
	if env := os.Getenv("SURREALDB_CHART_PATH"); env != "" {
		SurrealDBChartPath = env
	}
	if env := os.Getenv("KUBECTL_TIMEOUT"); env != "" {
		KubectlTimeout, err = time.ParseDuration(env)
		if err != nil {
			log.Fatalf("failed to parse KUBECTL_TIMEOUT: %s", err)
		}
	}
	if env := os.Getenv("DEPLOYMENT_REPLICAS_UPDATE_TIMEOUT"); env != "" {
		DeploymentReplicasUpdateTimeout, err = time.ParseDuration(env)
		if err != nil {
			log.Fatalf("failed to parse DEPLOYMENT_REPLICAS_UPDATE_TIMEOUT: %s", err)
		}
	}
	os.Exit(m.Run())
}

// This runs a series of helm-upgrade commands to test the HPA enable/disable functionality end-to-end
// It assumes there are a Kubernetes cluster available, and the helm and kubectl commands installed.
func TestHPAEnableDisable(t *testing.T) {
	const (
		ReleaseName    = "sdb1"
		DeploymentName = "sdb1-surrealdb"
	)

	t.Cleanup(func() {
		helmUninstall(t, ReleaseName)
	})

	helmUpgrade(t, ReleaseName)
	waitUntilDeploymentHasReplicas(t, DeploymentName, 1)

	helmUpgrade(t, ReleaseName, "--set", "horizontalPodAutoscaler.enabled=true", "--set", "horizontalPodAutoscaler.minReplicas=2", "--set", "horizontalPodAutoscaler.maxReplicas=3")
	waitUntilDeploymentHasReplicas(t, DeploymentName, 2)

	helmUpgrade(t, ReleaseName, "--set", "horizontalPodAutoscaler.enabled=false")
	waitUntilDeploymentHasReplicas(t, DeploymentName, 1)
}

func helmUpgrade(t *testing.T, release string, args ...string) {
	c := exec.Command("helm", append([]string{"upgrade", "--install", release, SurrealDBChartPath}, args...)...)
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("helm upgrade failed: %s", string(res))
	}
}

type deployment struct {
	Spec *deploymentSpec `json:"spec"`
}

type deploymentSpec struct {
	Replicas int `json:"replicas"`
}

func kubectlGetDeployment(t *testing.T, name string) deployment {
	ctx, cancel := context.WithTimeout(context.Background(), KubectlTimeout)
	defer cancel()
	c := exec.CommandContext(ctx, "kubectl", "get", "deployment", name, "-o", "json")
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl get deployment failed: %s", string(res))
	}

	var deployment deployment
	err = json.Unmarshal(res, &deployment)
	if err != nil {
		t.Fatalf("failed to unmarshal deployment: %s", string(res))
	}

	return deployment
}

func helmUninstall(t *testing.T, release string) {
	c := exec.Command("helm", "uninstall", release)
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("helm uninstall failed: %s", string(res))
	}
}

func waitUntilDeploymentHasReplicas(t *testing.T, name string, replicas int) {
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			deployment := kubectlGetDeployment(t, name)
			if deployment.Spec.Replicas == replicas {
				done <- struct{}{}
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case <-done:
		return
	case <-time.After(DeploymentReplicasUpdateTimeout):
		t.Fatalf("timed out waiting for deployment to have %d replicas", replicas)
	}
}
