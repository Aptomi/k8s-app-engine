# k8s @ Docker For Mac
* k8s will be running on local machine
* k8s support is only available in [Docker for Mac](https://docs.docker.com/docker-for-mac/install/) 17.12 CE Edge and higher, on the Edge channel.
* **Caveat:** support for Kubernetes in Docker For Mac is still early. There is an issue [https://github.com/docker/for-mac/issues/2445](https://github.com/docker/for-mac/issues/2445) with
  NodePorts. You will be able to deploy Aptomi example apps successfully, but keep in mind that **application endpoints will be accessible via localhost instead of 192.168.65.3 IP advertised by Docker For Mac**.  

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
