def init(self):
   self.nats = chart("https://charts.bitnami.com/bitnami/nats-4.2.6.tgz")
   self.auth = user_credential("nats-auth")
   self.cluster_auth = user_credential("nats-cluster-auth")

def credentials(self):
  return struct(user_credential=self.auth,port=self.nats.client["service"]["port"])


def apply(self,k8s):
  self.__apply(k8s)
  self.nats.auth["user"] = self.auth.username
  self.nats.auth["password"] = self.auth.password
  self.nats.clusterAuth["user"] = self.cluster_auth.username
  self.nats.clusterAuth["password"] = self.cluster_auth.password
  self.nats.apply(k8s)
