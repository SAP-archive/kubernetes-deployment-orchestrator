def init(self):
    pass

def apply(self):
    self.__apply()
    for nginx in self.k8s.watch("deployment",'nginx'):
        if nginx.status and nginx.status.get('readyReplicas') == 3:
            print("")
            print("Installation successful")
            break
        else:
            print("Waiting for nginx to come up")

