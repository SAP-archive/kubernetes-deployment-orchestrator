# Getting Started

## Setting up WordPress with an external database

During this tutorial, we would like to setup WordPress with an external production ready database. You can find the complete source code under `charts/examples/tutorial`.
The bitnami worldpress helm chart is already bundled with a mysql database, which has only a few configuration parameters.The goal of this tutorial will be to
setup a wordpress installation with a highly configurable database.

### Install kdo

Follow the [installation instructions](installation.md) to install kdo

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
      # Load chart from github
      self.pg_operator = chart("https://github.com/zalando/postgres-operator/archive/v1.4.0.zip#charts/postgres-operator")
      # Configure the postgres operator to use CRDs. This loads a yaml located in the same directory as the chart
      self.pg_operator.load_yaml("values-crd.yaml")
      # Configure the AWS region
      # You can modify any values of a helm chart
      self.pg_operator.configAwsOrGcp.aws_region =  "eu-central-1" 
    ```

3. Apply it to your cluster

    ```bash
    kdo apply postgres-operator
    ```

4. After this step, there should be a running pod inside the default namespace

    ```bash
    $ kubectl get pods
    NAME                                 READY   STATUS    RESTARTS   AGE
    postgres-operator-8677b8bc76-vm48f   1/1     Running   0          29s
    ```

### 2. Create a postgres database instance

1. Create a new folder `postgres-instance`
2. Create a `Chart.star` file inside this folder with the following content:

    ```python
    def init(self):
      pass
    ```

3. Create a new folder `postgres-instance/ytt-templates`
4. Put the following template into `postgres-instance/ytt-templates/postgres.yml` (if you would like a HA setup use [this one](https://github.com/zalando/postgres-operator/blob/master/manifests/complete-postgres-manifest.yaml))

    ```yaml
    apiVersion: "acid.zalan.do/v1"
    kind: postgresql
    metadata:
      name: postgres
      namespace: default
    spec:
      teamId: "acid"
      volume:
        size: 1Gi
      numberOfInstances: 2
      users:
        zalando:  #! database owner
        - superuser
        - createdb
        foo_user: []  #! role for application foo
      databases:
        foo: zalando  #! dbname: owner
      postgresql:
        version: "12"
    ```

5. As next step, we would like to configure the name of the database administrator and the user and also the name of the database. Therefore, we add parameters to the constructor and modify the template accordingly. If you have questions regarding the template syntax, please visit [ytt](https://get-ytt.io/)


    ```python
    def init(self,admin_user="admin",user="user",db="db"):
      self.admin_user = admin_user
      self.user = user
      self.db = db
    ```

    ```yaml
    apiVersion: "acid.zalan.do/v1"
    kind: postgresql
    metadata:
      name: postgres
      namespace: default
    spec:
      teamId: "acid"
      volume:
        size: 1Gi
      numberOfInstances: 2
      users: #@ { self.admin_user : [ "superuser" , "createdb"] , self.user : [] }
      databases: #@ { self.db : self.user }
      postgresql:
        version: "12"
    ```

6. Apply it to your cluster

  ```bash
  kdo apply postgres-instance
  ```

### More comming soon