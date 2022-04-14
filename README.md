# Onboarding Operator Kubernetes

## Overview

This kubernetes operator creates an environment for users desiring to use a kubernetes plateform. The environment is a Custom Resource that manages the creation kubernetes resources:
- a namespace: The namespace where the applications related to  this environment will be deployed
- a resourcequota: The  quota of the namespace
- a rolebinding: To give required permissions to the users of the namespaces.



## Prerequisites

- [go][go_tool] version v1.14+.
- [docker][docker_tool] version 17.03+
- [kubectl][kubectl_tool] v1.14.1+
- [operator-sdk][operator_install]
- Access to a Kubernetes v1.14.1+ cluster

## Getting Started

### Cloning the repository

Checkout this Onboarding Operator repository

```
$ mkdir -p $GOPATH/src/gitlab.beopenit.com
$ cd $GOPATH/src/gitlab.beopenit.com
$ git clone https://gitlab.beopenit.com/cloud/onboarding-operator-kubernetes.git
$ cd onboarding-operator-kubernetes
```
### Pulling the dependencies

Run the following command

```
$ go mod tidy
```

### Running locally

Set the name of the operator in an environment variable

```
export OPERATOR_NAME=onboarding-operator-kubernetes
```

Run the operator locally with the default Kubernetes config file present at $HOME/.kube/config.

```
make run WATCH_NAMESPACE=""
```


### Building the operator

Build the Onboarding operator image and push it to a public registry, such as quay.io:

```
$ export IMAGE=quay.io/example-inc/onboarding-operator-kubernetes:v0.0.1
$ operator-sdk build $IMAGE
$ docker push $IMAGE
```

### Using the image

```
# Update the operator manifest to use the built image name (if you are performing these steps on OSX, see note below)
$ sed -i 's|REPLACE_IMAGE|quay.io/example-inc/onboarding-operator-kubernetes:v0.0.1|g' deploy/operator.yaml
# On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/example-inc/onboarding-operator-kubernetes:v0.0.1|g' deploy/operator.yaml
```

**NOTE** The `quay.io/example-inc/onboarding-operator-kubernetes:v0.0.1` is an example. You should build and push the image for your repository.

### Installing

Run `make install` to install the operator. Check that the operator is running in the cluster, also check that the example Onboarding service was deployed.

Following the expected result.

```shell
$ kubectl get all -n onboarding
NAME                                      READY   STATUS    RESTARTS   AGE
pod/onboarding-operator-kubernetes-56f54d84bf-zrtfv   1/1     Running   0          69s

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/onboarding-operator-kubernetes-metrics   ClusterIP   10.108.67.82    <none>        8383/TCP,8686/TCP   66s

NAME                                 READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/onboarding-operator-kubernetes   1/1     1            1           70s

NAME                                            DESIRED   CURRENT   READY   AGE
replicaset.apps/onboarding-operator-kubernetes-56f54d84bf   1         1         1       70s
```

### Uninstalling

To uninstall all that was performed in the above step run `make uninstall`.

### Troubleshooting

Use the following command to check the operator logs.

```shell
kubectl logs deployment.apps/onboarding-operator-kubernetes -n onboarding
```

### Running Tests

Run `make test-e2e` to run the integration e2e tests with different options. For
more information see the [writing e2e tests][golang-e2e-tests] guide.

[dep_tool]: https://golang.github.io/dep/docs/installation.html
[go_tool]: https://golang.org/dl/
[kubectl_tool]: https://kubernetes.io/docs/tasks/tools/install-kubectl/
[docker_tool]: https://docs.docker.com/install/
[operator_sdk]: https://github.com/operator-framework/operator-sdk
[operator_install]: https://sdk.operatorframework.io/docs/install-operator-sdk/
[golang-e2e-tests]: https://sdk.operatorframework.io/docs/golang/e2e-tests/