kubevirt-vm-operator
====================

Operator which deploy various operating systems on top of Kubernetes.

# Installation
```bash
kubectl create -f deploy/crds/guest_v1alpha1_fedora_crd.yaml
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/operator.yaml
```

# Create Fedora virtual machine
```bash
cat <<EOF | kubectl create -f -
apiVersion: guest.kubevirt.io/v1alpha1
kind: Fedora
metadata:
  name: fedora-vm
spec:
  osVersion: "29"
  memory: "512Mi"
  cpuCores: 1
  cloudInit: |
    #!/bin/bash
    echo "fedora" | passwd fedora --stdin
EOF
```

Verify VM was created:

```bash
kubectl get vm fedora-vm
```

# Troubleshooting

Check operator pod for errors:
```bash
kubectl logs kubevirt-vm-oprator-xyz
```


# Development
After cloning the repository, run the operator locally using:
```bash
export GO111MODULE=on
go mod vendor
operator-sdk up local --namespace=default
```

After changes to types file run:
```bash
operator-sdk generate k8s
```

In order to debug the operator locally using 'dlv', start the operator locally:
```bash
operator-sdk build quay.io/$USER/kubevirt-vm-operator:v0.0.1
OPERATOR_NAME=kubevirt-vm-operator WATCH_NAMESPACE=default ./build/_output/bin/kubevirt-vm-operator
```

Kubernetes cluster should be avaiable and pointed by `~/.kube/config`.
The CRDs of `./deploy/crds/` should be applied on it.

From a second terminal window run:
```bash
dlv attach --headless --api-version=2 --listen=:2345 $(pgrep -f kubevirt-vm-operator) ./build/_output/bin/kubevirt-vm-operator
```

Connect to the debug session, i.e. if using vscode, create launch.json as:

```yaml
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Connect to kubevirt-vm-operator",
            "type": "go",
            "request": "launch",
            "mode": "remote",
            "remotePath": "${workspaceFolder}",
            "port": 2345,
            "host": "127.0.0.1",
            "program": "${workspaceFolder}",
            "env": {},
            "args": []
        }
    ]
}
```