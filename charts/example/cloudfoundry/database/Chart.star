def init(self):
   self.postgres = chart("https://charts.bitnami.com/bitnami/postgresql-ha-1.1.0.tgz")

def create_or_update_database(self,db="db",gigabytes=10):
  print("Create or update database " + db)
  return struct(user_credential=user_credential("postgresql-user-" + db) , hostname="postgresql." + self.namespace + ".svc.cluster.local",port="5432")
