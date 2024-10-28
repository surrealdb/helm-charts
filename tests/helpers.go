package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// testTemplate is a helper function that tests a "surrealdb" chart template.
//
// It runs helm-template to render the template at the specified path within the surrealdb chart,
// with the specified values, and compares the output to the expected output corresponds to the specified subject.
//
// We do snapshot testing here, so the caller does not need to manually populate the expected template output
// corresponds to the subject and the values.
//
// To let the test record the snapshot, set the environment variable `UPDATE_SNAPSHOT={path}/{subject}` before running the test.
//
// For example, to update the snapshot for the "default" subject of the "deployment.yaml" template:
// ```
// UPDATE_SNAPSHOT=deployment.yaml/default go test -v ./tests
// ```
//
// To update the snapshot for all subjects of the "deployment.yaml" template:
// ```
// UPDATE_SNAPSHOT="deployment.yaml/*" go test -v ./tests
// ```
//
// To update the snapshot for all templates:
// ```
// UPDATE_SNAPSHOT="*" go test -v ./tests
// ```
func testTemplate(t *testing.T, path string, subject string, values map[string]interface{}) {
	t.Helper()

	name := fmt.Sprintf("%s/%s", path, subject)

	t.Run(name, func(t *testing.T) {
		actual := renderTemplate(t, filepath.Join("templates", path), values)

		if shouldUpdateSnapshot(os.Getenv("UPDATE_SNAPSHOT"), path, subject) {
			writeSnapshot(t, name, actual)
			t.Skip("Updated snapshot")
		}

		expected := readSnapshot(t, name)

		assert.Equal(t, expected, actual)
	})
}

func renderTemplate(t *testing.T, path string, values map[string]interface{}) string {
	t.Helper()

	valuesFile := filepath.Join(t.TempDir(), "values.yaml")
	if err := writeValuesFile(valuesFile, values); err != nil {
		t.Fatalf("failed to write values file: %v", err)
	}

	helmCmd := exec.Command(
		"helm", "template",
		// Release name
		"testrelease",
		// Chart path
		"./charts/surrealdb",
		"--values", valuesFile,
		"--show-only", path,
	)
	// We assume the test is running from the tests directory,
	// which is one-level deep from the root of the project.
	helmCmd.Dir = ".."

	out, err := helmCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to render template: %v\n%s", err, out)
	}

	return string(out)
}

func writeValuesFile(path string, values map[string]interface{}) error {
	d, err := yaml.Marshal(values)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, d, 0644); err != nil {
		return err
	}

	return nil
}

func shouldUpdateSnapshot(env, path, subject string) bool {
	switch env {
	case "*", fmt.Sprintf("%s/*", path), fmt.Sprintf("%s/%s", path, subject):
		return true
	}

	return false
}

func readSnapshot(t *testing.T, name string) string {
	t.Helper()

	f := getSnapshotPath(name)
	d, err := os.ReadFile(f)
	if err != nil {
		t.Logf("Create snapshot file at %s", f)
		t.Fatalf("failed to read snapshot: %v", err)
	}

	return string(d)
}

func writeSnapshot(t *testing.T, name string, actual string) {
	t.Helper()

	f := getSnapshotPath(name)

	if err := os.MkdirAll(filepath.Dir(f), 0755); err != nil {
		t.Fatalf("failed to create snapshot directory: %v", err)
	}

	if err := os.WriteFile(f, []byte(actual), 0644); err != nil {
		t.Fatalf("failed to write snapshot: %v", err)
	}

	t.Logf("Snapshot updated at %s", f)
}

func getSnapshotPath(name string) string {
	return filepath.Join("testdata", "snapshots", name)
}
