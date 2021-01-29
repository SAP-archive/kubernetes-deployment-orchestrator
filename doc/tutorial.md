# Tutorial

This tutorial will guide you through the different topics of kdo

## Starlark

Starlark is a python dialect, which is also use in bazel or ytt. It allows you to run code
in a controlled sandbox without access to the outer world.

## Interoperability with helm

Kubernete deployment orchestrator is designed from ground up to be mostly compatible with helm packages



### Helm to kdo

You can deploy almost every helm chart using *kdo* except those, which uses [hooks](https://helm.sh/docs/topics/charts_hooks/) for deployment. 

```
kdo apply helm://charts.helm.sh/stable/mysql
kubectl get pods
kdo list -A
kdo delete helm://charts.helm.sh/stable/mysql

```

### Kubernete deployment orchestrator to helm

Kubernete deployment orchestrator charts can be wrapped into helm charts using `kdo package --helm`. Kubernete deployment orchestrator uses [hooks](https://helm.sh/docs/topics/charts_hooks/) to implement this.


```python
# Chart.star
def init(self):
  self.__class__.version = "0.0.1"
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.__class__.version = "0.0.1"
EOF
kdo package --helm .
helm upgrade -i example example-0.0.1.tgz
helm uninstall example
```


## Templating

In kdo, you can chose between go templating and [ytt](https://get-ytt.io/)

### ytt templating


You should prefer ytt templating because ytt also uses starlark. Therefore the interoperability between kdo and ytt is much better compared to helm templating.
You can directly call methods from ytt.

```python
# Chart.star
def init(self):
  self.timeout = 30
```

```yaml
kind: Secret
stringData:
  timeout: #@ self.timeout
```


```bash
rm -rf /tmp/example
mkdir -p /tmp/example/ytt-templates
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.timeout = 30
EOF
cat > ytt-templates/secret.yaml <<EOF
kind: Secret
stringData:
  timeout: #@ self.timeout
EOF
kdo template .
```

### Go Templating (helm)

Go templating is mostly compatible with helm. All required functions are provided. All attributes of `self` are available as `.Values`

```python
# Chart.star
def init(self):
  self.timeout = 30
```

```yaml
kind: Secret
stringData:
  timeout: {{ .Values.timeout | quote }}
```


```bash
rm -rf /tmp/example
mkdir -p /tmp/example/templates
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.timeout = 30
EOF
cat > templates/secret.yaml <<EOF
kind: Secret
stringData:
  timeout: {{ .Values.timeout | quote }}
EOF
kdo template .
```

#### Methods

You can also call methods from the go template

```python
# Chart.star
def name(self):
  return "my-name"
```
```yaml
kind: Secret
stringData:
  upper: {{ call .Methods.name }}
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example/templates
cd /tmp/example
cat > Chart.star <<EOF
def name(self):
  return "my-name"
EOF
cat > templates/secret.yaml <<EOF
kind: Secret
stringData:
  upper: {{ call .Methods.name }}
EOF
kdo template .
```



### Overriding methods

You can override the built in methods `apply`, `delete` and `template`. You can use the provided methods `__apply`, `__delete` and `__template` to call the base implementations


```yaml
# Chart.star
def template(self, glob="", k8s=None):
   print("Hello World")
   return self.__template(glob, k8s)
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def template(self, glob="", k8s=None):
   print("Hello World")
   return self.__template(glob, k8s)
EOF
kdo template .
```

## Deployment

You can choose between different deployment methods:
* kubectl 
* kapp
* helm (see [helm-subcharts](#helm-subcharts))

### kubectl

```bash
rm -rf /tmp/example
mkdir -p /tmp/example/ytt-templates
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  pass
EOF
cat > ytt-templates/secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: test
stringData:
  test: test
EOF
kdo apply .
kdo delete .
```
### kapp

```python
cat > Chart.star <<EOF
def apply(self, k8s):
  k8s.tool = "kapp"
  self.__apply(k8s)
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example/ytt-templates
cd /tmp/example
cat > Chart.star <<EOF
def apply(self, k8s):
  k8s.tool = "kapp"
  self.__apply(k8s)
EOF
cat > ytt-templates/secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: test
stringData:
  test: test
EOF
kdo apply .
kapp list -A
kdo delete .
```


## Properties

You can define properties, which can be set using the command lines. All items from `values.yaml` inside the (helm) chart are automatically converted to properties

```python
def init(self):
  self.timeout = property(default=30)
  self.docker_config = struct_property(password=property(),username=property(default="_json_key"))
```

```yaml
kind: Secret
stringData:
  timeout: #@ self.timeout
```

```bash
kdo template . --set timeout=60
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example/ytt-templates
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.timeout = property(default=30)
EOF
cat > ytt-templates/secret.yaml <<EOF
kind: Secret
stringData:
  timeout: #@ self.timeout
EOF
kdo template . --set timeout=60
```


## Associations

You can have 3 types of associations
* sub charts
* depedencies
* helm sub charts

### Subcharts

Subcharts can be used if another chart is owned (not shared) by the parent chart. You can set any property of the sub chart.

```python
def init(self):
  self.mysql = chart('helm://charts.helm.sh/stable/mysql')
  self.mysql.ssl.enabled = False
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.mysql = chart('helm://charts.helm.sh/stable/mysql')
  self.mysql.ssl.enabled = False
EOF
kdo template .
```

### Helm subcharts

This is mostly the same as the example above, except that `helm template` is used for templating and `helm upgrade -i` for application.

```python
def init(self):
  self.mysql = helm_chart('helm://charts.helm.sh/stable/mysql')
  self.mysql.ssl.enabled = False
```
### Dependencies

Dependencies can be used, if a subchart can be shared. If the dependency is not installed, it's automatically applied to the cluster.
You can also set properties of the sub chart. But this is not implemented correctly yet. 
Currently the properties are only taken into account, if the dependency is not already installed.


```python
def init(self):
  self.mysql = depends_on('helm://charts.helm.sh/stable/mysql',">= 1.0")
```

### Calling methods of associations

It's not only possible to set properties of associations. It's also possible to 
call methods. For associations created with `depends_on` this only possible
during `apply` or `delete`.

```python
def hello(self, name):
  print("Hello " + name)
```

```python
def init(self):
  self.__class__.version = "1.0.0"
  self.hello = depends_on('../hello',">= 0.0")
def apply(self,k8s):
  self.hello.hello("Kyma")
```

```bash
rm -rf /tmp/hello
mkdir -p /tmp/hello
cd /tmp/hello
cat > Chart.star <<EOF
def hello(self, name):
  print("Hello " + name)
EOF
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.__class__.version = "1.0.0"
  self.hello = depends_on('../hello',">= 0.0")
def apply(self,k8s):
  self.hello.hello("Kyma")
EOF
kdo apply /tmp/example
```

## URLs

| URL                                                               | Description                               |
|-------------------------------------------------------------------|-------------------------------------------|
| `./`                                                                | current directory                         |
| `helm://charts.helm.sh/stable/mysql`                                | latest helm chart in this helm repository |
| `helm://charts.helm.sh/stable/mysql`                                | latest helm chart in this helm repository |
| `https://github.com/<repo>/archive/<branch-or-tag>.zip`             | Github repository                         |
| `https://github.com/sap/kubernetes-deployment-orchestrator/archive/master.zip#charts/kdo` | Subdirectory in github repository         |
| `https://<host>/api/v3/repos/<owner>/<repo>/zipball/<branch>`       | Enterprise github repository              |
| `catalog:<chart>`                                                   | Chart from catalog (see below)            |

### Catalog

You can define a catalog, which contains all charts, you would like to be able to deploy. 
The chart name is appended to the configured catalog URL.
The catalog can be configured in `~/.kdo/config`. You can have many catalogs. They are tried in the given order.

```yaml
catalogs:
  - /Users/d001323/workspace/catalog
```

You can use the urls like

```bash
kdo template catalog:cluster-essentials
```

## Utilities
### User credentials

```python
def init(self):
  self.credential = user_credential('test')
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.credential = user_credential('test')
EOF
kdo template .
```

### Certificates

```python
def init(self):
  self.certificate = certificate('test',is_ca=True,domain="example.com")
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def init(self):
  self.certificate = certificate('test',is_ca=True,domain="example.com")
EOF
kdo template .
```

## kubernetes client

Kubernete deployment orchestrator includes a full featured kubernetes client which can be used in `apply`, `delete` or `template` actions

### Get

```python
def apply(self,k8s):
  print(k8s.get('service','kubernetes').status)
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def apply(self,k8s):
  print(k8s.get('service','kubernetes').status)
EOF
kdo apply .
```

### Watch

The `watch` loop will never end by default. You need to use `break` for this purpose.

```python
def apply(self,k8s):
  for service in k8s.watch('service','kubernetes'):
    print(service)
```

```bash
rm -rf /tmp/example
mkdir -p /tmp/example
cd /tmp/example
cat > Chart.star <<EOF
def apply(self,k8s):
  for service in k8s.watch('service','kubernetes'):
    print(service)
EOF
kdo apply .
```


### Apply

```python
def apply(self,k8s):
  service = k8s.get('service','kubernetes')
  service.metadata.labels.test = "test"
  k8s.apply(service)
```

### Delete


## Unit testing

There is an in memory implementation of the kubernetes client

```python
c = chart("../charts/example/simple/uaa")
c.apply(k8s)
uaa = k8s.get("statefulset","uaa-master")
assert.eq(uaa.metadata.name,"uaa-master")
assert.neq(uaa.metadata.name,"uaa-masterx")
```

```bash
kdo test test/*.star
```

## Controller

There is also a controller available, which reconciles installation. Currently it's broken.
## Extend

It's easy to extend kdo with you own DSL: https://github.com/wonderix/cfpkg

