package tests

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var (
	SurrealDBChartPath              = "../charts/surrealdb"
	SurrealDBImageTag               = ""
	KubectlTimeout                  = 1 * time.Second
	DeploymentReplicasUpdateTimeout = 20 * time.Second
	DeploymentReadyTimeout          = 3 * time.Minute
)

func TestMain(m *testing.M) {
	var err error
	if env := os.Getenv("SURREALDB_CHART_PATH"); env != "" {
		SurrealDBChartPath = env
	}
	if env := os.Getenv("SURREALDB_IMAGE_TAG"); env != "" {
		SurrealDBImageTag = env
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
	if env := os.Getenv("DEPLOYMENT_READY_TIMEOUT"); env != "" {
		DeploymentReadyTimeout, err = time.ParseDuration(env)
		if err != nil {
			log.Fatalf("failed to parse DEPLOYMENT_READY_TIMEOUT: %s", err)
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

// TestPersistence installs SurrealDB with a PVC, waits for Ready, deletes the pod,
// and asserts it becomes Ready again on the same claim (kind local-path / default SC).
// Uses podSecurityContext.fsGroup=65532 because the official image USER is nonroot
// (https://github.com/surrealdb/surrealdb/blob/main/docker/Dockerfile).
func TestPersistence(t *testing.T) {
	const (
		ReleaseName    = "sdb-persist"
		DeploymentName = "sdb-persist-surrealdb"
		PVCName        = "sdb-persist-surrealdb"
	)

	assertStorageClassAvailable(t)

	t.Cleanup(func() {
		helmUninstall(t, ReleaseName)
	})

	helmUpgrade(t, ReleaseName,
		"--set", "strategy.type=Recreate",
		"--set", "persistence.enabled=true",
		"--set", "persistence.size=1Gi",
		"--set", "surrealdb.path=surrealkv:/data",
		"--set", "podSecurityContext.fsGroup=65532",
	)

	waitUntilDeploymentReady(t, DeploymentName, 1)
	assertPVCBound(t, PVCName)

	uidBefore := podUIDForDeployment(t, DeploymentName)
	deletePodsForDeployment(t, DeploymentName)
	waitUntilDeploymentReady(t, DeploymentName, 1)

	uidAfter := podUIDForDeployment(t, DeploymentName)
	if uidAfter == "" || uidAfter == uidBefore {
		t.Fatalf("expected a new pod after delete, before=%q after=%q", uidBefore, uidAfter)
	}
	assertPVCBound(t, PVCName)
}

func helmUpgrade(t *testing.T, release string, args ...string) {
	t.Helper()
	cmdArgs := []string{"upgrade", "--install", release, SurrealDBChartPath}
	if SurrealDBImageTag != "" {
		cmdArgs = append(cmdArgs, "--set", "image.tag="+SurrealDBImageTag)
	}
	cmdArgs = append(cmdArgs, args...)
	c := exec.Command("helm", cmdArgs...)
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("helm upgrade failed: %s", string(res))
	}
}

type deployment struct {
	Spec   *deploymentSpec   `json:"spec"`
	Status *deploymentStatus `json:"status"`
}

type deploymentSpec struct {
	Replicas int `json:"replicas"`
}

type deploymentStatus struct {
	ReadyReplicas     int `json:"readyReplicas"`
	AvailableReplicas int `json:"availableReplicas"`
}

type storageClassList struct {
	Items []storageClass `json:"items"`
}

type storageClass struct {
	Metadata struct {
		Name        string            `json:"name"`
		Annotations map[string]string `json:"annotations"`
	} `json:"metadata"`
}

type pvc struct {
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

type podList struct {
	Items []pod `json:"items"`
}

type pod struct {
	Metadata struct {
		Name string `json:"name"`
		UID  string `json:"uid"`
	} `json:"metadata"`
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
}

func kubectlGetDeployment(t *testing.T, name string) deployment {
	t.Helper()
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
	t.Helper()
	c := exec.Command("helm", "uninstall", release)
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("helm uninstall failed: %s", string(res))
	}
}

func waitUntilDeploymentHasReplicas(t *testing.T, name string, replicas int) {
	t.Helper()
	deadline := time.Now().Add(DeploymentReplicasUpdateTimeout)
	for time.Now().Before(deadline) {
		deployment := kubectlGetDeployment(t, name)
		if deployment.Spec != nil && deployment.Spec.Replicas == replicas {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("timed out waiting for deployment to have %d replicas", replicas)
}

func waitUntilDeploymentReady(t *testing.T, name string, replicas int) {
	t.Helper()
	deadline := time.Now().Add(DeploymentReadyTimeout)
	var last deployment
	for time.Now().Before(deadline) {
		last = kubectlGetDeployment(t, name)
		if last.Status != nil &&
			last.Status.ReadyReplicas >= replicas &&
			last.Status.AvailableReplicas >= replicas {
			return
		}
		time.Sleep(2 * time.Second)
	}
	dump := kubectlDebug(t, "get", "pods,pvc,events", "-o", "wide")
	t.Fatalf("timed out waiting for deployment %s to be ready (%d replicas); last status=%+v\n%s",
		name, replicas, last.Status, dump)
}

func assertStorageClassAvailable(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), KubectlTimeout)
	defer cancel()
	c := exec.CommandContext(ctx, "kubectl", "get", "storageclass", "-o", "json")
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl get storageclass failed: %s", string(res))
	}

	var list storageClassList
	if err := json.Unmarshal(res, &list); err != nil {
		t.Fatalf("failed to unmarshal storageclasses: %s", err)
	}
	if len(list.Items) == 0 {
		t.Fatal("no StorageClass found; kind provides local-path as 'standard' (or 'local-path') by default — create a kind cluster with default storage before running this test")
	}

	for _, sc := range list.Items {
		if sc.Metadata.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			t.Logf("using default StorageClass %q", sc.Metadata.Name)
			return
		}
	}
	for _, sc := range list.Items {
		if sc.Metadata.Name == "standard" || sc.Metadata.Name == "local-path" {
			t.Logf("no default StorageClass; found kind-compatible StorageClass %q", sc.Metadata.Name)
			return
		}
	}
	names := make([]string, 0, len(list.Items))
	for _, sc := range list.Items {
		names = append(names, sc.Metadata.Name)
	}
	t.Fatalf("no usable StorageClass (want default, standard, or local-path); found: %s", strings.Join(names, ", "))
}

func assertPVCBound(t *testing.T, name string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), KubectlTimeout)
	defer cancel()
	c := exec.CommandContext(ctx, "kubectl", "get", "pvc", name, "-o", "json")
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl get pvc failed: %s", string(res))
	}
	var claim pvc
	if err := json.Unmarshal(res, &claim); err != nil {
		t.Fatalf("failed to unmarshal pvc: %s", err)
	}
	if claim.Status.Phase != "Bound" {
		t.Fatalf("expected PVC %s to be Bound, got %q", name, claim.Status.Phase)
	}
}

