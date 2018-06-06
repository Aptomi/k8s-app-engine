# k8s @ Minikube
* k8s will be running on a local machine in [Minikube](https://github.com/kubernetes/minikube)

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
   
2. Import it into Aptomi under name `k8s-example`:
    ```
    aptomictl login -u admin -p admin
    aptomictl gen cluster -n k8s-example -c minikube | aptomictl policy apply -f -
    ```
# Next Steps
You are now ready to run our examples!

Example    | Description
-----------|------------
[twitter-analytics](../examples/twitter-analytics) | Twitter Analytics Application, multiple services, multi-cloud, based on Helm
[guestbook](../examples/guestbook) | K8S Guestbook Application, multi-cloud, based on K8S YAMLs