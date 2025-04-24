# replaceimage

A Kubernetes admission webhook that replaces container images for airgapped Kubernetes setups

## Installation

### 1. Prepare configuration for `replaceimage`

```json
{
  "imageMappings": {
    "clickhouse:latest": "myregistry/clickhouse:latest"
  },
  "imageAutoMapping": {
    "match": {
      "namespaces": ["default", "kube-.+"],
      "images": ["^docker.io/.*", "^ghcr.io/.*", "^gcr.io/.*"]
    },
    "registry": "registry-vpc.mycompany.com",
    "webhook": {
      "override": {
        "registry": "registry.mycompany.com"
      },
      "url": "https://webhook.mycompany.com/jenkins/jobs/replaceimage",
      "headers": {
        "Authorization": "Basic YWJjZGVmZ2g6aGk="
      },
      "query": {
        "ARG_SRC": "$SOURCE_IMAGE",
        "ARG_DST": "$TARGET_IMAGE"
      },
      "form": {
        "ARG_SRC": "$SOURCE_IMAGE",
        "ARG_DST": "$TARGET_IMAGE"
      },
      "json": {
        "ARG_SRC": "$SOURCE_IMAGE",
        "ARG_DST": "$TARGET_IMAGE"
      }
    }
  },
  "imagePullSecrets": [
    {
      "name": "myregistrykey"
    }
  ]
}
```

**Auto Image Mapping Rules**

- `.`, `/` will be replaced with `-` in the target image name.
- Duplicated path components will be removed.

Examples:

- `source.registry/my-org/my-image` -> `registry.mycompany.com/namespace/source-registry-my-org-my-image`
- `source.registry/my-image/my-image` -> `registry.mycompany.com/namespace/source-registry-my-image`
- `source.registry/my-org/my-image/my-image` -> `registry.mycompany.com/namespace/source-registry-my-org-my-image`

### 2. Prepare ConfigMap for `replaceimage`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: replaceimage-config
  namespace: autoops
data:
  replaceimage.json: |
    PUT CONFIGURATION HERE!!!
```

### 3. Complete RBAC initialization for `ezadmis-install`

See https://github.com/yankeguo/ezadmis/tree/main/cmd/ezadmis-install#rbac-initialization

### 4. Prepare configuration for `ezadmis-install`

```json
{
  "name": "replaceimage",
  "namespace": "autoops",
  "mutating": true,
  "admissionRules": [
    {
      "operations": ["CREATE"],
      "apiGroups": [""],
      "apiVersions": ["v1"],
      "resources": ["pods"]
    }
  ],
  "sideEffects": "Some",
  "failurePolicy": "Ignore",
  "image": "yankeguo/replaceimage:0.2.0",
  "imagePullPolicy": "IfNotPresent",
  "volumes": [
    {
      "name": "vol-cfg",
      "configMap": {
        "name": "replaceimage-config"
      }
    }
  ],
  "volumeMounts": [
    {
      "name": "vol-cfg",
      "mountPath": "/config"
    }
  ]
}
```

More details can be found at https://github.com/yankeguo/ezadmis/tree/main/cmd/ezadmis-install

### 5. Install `replaceimage` with `ezadmis-install`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: install-replaceimage-cfg
  namespace: autoops
data:
  config.json: |
    PUT CONFIGURATION HERE!!!
---
# Job
apiVersion: batch/v1
kind: Job
metadata:
  name: install-replaceimage
  namespace: autoops
spec:
  template:
    spec:
      serviceAccountName: ezadmis-install
      automountServiceAccountToken: true
      containers:
        - name: install-replaceimage
          image: yankeguo/ezadmis-install:0.4.1
          imagePullPolicy: IfNotPresent
          args:
            - /ezadmis-install
            - -conf
            - /config.json
          volumeMounts:
            - name: vol-cfg
              mountPath: /config.json
              subPath: config.json
      volumes:
        - name: vol-cfg
          configMap:
            name: install-replaceimage-cfg
      restartPolicy: OnFailure
```

## Credits

GUO YANKE, MIT License
