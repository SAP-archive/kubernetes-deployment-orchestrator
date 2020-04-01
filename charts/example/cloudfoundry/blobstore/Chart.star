def init(self):
  pass

def create_or_update_blob_store(self,bucket="bucket"):
  return struct(user_credential=user_credential("blobstore-user-" + bucket))
