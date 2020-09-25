load("@shalm:osb","osb")


def init(self):
    self.postgres = osb.binding("postgres",service="service",plan="plan",parameters={"test":"test"})
