# Using kdo controller

Charts can be also applied (in parts) using the kdo controller. Two proxy modes are support

* `local` the chart `CR` is applied to the same cluster.
* `remote` the chart `CR` is applied to the cluster where the current `k8s` value points to.

These modes only behave different, if you are applying charts to different clusters.

## Install kdo controller

```bash
kdo apply charts/kdo
```

## Install a kdo chart using the controller

```bash
kdo apply --proxy remote <chart>
```

or from inside another kdo chart

```python
def init(self):
  self.uaa = chart("uaa",proxy="local")
```
