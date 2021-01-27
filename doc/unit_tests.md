## Testing


Tests can be written in starlark.  The following symbols are predefined

| Name                    | Description                                                                      |
|-------------------------|----------------------------------------------------------------------------------|
| `chart(url,...)`        | Function to load shalm chart. The `url` can be given relative to the test script |
| `k8s`                   | In memory implemention of k8s                                                    |
| `env(name)`             | Read environment variable                                                        |
| `assert.fail(msg)`      | Make test fail with given message                                                |
| `assert.true(cond,msg)` | Make test fail with given message if `cond` is false                             |
| `assert.eq(v1,v2)`      | Assert equals                                                                    |
| `assert.neq(v1,v2)`     | Assert not equals                                                                |

```python
c = chart("../charts/example/simple/uaa")
c.apply(k8s)
uaa = k8s.get("statefulset","uaa-master")
assert.eq(uaa.metadata.name,"uaa-master")
assert.neq(uaa.metadata.name,"uaa-masterx")
```

### Running tests

```bash
shalm test test/*.star
```