func deletePodsForDeployment(t *testing.T, deploymentName string) {
	t.Helper()
	// DeploymentName is "<release>-surrealdb"; instance label is the release name.
	release := strings.TrimSuffix(deploymentName, "-surrealdb")
	c := exec.Command("kubectl", "delete", "pod", "-l", "app.kubernetes.io/instance="+release, "--wait=true")
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl delete pod failed: %s", string(res))
	}
}

func podUIDForDeployment(t *testing.T, deploymentName string) string {
	t.Helper()
	release := strings.TrimSuffix(deploymentName, "-surrealdb")
	ctx, cancel := context.WithTimeout(context.Background(), KubectlTimeout)
	defer cancel()
	c := exec.CommandContext(ctx, "kubectl", "get", "pods", "-l", "app.kubernetes.io/instance="+release, "-o", "json")
	res, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("kubectl get pods failed: %s", string(res))
	}
	var list podList
	if err := json.Unmarshal(res, &list); err != nil {
		t.Fatalf("failed to unmarshal pods: %s", err)
	}
	if len(list.Items) == 0 {
		return ""
	}
	return list.Items[0].Metadata.UID
}

func kubectlDebug(t *testing.T, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c := exec.CommandContext(ctx, "kubectl", args...)
	res, _ := c.CombinedOutput()
	return string(res)
}
