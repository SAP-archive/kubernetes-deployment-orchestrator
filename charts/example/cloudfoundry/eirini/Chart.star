def init(self,blobstore=None):
  self.blobstore_credentials = blobstore.create_or_update_blob_store("eirini")
