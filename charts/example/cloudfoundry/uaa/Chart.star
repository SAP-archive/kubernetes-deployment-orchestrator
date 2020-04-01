def init(self,database=None,ingress=None,logging=None):
  self.database_credentials = database.create_or_update_database("uaa",gigabytes=100)
  self.logging_credentials = logging.create_or_update_log_channel("uaa")
  ingress.create_or_update_route("uaa","uaa." + self.namespace + "svc.cluster.local")
