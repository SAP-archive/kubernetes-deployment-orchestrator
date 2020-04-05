def init(self,local=False):
  if local:
    self.image_pull_policy= "Never"

def apply(self):
  self.__apply(glob="crd.yaml")
  self.k8s.wait("customresourcedefinition.apiextensions.k8s.io", "shalmcharts.wonderix.github.com","condition=established")
  self.__apply(glob="[^c][^r][^d]*.yaml")
