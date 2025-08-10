# SurrealDB Helm Chart

![Version: 0.3.7](https://img.shields.io/badge/Version-0.3.7-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.0](https://img.shields.io/badge/AppVersion-1.0.0-informational?style=flat-square)

SurrealDB is the ultimate cloud database for tomorrow's applications.

## Introduction

This chart facilitates the deployment of [SurrealDB](https://surrealdb.com/docs/surrealdb/) on Kubernetes clusters.

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| surrealdb |  | <https://github.com/surrealdb> |

## Usage

Read the Kubernetes Deployment guides in https://surrealdb.com/docs/deployment

## Overrides

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `""` | String to fully override `"surrealdb"` |
| nameOverride | string | `""` | Provide a name in place of `surrealdb` |

## General parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Assign custom [affinity] rules to the deployment |
| args | list | `["start"]` | Command line arguments to pass to SurrealDB |
| horizontalPodAutoscaler.enabled | bool | `false` | Enable the horizontal pod autoscaler for Surrealdb pods |
| horizontalPodAutoscaler.maxReplicas | int | `1` | Max pod replicas |
| horizontalPodAutoscaler.metrics | list | `[]` (See [values.yaml]) | Metrics which the autoscaler reacts to. See [kubernetes autoscale docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/) for metric format. |
| horizontalPodAutoscaler.minReplicas | int | `1` | Min pod replicas |
| nodeSelector | object | `{}` | [Node selector] |
| podAnnotations | object | `{}` | Annotations to be added to SurrealDB pods |
| podExtraEnv | list | `[]` | Extra env entries added to the SurrealDB pods |
| podSecurityContext | object | `{}` (See [values.yaml]) | Toggle and define pod-level security context. |
| replicaCount | int | `1` | The number of SurrealDB pods to run  Note that you usually scale this only when the backend supports it. For example, if you specify volumes and volumeMounts to make this SurrealDB instance stateful, you should not scale it, as it will result in two or more instances writing to the same volume or working independently. |
| resources | object | `{}` | Resource limits and requests |
| securityContext | object | `{}` (See [values.yaml]) | SurrealDB container-level security context |
| tolerations | list | `[]` | [Tolerations] for use with node taints |
| volumeMounts | list | `[]` | Additional volume mounts for SurrealDB container |
| volumes | list | `[]` | Additional volumes for SurrealDB pod |

## SurrealDB parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| surrealdb.log | string | `"info"` | Log configuration |
| surrealdb.path | string | `"memory"` | path: tikv://tikv-pd:2379 |
| surrealdb.port | int | `8000` | SurrealDB container port |

## Image parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy for SurrealDB |
| image.repository | string | `"surrealdb/surrealdb"` | Repository to use for SurrealDB |
| image.tag | string | `""` (defaults to chart appVersion) | Tag to use for SurrealDB |
| imagePullSecrets | list | `[]` | Secrets with credentials to pull images from a private registry |

## Service parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| service.annotations | object | `{}` | Service annotations |
| service.port | int | `8000` | Service port |
| service.targetPort | string | `"http"` | Target container port |
| service.type | string | `"ClusterIP"` | Service type |

## Service account parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` (defaults to the fullname template) | The name of the service account to use. |

## Horizontal Pod Autoscaler parameters

An optional horizontal pod autoscaler that, when defined, will use metrics to scale Surrealdb pods. Note that the replicaCount variable will be ignored when the horizontal pod autoscaler is used and is replaced by the minReplicas and maxReplicas defined here. The HPA can be added or removed at anytime using `helm upgrade`.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| horizontalPodAutoscaler.enabled | bool | `false` | Enable the horizontal pod autoscaler for Surrealdb pods |
| horizontalPodAutoscaler.maxReplicas | int | `1` | Max pod replicas |
| horizontalPodAutoscaler.metrics | list | `[]` (See [values.yaml]) | Metrics which the autoscaler reacts to. See [kubernetes autoscale docs](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/) for metric format. |
| horizontalPodAutoscaler.minReplicas | int | `1` | Min pod replicas |

## Ingress parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| ingress.annotations | object | `{}` | Additional ingress annotations |
| ingress.className | string | `""` | Defines which ingress controller will implement the resource |
| ingress.defaultBackend | bool | `true` | Create default backend |
| ingress.enabled | bool | `false` | Enable an ingress resource |
| ingress.hosts | list | `[]` (See [values.yaml]) | List of hosts to be covered by ingress record |
| ingress.tls | list | `[]` (See [values.yaml]) | List of TLS configuration |

## Persistence parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| persistence.accessModes | list | `["ReadWriteOnce"]` | Access modes for the PVC |
| persistence.annotations | object | `{}` | Annotations for the PVC |
| persistence.enabled | bool | `false` | Enable persistent storage |
| persistence.mountPath | string | `"/data"` | Mount path for the persistent volume |
| persistence.selector | object | `{}` | Selector to match an existing Persistent Volume |
| persistence.size | string | `"10Gi"` | Size of the persistent volume |
| persistence.storageClassName | string | `""` (uses default storage class) | Storage class name for the PVC |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs](https://github.com/norwoodj/helm-docs)