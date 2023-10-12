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

2. Run the tests to ensure everything is working as expected:
    ```bash
    go test ./... -v
    ```

3. Compile the controller binary:
    ```bash
    go build -o bin/controller main.go
    ```

## Deploying the Controller

1. Build the Docker image for the controller:
    ```bash
    docker build -t your-username/controller:latest .
    ```

2. Push the Docker image to a registry:
    ```bash
    docker push your-username/controller:latest
    ```

3. Apply the necessary RBAC, CRD, and Deployment manifests for the controller:
    ```bash
    kubectl apply -f manifests/
    ```

## Deploying the Example Custom Resource

1. Apply the example custom resource manifest to deploy the Podinfo application and Redis:
    ```bash
    kubectl apply -f examples/myAppResource.yaml
    ```

2. Once the custom resource is deployed, the controller will create the necessary resources for the Podinfo application and Redis. Verify the deployment:

    ```bash
    kubectl get deployments,pods,services -n <namespace>
    ```

## Accessing the Podinfo UI

- If you've exposed the Podinfo application using a NodePort or LoadBalancer service, you can access the Podinfo UI by navigating to the service's IP address and port.
- For example, if using a NodePort service on a local cluster, you might access the UI at `http://localhost:30098`.

## Interacting with the Cache

You can interact with the Podinfo application's cache via HTTP requests to the `/cache/{key}` endpoint. For example, to save data to the cache:

```bash
curl -X POST http://<podinfo-url>/cache/my-key -d '"hello world"'