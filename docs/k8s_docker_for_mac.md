# k8s @ Docker For Mac
* k8s will be running on local machine
* k8s support is only available in [Docker for Mac](https://docs.docker.com/docker-for-mac/install/) 17.12 CE Edge and higher, on the Edge channel.
* **Note:** support for Kubernetes in Docker For Mac is still early. There is a bug [https://github.com/docker/for-mac/issues/2445](https://github.com/docker/for-mac/issues/2445) that
  make NodePorts unusable. So you will be able to deploy Aptomi example apps successfully, but **will NOT be able to open any of their endpoints (e.g. HTTP)**.  

# Configuration
1. Once Docker for Mac is installed, ensure it has enough resources and enable k8s support:
    * Preferences -> Advanced -> CPUs=4, Memory=12.2 GiB 
    * Preferences -> Kubernetes -> Enable Kubernetes
    * It will take a few minutes to install and start a local k8s cluster
    * Once you see "Kubernetes is running", check that context `docker-for-desktop` has been created:
    ```
    kubectl config get-contexts
    ```   

2. Import it into Aptomi as two separate clusters *cluster-us-east* and *cluster-us-west* (corresponding to two namespaces `east` and `west` in a local k8s cluster):
    ```
    aptomictl gen cluster -c docker-for-desktop -n cluster-us-east -N east | aptomictl policy apply --username admin -f -
    aptomictl gen cluster -c docker-for-desktop -n cluster-us-west -N west | aptomictl policy apply --username admin -f -
    ```

Now you can move on to running the examples.

# Useful Commands

## View Pods
Once you deploy Aptomi example apps, you can run `kubectl` to get workloads running on each cluster: 
```
watch -n1 -d -- kubectl -n east get pods
watch -n1 -d -- kubectl -n west get pods
```
