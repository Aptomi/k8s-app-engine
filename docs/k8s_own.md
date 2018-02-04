# Your own k8s cluster 
* Make sure your k8s/Docker has enough resources available to run examples:
    * At least 4 CPU cores and 12GB RAM

# Configuration
1. Check that your k8s cluster is configured properly and has **at least one context** defined:
   ```
   kubectl config get-contexts
   ```   
   
2. Import it into Aptomi as two separate clusters *cluster-us-east* and *cluster-us-west* (corresponding to two namespaces `east` and `west` in a local k8s cluster). Make
   sure to replace `[CONTEXT_NAME]` with the name of your context :
    ```
    aptomictl gen cluster -c [CONTEXT_NAME] -n cluster-us-east -N east | aptomictl policy apply --username admin -f -
    aptomictl gen cluster -c [CONTEXT_NAME] -n cluster-us-west -N west | aptomictl policy apply --username admin -f -
    ```

Now you can move on to running the examples.

# Useful Commands

## View Pods
Once you deploy Aptomi example apps, you can run `kubectl` to get workloads running on each cluster: 
```
watch -n1 -d -- kubectl -n east get pods
watch -n1 -d -- kubectl -n west get pods
```
