package tests

import "testing"

func TestDeployment(t *testing.T) {
	testTemplate(t, "deployment.yaml", "default", map[string]interface{}{})
	testTemplate(t, "deployment.yaml", "release deployed using pre v0.3.5 chart with the default surrealdb.auth", map[string]interface{}{
		"surrealdb": map[string]interface{}{
			"auth": true,
		},
	})
	testTemplate(t, "deployment.yaml", "disable auth using deprecated surrealdb.auth", map[string]interface{}{
		"surrealdb": map[string]interface{}{
			"auth": false,
		},
	})
	testTemplate(t, "deployment.yaml", "surrealdb.unauthenticated=false has no effect", map[string]interface{}{
		"surrealdb": map[string]interface{}{
			"unauthenticated": false,
		},
	})
	testTemplate(t, "deployment.yaml", "disable auth using surrealdb.unauhenticated", map[string]interface{}{
		"surrealdb": map[string]interface{}{
			"unauthenticated": true,
		},
	})
	testTemplate(t, "deployment.yaml", "volumes and volumeMounts are set", map[string]interface{}{
		"args": []string{"start", "surrealkv:/var/lib/surrealdb"},
		"volumeMounts": []interface{}{
			map[string]interface{}{
				"mountPath": "/var/lib/surrealdb",
				"name":      "surrealdb-data",
			},
		},
		"volumes": []interface{}{
			map[string]interface{}{
				"name": "surrealdb-data",
				"persistentVolumeClaim": map[string]interface{}{
					"claimName": "surrealdb-data",
				},
			},
		},
	})
	testTemplate(t, "deployment.yaml", "init and extra containers are set", map[string]interface{}{
		"initContainers": []interface{}{
			map[string]interface{}{
				"name":  "init-myservice",
				"image": "myservice:latest",
				"command": []string{
					"/bin/sh",
					"-c",
					"echo Init Container",
				},
			},
		},
		"extraContainers": []interface{}{
			map[string]interface{}{
				"name":  "extra-myservice",
				"image": "myservice:latest",
				"command": []string{
					"/bin/sh",
					"-c",
					"echo Extra Container",
				},
			},
		},
	})
}
