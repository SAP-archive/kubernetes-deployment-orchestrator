
c = chart("../charts/example/simple/mariadb")
c.apply(k8s)

print(env("HOME"))
mariadb = k8s.get("statefulset","mariadb-master")

assert.eq(mariadb.metadata.name,"mariadb-master")

assert.neq(mariadb.metadata.name,"mariadb-masterx")