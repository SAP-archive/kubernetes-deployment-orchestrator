load("@ytt:base64","base64")

def init(self):
  self.ca = certificate("ca",is_ca=True,validity="P10Y",domains=["ca.com"])
  self.cert = certificate("server",signer=self.ca,domains=["example.com"],validity="P1Y")

