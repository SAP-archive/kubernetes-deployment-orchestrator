def init(self):
    pass

def apply(self,k8s):
    self.__apply(k8s)
    for nginx in k8s.watch("deployment",'nginx'):
        if nginx.status and nginx.status.get('readyReplicas') == 3:
            print("")
            print("Installation successful")
            break
        else:
            print("Waiting for nginx to come up")
    x = k8s.get("deployment",'nginx')

