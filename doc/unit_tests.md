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
c = chart("../charts/example/simple/mariadb")
c.apply(k8s)
mariadb = k8s.get("statefulset","mariadb-master")
assert.eq(mariadb.metadata.name,"mariadb-master")
assert.neq(mariadb.metadata.name,"mariadb-masterx")
```

### Running tests

```bash
shalm test test/*.star
```
