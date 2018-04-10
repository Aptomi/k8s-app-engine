# Benefits

## For Developers
Aptomi gives developers the power to:
* __Keep using existing app packaging in CI__: Helm, k8s YAMLs, Ksonnet, and more!
* __Leverage intelligent app delivery through Aptomi in CD__: multi-container apps across multiple environments (dev, stage, prod) and k8s clusters.
* __Stand up new environments quickly__: use one CLI command to deploy the full application-- no need to deploy individual containers/charts by hand or hardcode variables!
* __Deliver individual updates to app components__: Aptomi knows which part of an application graph needs to be updated, so the impact of a change is known and disruption is minimized.
* __Benefit from lazy allocation of resources__: containers are only running when needed, so unused environments can be garbage-collected automatically.

## For Helm Chart Developers

* **Service discovery** allows you to glue Helm charts together into a larger application.
* **Well-defined dependencies** between application components allow Aptomi to build dependency graphs and calculate the impact of changes. 

## For Operators

Ops teams can plug into the same platform and have control over:

* **Managing Kubernetes clusters**
* **Injecting policies/rules** - access, security, and governance "as code."

Rule examples:
* **Access**: *Development* teams can never deploy to *Production*.
* **Security**: All instances of *nginx* in *Production* must use *HTTPS*.
* **Placement**: *Production Instances* get deployed to *us-west*, *Staging Instances* get deployed to *us-west*.
* **Service Reuse**: *Web* and *Mobile* teams always share the same service in *Staging*, while *Healthcare* team gets a dedicated *high-performance* instance.

# How Aptomi fits into CI/CD (Jenkins, Spinnaker)

* Aptomi sits between CI/CD and Kubernetes in the stack.
* Demos of integration with Spinnaker and Jenkins will be published soon!

# Why would I want to use Aptomi if I'm already implementing Kubernetes/OpenShift?

Kubernetes is fantastic when it comes to container orchestration, and the Kubernetes API/CLI is a standard that everyone is building upon.

However, everyone implementing Kubernetes ends up building their own custom layer around it for application delivery. For example:
* Using Helm, k8s YAML or ksonnet for app definition.
* Building service discovery/text templating to glue components together.
* Applying generated manifests using [kube-applier](https://github.com/box/kube-applier), [helm install](https://github.com/kubernetes/helm/blob/master/docs/using_helm.md#helm-install-installing-a-package), [ks apply](https://github.com/ksonnet/ksonnet/blob/master/docs/cli-reference/ks_apply.md) or [kubectl apply](https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/#kubectl-apply), based on what you have chosen.  
* Building app delivery pipelines, which have to properly account for running apps across multiple envs/clusters.
* Allowing for non-blocking collaboration of teams building their own pieces, including an ability to control every part of the process, from component updates and operational rules to security and other concerns.   
    
As more and more requirements pile up, this custom layer around Kubernetes becomes more and more complicated. Aptomi solves the challenging problems this custom layer presents by providing an open and intelligent application delivery platform for Developers and Operators within an organization.
