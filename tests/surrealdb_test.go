package tests

import "testing"

func TestDeployment(t *testing.T) {
	testTemplate(t, "deployment.yaml", "default", map[string]interface{}{})
	testTemplate(t, "deployment.yaml", "disable auth using deprecated surrealdb.auth", map[string]interface{}{
		"surrealdb": map[string]interface{}{
			"auth": false,
		},
	})
	testTemplate(t, "deployment.yaml", "disable auth using surrealdb.unauhenticated", map[string]interface{}{
		"unauthenticated": true,
	})
}
