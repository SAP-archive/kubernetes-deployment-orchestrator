def init(self,database=None,ingress=None,logging=None,uaa=None,blobstore=None,eirini = None):
  self.database_credentials = database.create_or_update_database("capi","100GB")
  self.logging_credentials = logging.create_or_update_log_channel("capi")
  self.blobstore_credentials = blobstore.create_or_update_blob_store("capi")
  ingress.create_or_update_route("capi","capi." + self.namespace + "svc.cluster.local")
  self.uaa = uaa
