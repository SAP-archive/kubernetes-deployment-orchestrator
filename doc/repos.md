# Repositories

Shalm supports a huge set of repository types

* Directories
* Helm Repos
* Github Repos
* Github Releases

## Github Repos

It's possible to use github repositories directly as source for shalm charts.

```bash
shalm apply https://github.com/<repo>/archive/<branch-or-tag>.zip
shalm apply https://github.com/wonderix/cf-for-k8s/archive/shalm.zip  --set domain=cf.example.com
```

The zip file always contains a root directory, which always stripped off.

### Fragments

You can also specify a fragment to get only a part of a zip archive. 

```bash
shalm apply https://github.com/wonderix/shalm/archive/master.zip#charts/shalm
```

Normally, a zip file always contains a root folder. This root folder is always added to the path given in the fragment to ease the usage.


### Enterprise github repos

```bash
https://<host>/api/v3/repos/<owner>/<repo>/zipball/<branch>
```

### Download credentials

Download credentials can be configured in`$HOME/.shalm/config`. Example

```yaml
credentials:
  - url: https://<host>/
    token: 123j9iasdfj2j3412934
```
