# k8s @ Google Kubernetes Engine
* Use k8s @ Google Kubernetes Engine
* Aptomi comes with a script that can create k8s cluster @ GKE for you

# Configure Google Cloud SDK
1. Open [Google Cloud Console](https://console.cloud.google.com/)
    * Create a new project
    * APIs & Services -> Dashboard -> Enable APIs and Services
        * Google Kubernetes Engine API -> Enable
        * Google Compute Engine API -> Enable
1. Install Google Cloud SDK
    * ```curl https://sdk.cloud.google.com | bash```
1. Authenticate
    * ```gcloud auth login```
1. Set your project ID
    * ```gcloud config set project <YOUR_PROJECT_ID>```
    
# Create Cluster
1. Run the provided script, it will create k8s cluster via GKE API:
    ```
    curl https://raw.githubusercontent.com/Aptomi/aptomi/master/tools/demo-gke.sh | bash /dev/stdin up
    ```
      
2. Import it into Aptomi as two separate clusters *cluster-us-east* and *cluster-us-west* (corresponding to two namespaces `east` and `west` in a local k8s cluster):
    ```
    aptomictl login -u admin -p admin
    aptomictl gen cluster -n cluster-us-east -c demo-gke -N east | aptomictl policy apply -f -
    aptomictl gen cluster -n cluster-us-west -c demo-gke -N west | aptomictl policy apply -f -
    ```

Now you can move on to running the examples.

# Useful Commands

## Destroy Cluster
After you are done with examples, it's a good idea to destroy the cluster so you don't continue to spend money with GKE: 
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/tools/demo-gke.sh | bash /dev/stdin down
```  

## Clean up Cluster
If you want to delete all workloads from the cluster and start running examples from scratch, use can use:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/tools/demo-gke.sh | bash /dev/stdin cleanup
```
