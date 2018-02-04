# k8s @ Google Kubernetes Engine
* Use k8s @ Google Kubernetes Engine
* Aptomi comes with a script that can create k8s clusters @ GKE for you

# Configure Google Cloud SDK
1. Open [Google Cloud Console](https://console.cloud.google.com/)
    * Create new project with any name
    * API Manager -> Enable API
        * Google Container Engine API
        * Google Compute Engine API
1. Install Google Cloud SDK
    * ```curl https://sdk.cloud.google.com | bash```
1. Authenticate
    * ```gcloud auth login```
1. Set your project ID
    * ```gcloud config set project <YOUR_PROJECT_ID>```
    
# Create Clusters
1. Run the provided script, it will create 2 k8s clusters via GKE API:
    * ```
      curl https://raw.githubusercontent.com/Aptomi/aptomi/master/tools/demo-gke.sh | bash /dev/stdin up
      ```
      
1. Generate YAMLs for these k8s clusters and upload them to Aptomi as *cluster-us-east* and *cluster-us-west*:
    * ```
      aptomictl gen cluster -c cluster-us-east | aptomictl policy apply --username admin -f -
      aptomictl gen cluster -c cluster-us-west | aptomictl policy apply --username admin -f -
      ```

Now you can move on to running the examples.

# Useful Commands

## Destroy Clusters
After you are done with examples, it's a good idea to destroy the clusters so you don't continue to spend money with GKE: 
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/tools/demo-gke.sh | bash /dev/stdin down
```  

## Clean up Clusters
If you want to delete all workloads from the clusters and start running examples from scratch, use can use:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/tools/demo-gke.sh | bash /dev/stdin cleanup
```

## View Pods
Once you deploy Aptomi example apps, you can run `kubectl` to get workloads running on each cluster: 
```
watch -n1 -d -- kubectl --context cluster-us-east -n demo get pods
watch -n1 -d -- kubectl --context cluster-us-west -n demo get pods
```
