# Cassandra

# Deploy chart
```console
$ helm repo add mirantisworkloads https://mirantisworkloads.storage.googleapis.com
$ helm install mirantisworkloads/cassandra
```

# You can get your cassandra cluster status by running the command
``console
 kubectl exec -it {pod_name} nodetool status
