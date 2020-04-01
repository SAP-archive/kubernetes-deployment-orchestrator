def init(self):
  # Load chart from github
  self.pg_operator = chart("https://github.com/zalando/postgres-operator/archive/v1.4.0.zip#charts/postgres-operator")
  # Configure the postgres operator to use CRDs. This loads a yaml located in the same directory as the chart
  self.pg_operator.load_yaml("values-crd.yaml")
  # Configure the AWS region
  # You can modify any values of a helm chart
  self.pg_operator.configAwsOrGcp.aws_region =  "eu-central-1" 
