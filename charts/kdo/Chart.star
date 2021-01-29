def init(self,local=False):
  if local:
    self.image_pull_policy= "Never"

def apply(self,k8s):
  self.__apply(k8s,glob="crd.yaml")
  k8s.wait("customresourcedefinition.apiextensions.k8s.io", "kdocharts.sap.github.com","condition=established")
  self.__apply(k8s,glob="[^c][^r][^d]*.yaml")
