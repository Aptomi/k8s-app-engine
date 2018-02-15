# Benefits

## For Developers

* Keep using existing app packaging in CI: *Helm, k8s YAMLs, Ksonnet, etc*
* Leverage intelligent app delivery through Aptomi in CD: *multi-container apps, across multiple envs (dev, stage, prod) and k8s clusters*
* Stand up new environments quickly: *use one CLI command to deploy the full application, no need to deploy individual containers/charts by hand and hardcode variables*
* Deliver individual updates to app components: *Aptomi knows which part of an application graph needs to be updated, so the impact of change is known and disruption is minimized*

## For Helm Chart Developers

* **Service discovery** allows to glue Helm charts together into a larger application
* **Well-defined dependencies** between applications components allow Aptomi to build a graph and calculate impact of changes 

## For Operators

Ops team, if exists in an organization, can plug into the same platform and have control over:

* **Managing Kubernetes clusters**
* **Injecting policies/rules** - access, security, governance "as code"

Rule examples:
* **Access**: *Development* teams can never deploy to *Production*
* **Security**: All instances of *nginx* in *Production* must use *HTTPS*
* **Placement**: *Production Instances* get deployed to *us-west*, *Staging Instances* get deployed to *us-west*
* **Service Reuse**: *Web* and *Mobile* teams always share the same service in *Staging*, while *Healthcare* team gets a dedicated *high-performance* instance

# How Aptomi fits into CI/CD (Jenkins, Spinnaker)

* Aptomi sits between CI/CD and Kubernetes in the stack
* Demos of integration with Spinnaker and Jenkins will be published soon

# Why would I want to use Aptomi if I'm already implementing Kubernetes/OpenShift

* Kubernetes is great when it comes to container orchestration. Kubernetes API/CLI became a standard that everyone is building upon
* However, everyone end up building their own custom layer around Kubernetes for application delivery:
    * Use Helm, k8s YAML or ksonnet for app definition
    * Build service discovery/text templating around to glue components together
    * Apply generated manifests using [kube-applier](https://github.com/box/kube-applier), [helm install](https://github.com/kubernetes/helm/blob/master/docs/using_helm.md#helm-install-installing-a-package), [ks apply](https://github.com/ksonnet/ksonnet/blob/master/docs/cli-reference/ks_apply.md) or [kubectl apply](https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/#kubectl-apply) based on what you have chosen  
    * Build app delivery pipelines, which have to properly account for running apps across multiple envs/clusters
    * Allow for non-blocking collaboration of teams building their own pieces. Including an ability to control every part of the process (component updates, operational rules, security, etc)   
    
As more and more requirements pile up, this custom layer around Kubernetes becomes more and more complicated. Aptomi solves exactly for that, providing an open intelligent application delivery platform for Developers and Operators within an organization.
