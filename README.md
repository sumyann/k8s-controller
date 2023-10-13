# Kubernetes Custom Controller for Podinfo Application

This project demonstrates a custom Kubernetes controller for deploying the Podinfo application along with a Redis data store.

## Prerequisites

- Go 1.16 or higher
- Docker (for building and pushing Docker image)
- A local Kubernetes cluster like Minikube, kind, etc., or a remote cluster.
- `kubectl` installed and configured to interact with your cluster

## Compiling the Code and Running Tests

1. Clone the repository to your local machine:
    ```bash
    git clone https://github.com/sumyann/k8s-controller.git
    cd your-repo
    ```

2. Compile the controller binary:
    ```bash
    go build -o bin/controller ./internal/controller
    ```
3. (Optional): Install the CRDs into the cluster: (encountered some issues with tests due to control plane running on kubernetes with Docker Desktop)
    ```bash
    make install
    ```
4. Run the tests to ensure everything is working as expected:
    ```bash
    go test ./... -v
    ```
## Deploying the Controller

1. Build the Docker image for the controller:
    ```bash
    make docker-build docker-push IMG=<your-registry>/controller-image:tag
    ```

2. **Update the Controller Deployment Manifest**:
    - Open the controller deployment manifest file (`config/base/controller-deployment.yaml`).
    - Locate the `image` field under the `containers` section and replace it with the image path from the registry.
    ```yaml
    
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: controller-deployment
      namespace: production
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: controller
      template:
        metadata:
          labels:
            app: controller
        spec:
          containers:
          - name: controller
            image: <your-registry>/controller-image:tag  # Replace with your image path
            ...
    ```

## Deploying the Example Custom Resource

1. Apply the example custom resource manifest to deploy the Podinfo application and Redis:
    ```bash
    make deploy
    ```

2. Once the custom resource is deployed, the controller will create the necessary resources for the Podinfo application and Redis. Verify the deployment:

    ```bash
    kubectl get all -n production
    ```

## Accessing the Podinfo UI

- The Podinfo application using a NodePort service, you can access the Podinfo UI by navigating to the service's IP address and port.
- For example, if using a NodePort service on a local cluster, you might access the UI at `http://localhost:30098`.

## Interacting with the Cache

You can interact with the Podinfo application's cache via HTTP requests to the `/cache/{key}` endpoint. For example, to save data to the cache:

```bash
curl -X POST http://<podinfo-url>:30098/cache/my-key -d '"hello world"'
curl -X PUT http://<podinfo-url>:30098/cache/my-new-key -d 'hello world'

```
To get the data from the cache:
```bash
curl http://<podinfo-url>:30098/cache/my-new-key
```

Alternative way to verify API execute script:
```
./app_verify.sh
```

## Clean Up
    ```bash
    make undeploy
    ```

