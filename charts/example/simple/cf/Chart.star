
def init(self,dummy="test"):
  self.mariadb = chart("../mariadb")
  self.mariadb.slave['replicas'] = 2
  self.uaa = chart("../uaa",database=self.mariadb,namespace="uaa")
  self.name = "my-first-chart"
  self.password = "test-pass"
  return self


def __secret_name(self):
  return "mysecret"

def apply(self, k8s):
  self.mariadb.apply(k8s)
  k8s.rollout_status("statefulset","mariadb-master")
  self.uaa.apply(k8s)
  self.__apply(k8s)