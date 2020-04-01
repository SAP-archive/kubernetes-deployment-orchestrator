load("@extension:message","message")
load("@extension:myvault","myvault")



def init(self):
  # Prints "hello world"
  print(message)
  self.vault = myvault("name","prefixxxx")
