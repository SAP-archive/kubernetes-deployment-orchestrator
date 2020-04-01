# Getting Started

## Setting up WordPress with an external database

During this tutorial, we would like to setup WordPress with an external production ready database.

### Install shalm

Follow the [installation instructions](installation.md) to install shalm

### Setting up an kubernetes cluster

The easiest way to get a kubernetes cluster is to use [docker for desktop](https://www.docker.com/products/docker-desktop) and to enable kubernetes in the configuration menu.

After this, your cluster should look like this:

```bash
$ kubectl get namespaces
NAME              STATUS   AGE
default           Active   55d
docker            Active   55d
kube-node-lease   Active   55d
kube-public       Active   55d
kube-system       Active   55d
```

### 1. Install the Zalando Postgres Operator

1. Create a new folder `postgres-operator`
2. Create a `Chart.star` file inside this folder with the following content:

    ```python
    def init(self):
        self.pg_operator = chart("https://github.com/zalando/postgres-operator/archive/v1.4.0.zip#charts/postgres-operator")
        self.pg_operator.load_yaml("values-crd.yaml") # Configure the postgres operator to use CRDs
        self.pg_operator.configAwsOrGcp.aws_region =  "eu-central-1" # Configure the AWS region
    ```

3. Apply it to your cluster

```bash
shalm apply postgres-operator
```

4. After this step, there should be a running pod inside the default namespace

```bash
$ kubectl get pods
NAME                                 READY   STATUS    RESTARTS   AGE
postgres-operator-8677b8bc76-vm48f   1/1     Running   0          29s
```

### 2. Create a postgres database instance

1. Create a new folder `postgres-instance`
2. Create a new folder `postgres-instance/ytt-templates`
3. Create a `Chart.star` file inside this folder with the following content:

    ```python
    def init(self,username=):
        self.pg_operator = chart("https://github.com/zalando/postgres-operator/archive/v1.4.0.zip#charts/postgres-operator")
        self.pg_operator.load_yaml("values-crd.yaml") # Configure the postgres operator to use CRDs
        self.pg_operator.configAwsOrGcp.aws_region =  "eu-central-1" # Configure the AWS region
    ```


4. Put the following template into (if you would like a HA setup use [this one](https://github.com/zalando/postgres-operator/blob/master/manifests/complete-postgres-manifest.yaml))

    ```yaml
    apiVersion: "acid.zalan.do/v1"
    kind: postgresql
    metadata:
    name: acid-minimal-cluster
    namespace: default
    spec:
    teamId: "acid"
    volume:
        size: 1Gi
    numberOfInstances: 2
    users:
        zalando:  # database owner
        - superuser
        - createdb
        foo_user: []  # role for application foo
    databases:
        foo: zalando  # dbname: owner
    postgresql:
        version: "12"
    ```
### More comming soon