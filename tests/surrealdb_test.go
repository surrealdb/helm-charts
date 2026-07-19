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
	testTemplate(t, "deployment.yaml", "update strategy to Recreate", map[string]interface{}{
		"strategy": map[string]interface{}{
			"type": "Recreate",
		},
	})
	testTemplate(t, "deployment.yaml", "liveness and readiness probes are overridden", map[string]interface{}{
		"livenessProbe": map[string]interface{}{
			"httpGet": map[string]interface{}{
				"path": "/health",
				"port": "http",
			},
			"timeoutSeconds": 10,
		},
		"readinessProbe": map[string]interface{}{
			"httpGet": map[string]interface{}{
				"path": "/health",
				"port": "http",
			},
			"timeoutSeconds": 15,
		},
	})
	testTemplate(t, "deployment.yaml", "liveness and readiness probes are missing in values", map[string]interface{}{
		"livenessProbe":  nil,
		"readinessProbe": nil,
	})
	testTemplate(t, "deployment.yaml", "lifecycle and terminationGracePeriodSeconds are set", map[string]interface{}{
		"lifecycle": map[string]interface{}{
			"preStop": map[string]interface{}{
				"exec": map[string]interface{}{
					"command": []string{"/bin/sh", "-c", "sleep 60"},
				},
			},
		},
		"terminationGracePeriodSeconds": 300,
	})
}

func TestPVC(t *testing.T) {
	testTemplate(t, "pvc.yaml", "persistence enabled with storageClassName and selector", map[string]interface{}{
		"persistence": map[string]interface{}{
			"enabled":          true,
			"storageClassName": "fast-ssd",
			"size":             "20Gi",
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": "surrealdb-data",
				},
			},
		},
	})
}
