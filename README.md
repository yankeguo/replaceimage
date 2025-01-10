# replaceimage

A Kubernetes admission webhook that replaces container images for airgapped Kubernetes setups

## Installation

### 1. Prepare configuration for `replaceimage`

```json
{
  "imageMappings": {
    "clickhouse:latest": "myregistry/clickhouse:latest"
  },
  "imagePullSecrets": [
    {
      "name": "myregistrykey"
    }
  ]
}
```

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
  "image": "yankeguo/replaceimage:latest",
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
          image: yankeguo/ezadmis-install
          imagePullPolicy: Always
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
