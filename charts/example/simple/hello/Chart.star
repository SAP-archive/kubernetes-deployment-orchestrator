
def init(self,arg='world'):
    self.__class__.name = "hello"
    self.__class__.maintainers = [ {"name" : "kramer" } ]
    self.message = "Hello " + arg

