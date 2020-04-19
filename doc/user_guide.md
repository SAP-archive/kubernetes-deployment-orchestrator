# User Guide

## Writing charts

Just follow the rules of helm to write charts. Additionally, you can put a `Chart.star` file in the charts folder

```bash
<chart>/
├── Chart.yaml
├── values.yaml
├── Chart.star
└── templates/
```

### Using embedded ytt yaml templates

You can use ytt yaml templates to render kubernetes artifacts. You simply put them in the any folder inside a chart.
There is currently no support for `data`, `star` or `text` files. The only value supplied to the templates is `self`,
which is the current chart. You can access all values and methods within your chart. To use this feature, you need to override the template method

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: #@ self.namespace
```

```python
def template(self,glob=''):
  return self.eytt("ytt",glob=glob)  # Use ytt templating with templates in directory 'ytt'
```

### Using full featured ytt yaml templates

To use this feature, you need to override the template method

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: #@ self.namespace
```

```python
def template(self,glob=''):
  return self.ytt("yttx",self.helm())  # Use ytt templating with templates in directory 'yttx' feeding in output from another helm template
```

## Packaging charts

You can package `shalm` charts using the following command:

```bash
shalm package <shalm chart>
```

It's also possible to convert `shalm` charts to `helm` charts:

```bash
shalm package --helm <shalm chart>
```

In this case, the helm chart only includes two jobs (`post-install` and `pre-delete` hooks) which do the whole work.

## kapp Support

Shalm charts can be applied/deleted using kapp. Therefore, you can pass `--tool kapp` at the command line.

## Examples

### Share database

The following example shows how a database manager could be shared.

* Define an API for a database manager (e.g. mariadb)

```python
def create_database(self,db="db",username="",password=""):
   ...
```

* Define a constructor for a service, which requires a database

```python
def init(self,database=None):
  if database:
    database.create_database(db="uaa",username="uaa",password="randompass")
```

* Use the API within another chart

```python
def init(self):
  self.mariadb = chart("mariadb")
  self.uaa = chart("uaa",database = self.mariadb)
```

### Override apply, delete or template

With `shalm` it's possible to override the `apply`, `delete` and `template` methods. The following example illustrates how this could be done

```python
def init(self):
  self.mariadb = chart("mariadb")
  self.uaa = chart("uaa",database = self.mariadb)

def apply(self,k8s):
  self.mariadb.apply(k8s) # Apply mariadb stuff (recursive)
  k8s.rollout_status("statefulset","mariadb-master")  # Interact with kubernetes
  self.uaa.apply(k8s)     # Apply uaa stuff (recursive)
  self.__apply(k8s)       # Apply everthing defined in this chart (not recursive)

def template(self,glob=''):
  return self.helm(glob=glob)  # Use helm templating (default)
```

### Jewels

Shalm provides the concept of jewels to store things like

* certificates
* user credentials

with the help of secrets in kubernetes.

It's also possible to extend shalm to provide other types of jewels:

* AWS users
* GCP users
* letsencrypt certificates
* ...

#### Create User Credentials

User credentials are used to manage username and password pairs. They are mapped to kubernets `Secrets`.
If the secret doesn't exist, the username and password are created with random content, otherwise the fields are
read from the secret. The keys used to store the username and password inside the secret can be modified.

The content of username and password can only be accessed after the call to `__apply`.
Therefore, you need to override the `apply` method.

All user credentials created inside a `Chart.star` file are automatically applied to kubernetes.
If you run `shalm template`, the content of the username and password is undefined.

```python
def init(self):
   self.nats = chart("https://charts.bitnami.com/bitnami/nats-4.2.6.tgz")
   self.auth = user_credential("nats-auth")

def apply(self,k8s):
  self.__apply(k8s)
  self.nats.auth["user"] = self.auth.username
  self.nats.auth["password"] = self.auth.password
  self.nats.apply(k8s)
```

#### Create Certificates

Shalm provides creation on self signed certificactes out of the box. These certificates can be used for
* Mutual TLS within your application
* Register k8s webhooks

The certificates are stored as secrets inside kubernetes. If you deploy your application again, the certificates will not be changed. Certificate rotation or renewal is not implemented yet. For this purpose you need to remove the secrets manually from kubernetes.

The following example will deploy 3 artifacts to kubernetes

* A secret `ca`, which containts your ca certificate
* A secret `server`, which containts your server certificate
* A configmap `configmap`, which containts the ca of your server certificate

```python
def init(self):
  self.ca = certificate("ca",is_ca=True,validity="P10Y",domains=["ca.com"]) # Create CA
  self.cert = certificate("server",signer=self.ca,domains=["example.com"],validity="P1Y")

def template(self,glob=""):
  return self.eytt("eytt") # use embedded ytt for templating
```

Put this template info `eytt/configmap.yml`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: configmap
data:
  #! render ca of server certificate
  ca: #@ self.cert.ca
```


## Extending and embedding shalm

It's possible to extend shalm with starlark modules. See `examples/extension` directory for details.
