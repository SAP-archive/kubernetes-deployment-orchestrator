
def init(self,database=None):
  self.name = "test"
  if database:
    database.create_database(db="uaa",username="uaa",password="87612349234")
  # self.ca = certificate("ca")
  # self.server_cert = certificate("uaa",signer=self.ca,dns_names=["example.com"],expiry=)
  return self

