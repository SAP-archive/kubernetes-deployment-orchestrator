# Tutorial

This tutorial will guide you through the different topics of shalm

## Interoperability with helm

Shalm is designed from ground up to be mostly compatible with helm packages



### Helm to shalm

You deploy almost every helm chart using *shalm*. But *shalm* doesn't support hooks. 

```
shalm apply helm://charts.helm.sh/stable/mysql
kubectl get pods
shalm list -A
shalm delete helm://charts.helm.sh/stable/mysql

```

### Shalm to helm

Shalm chart can be wrapped into helm charts using `shalm package --helm`. Shalm uses `pre-upgrade` hooks to implement this.


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
shalm package --helm .
helm upgrade -i example example-0.0.1.tgz
helm uninstall example
```


## Templating

In shalm, you can chose between go templating and [ytt](https://get-ytt.io/)

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
shalm template .
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
shalm template .
```


### ytt templating


You should prefer ytt templating because ytt also uses starlark. Therefore the interoperability between shalm and ytt is much better compared to helm templating.

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
shalm template .
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
shalm template .
```

## Deployment

You can choose between different deployment methods:
* kubectl 
* kapp

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
shalm apply .
shalm delete .
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
shalm apply .
kapp list -A
shalm delete .
```


## Properties

You can define properties, which can be set using the command lines. All items from `values.yaml` inside the (helm) chart are automatically converted to properties

```python
def init(self):
  self.timeout = property(default=30)
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
  self.timeout = property(default=30)
EOF
cat > ytt-templates/secret.yaml <<EOF
kind: Secret
stringData:
  timeout: #@ self.timeout
EOF
shalm template . --set timeout=60
```


## Associations
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
shalm template .
```

### Helm subcharts

This is mostly the same as the example above, except that `helm template` is used for templating and `helm upgrade -i` fro application.

```python
def init(self):
  self.mysql = helm_chart('helm://charts.helm.sh/stable/mysql')
  self.mysql.ssl.enabled = False
```
### Dependencies

Dependencies can be used, if a subchart can be shared. If the dependency is not installed, it's automatically applied to the cluster.
You can also set properties of the sub chart. But this is not implemented correctly yet.

```python
def init(self):
  self.mysql = depends_on('helm://charts.helm.sh/stable/mysql',">= 1.0")
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
shalm template .
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
shalm template .
```

## kubernetes client

Shalm includes a full featured kubernetes client which can be used in `apply`, `delete` or `template` actions

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
shalm apply .
```

### Watch

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
shalm apply .
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
c = chart("../charts/example/simple/mariadb")
c.apply(k8s)
mariadb = k8s.get("statefulset","mariadb-master")
assert.eq(mariadb.metadata.name,"mariadb-master")
assert.neq(mariadb.metadata.name,"mariadb-masterx")
```

```bash
shalm test test/*.star
```

## Controller

There is also a controller available, which reconciles installation. Currently it's broken.
## Extend

It's easy to extend shalm with you own DSL: https://github.com/wonderix/cfpkg

