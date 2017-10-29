If you don't have k8s clusters handy, the easiest way to run examples would be to create 2 k8s clusters with Helm
on Google Container Engine (GKE), where Aptomi will deploy all apps/services to. Aptomi comes with a script which
can create those k8s clusters for you automatically.

### Creating k8s clusters
1. Install Google Cloud SDK - ```curl https://sdk.cloud.google.com | bash```
1. Create new project in [Google Cloud Console](https://console.cloud.google.com/)
1. Authenticate - ```gcloud auth login```
1. Set your project ID - ```gcloud config set project <YOUR_PROJECT_ID>```
1. Open [Google Cloud Console](https://console.cloud.google.com/) -> API Manager -> Enable API
  1. Google Container Engine API
  1. Google Compute Engine API
1. Run ```./tools/gke-demo.sh up```. It will automatically create 2 k8s clusters *cluster-us-west* and *cluster-us-west*
with Helm installed and create configs for kubectl  

### Checking status of k8s clusters
1. ```./tools/gke-demo.sh status```

To view pods via kubectl:
```
watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods
watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods
```

### Cleaning up k8s clusters
1. ```./tools/gke-demo.sh cleanup```

Or, via kubectl directly:

```
kubectl --context cluster-us-west -n demo delete ns demo
kubectl --context cluster-us-east -n demo delete ns demo
```

### Destroying k8s clusters
Don't forget to destroy your clusters after running the examples, so you won't continue to get billed for them
1. ```./tools/gke-demo.sh down``` 
