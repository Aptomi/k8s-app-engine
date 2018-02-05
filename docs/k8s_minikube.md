# k8s @ Minikube
* k8s will be running on local machine in [Minikube](https://github.com/kubernetes/minikube)

# Configuration
1. Start Minikube
    * For example, on Mac OS (pass params to ensure Minikube has enough resources, use xhyve):
      ```
      minikube start --cpus 4 --memory 12288 --vm-driver xhyve
      ```
    * Once it's up, check that context `minikube` has been created:
      ```
      kubectl config get-contexts
      ```   
   
2. Import it into Aptomi as two separate clusters *cluster-us-east* and *cluster-us-west* (corresponding to two namespaces `east` and `west` in a local k8s cluster):
    ```
    aptomictl gen cluster -c minikube -n cluster-us-east -N east | aptomictl policy apply --username admin -f -
    aptomictl gen cluster -c minikube -n cluster-us-west -N west | aptomictl policy apply --username admin -f -
    ```

Now you can move on to running the examples.
