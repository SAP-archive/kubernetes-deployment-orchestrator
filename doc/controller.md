# Using shalm controller

Charts can be also applied (in parts) using the shalm controller. Two proxy modes are support

* `local` the chart `CR` is applied to the same cluster.
* `remote` the chart `CR` is applied to the cluster where the current `k8s` value points to.

These modes only behave different, if you are applying charts to different clusters.

## Install shalm controller

```bash
shalm apply charts/shalm
```

## Install a shalm chart using the controller

```bash
shalm apply --proxy remote <chart>
```

or from inside another shalm chart

```python
def init(self):
  self.uaa = chart("uaa",proxy="local")
```
