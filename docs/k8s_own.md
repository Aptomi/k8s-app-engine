# Your own k8s cluster 
* Make sure your k8s/Docker has enough resources available to run our examples:
    * At least 4 CPU cores and 12GB RAM

# Configuration
1. Check that your k8s cluster is configured properly and has **at least one context** defined:
    ```
    kubectl config get-contexts
    ```   
   
2. Import it into Aptomi under name `k8s-example`, making sure to replace `[CONTEXT_NAME]` with the name of your context:
    ```
    aptomictl login -u admin -p admin
    aptomictl gen cluster -n k8s-example -c [CONTEXT_NAME] | aptomictl policy apply -f -
    ```

# Next Steps
You are now ready to run our examples!

Example    | Description
-----------|------------
[twitter-analytics](../examples/twitter-analytics) | Twitter Analytics Application, multiple services, multi-cloud, based on Helm
[guestbook](../examples/guestbook) | K8S Guestbook Application, multi-cloud, based on K8S YAMLs
