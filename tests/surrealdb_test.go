package tests

import "testing"

func TestDeployment(t *testing.T) {
	testTemplate(t, "deployment.yaml", "default", map[string]interface{}{})
}
