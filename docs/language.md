# Table of contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Concepts](#concepts)
- [Objects](#objects)
  - [ACL](#acl)
  - [Service](#service)
  - [Contract](#contract)
  - [Cluster](#cluster)
  - [Dependency](#dependency)
  - [Rule](#rule)
- [Common constructs](#common-constructs)
  - [Labels](#labels)
  - [Expressions](#expressions)
  - [Criteria](#criteria)
  - [Templates](#templates)
  - [Namespace references](#namespace-references)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Concepts
* A single Aptomi instance is called Aptomi **domain**
* There can be multiple **namespaces** within a domain, including a special `system` namespace
* Namespaces can have different **access rights** for different groups of users. So one typical use of namespaces would be to
  allocate one namespace (or a set of namespaces) for each team within an organization, where the corresponding team will have
  full control over defining its applications
* Every **object** defined in Aptomi must belong to a certain namespace

When defining any object in Aptomi, you must specify the essential metadata:
* kind - type of object
* metadata.namespace - which namespace the object is defined in
* metadata.name - name of the object, must be unique within the given namespace and given object kind
```yaml
- kind: service
  metadata:
    namespace: main
    name: wordpress
```

# Objects

## ACL

Before using Aptomi for the first time, it's required to set up [access control rights](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#ACLRule) for Aptomi **domain** and **namespaces** for different users/teams.

There are three built-in user roles in Aptomi:
* **domain admin** - has full access rights to all Aptomi namespaces, including `system` namespace
  * Domain admin can change global ACL, global list of rules and global list of clusters, which all reside in `system` namespace
  * Domain admin can define and publish services, contracts, dependencies, rules in any namespace
* **namespace admin** - has full access rights for a given list of Aptomi namespaces
  * Namespace admin can only view, but not manage objects in `system` namespace
  * Namespace admin can define and publish services, contracts, dependencies, rules in a given set of Aptomi namespaces
* **service consumer** - can only consume services in a given list of Aptomi namespaces
  * Service consumer has view only access for all namespaces
  * Service consumer can declare dependencies (and therefore consume services) in a given list of Aptomi namespaces

For example, this would promote all users with 'global_ops == true' label into domain admins:
```yaml
- kind: aclrule
  metadata:
    namespace: system
    name: domain_admins
  criteria:
    require-all:
      - global_ops
  actions:
    add-role:
      domain-admin: '*'
```

This would promote all users with 'is_operator == true' label into namespace admins for namespace 'main':
```yaml
- kind: aclrule
  metadata:
    namespace: system
    name: namespace_admins_for_main
  criteria:
    require-all:
      - is_operator
  actions:
    add-role:
      namespace-admin: main
```

This would make all users with 'org == dev' label into service consumers for namespace 'main':
```yaml
- kind: aclrule
  metadata:
    namespace: system
    name: service_consumers_for_main
  criteria:
    require-all:
      - org == 'dev'
  actions:
    add-role:
      service-consumer: main
```

## Service

[Service](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Service) is an entity that you would use to define structure of your application and its dependencies.

Service consists of one of more [components](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#ServiceComponent) which correspond to the components of your application.
Every component can be either a piece of code that needs to be deployed/instantiated, or a reference to another instantiated service. This way, you can construct your application from
the code that you own and also leverage services which could be owned by someone else in the organization.

Service can have also have labels attached to it. You can refer to those labels in expressions, typically when writing rules.

When defining a **code component**, you must define the following fields:
* `name` - component name unique within the service
* `code` - section which describes application component that needs to be instantiated and managed
    * `type` - right now the only supported code type is [aptomi/code/kubernetes-helm](https://helm.sh/), which is Helm package manager for k8s. But Aptomi is completely
      pluggable and allows developers to use their favorite framework for packaging applications. It can support applications manifests defined via ksonnet, k8s YAMLs and more.
* `discovery` - every component can expose arbitrary discovery information about itself in a form of labels to other components.
* `dependencies` - other components within the service, which current component depends on. It helps Aptomi to process discovery information and propagate parameters
  in the right order, as well as controls correct instantiation/destruction order of application components.

For example, here is how you would define an application which consists of Wordpress and MySQL database:
```yaml
- kind: service
  metadata:
    namespace: main
    name: wordpress

  labels:
    blog: true

  components:
    - name: wordpress_component
      code:
        type: aptomi/code/kubernetes-helm
        params:
          chartRepo: https://myhelmcharts.com/repo
          chartName: wordpress
          chartVersion: 4.9.1
          cluster: "{{ .Labels.cluster }}"
          db_url: "{{ .Discovery.mysql_component.url }}"
          db_name: "wp"
      dependencies:
        - mysql_component

    - name: mysql_component
      code:
        type: aptomi/code/kubernetes-helm
        params:
          chartRepo: https://myhelmcharts.com/repo
          chartName: mysql
          chartVersion: 5.7.21
          cluster: "{{ .Labels.cluster }}"
      discovery:
        url: "mysql-{{ .Discovery.instance }}:3306"
```

For Helm plugin, you need to provide the following parameters under "params" section in `code`, while the rest of the parameters will be passed "as is" to the instantiated Helm chart:
* `chartRepo` - URL of repository with Helm charts
* `chartName` - name of the Helm chart
* `chartVersion` - *(optional)* version of the Helm chart. if not specified, the latest version will be used
* `cluster` - name of the cluster to which the code will be deployed

Every parameter under "params" section can be a fixed value or an expression which can refer to various labels.

## Contract
Once a service is defined, it has to be exposed through a [contract](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Contract).

Contract represents a consumable service. When someone wants to consume an instance of a service, it has to be done through a contract. When
a service depends on another service instance, this dependency has to be dependency on a contract as well.

Contract has a number of contexts, which define different implementations of that contract. Each context allows you to define specific implementation and criteria under which it will be picked.
It means that a contract will be fulfilled via one of the defined contexts at runtime, based on the defined criteria. Different parameters/services can be picked based on
based on type of the environment (e.g. 'dev' vs. 'prod'), based on properties of a consumer (e.g. which 'team' consumer belongs to),
based on time of the day, or any other labels/properties. It can also control whether the service instance will be dedicated or shared.

For example, a team responsible for running databases can define a contract for "sql-database" and its specific implementations:
* 'dev' - for people from 'dev' team, which will be implemented via instantiating service 'sqlite'
* 'prod' - for people from outside of 'dev' team, which will be implemented via instantiating service 'mysql'
```yaml
- kind: contract
  metadata:
    namespace: main
    name: sql-database

  contexts:
    - name: dev
      criteria:
        require-all:
          - team == 'dev'
      allocation:
        service: sqlite

    - name: prod
      criteria:
        require-none:
          - team == 'dev'
      allocation:
        service: mysql
```

In that case, if the team who runs wordpress application wants to rely on an external database, they would change their service definition to
use a special **contract component**, which points to a contract instead of a code:
```yaml
- kind: service
  metadata:
    namespace: main
    name: wordpress

  labels:
    blog: true

  components:
    - name: wordpress_component
      code:
        type: aptomi/code/kubernetes-helm
        params:
          chartRepo: https://myhelmcharts.com/repo
          chartName: wordpress
          chartVersion: 4.9.1
          cluster: "{{ .Labels.cluster }}"
          db_url: "{{ .Discovery.db_component.database.url }}"
          db_name: "wp"
      dependencies:
        - db_component

    - name: db_component
      contract: sql-database
```

If you don't need the power of contracts and want your service to be instantiated directly, you can create a contract which maps 1-to-1 to
the corresponding service, while keeping the criteria empty (i.e. always true):
```yaml
- kind: contract
  metadata:
    namespace: main
    name: mysql

  contexts:
    - name: primary
      allocation:
        service: mysql
```

In order to control service dedication/sharing, one would use a special "keys" field inside an allocation.
For example, this would mean that each team will receive its own dedicated instance of MySQL:
```yaml
- kind: contract
  metadata:
    namespace: main
    name: mysql

  contexts:
    - name: primary
      allocation:
        service: mysql
        keys:
          - "{{ .User.Labels.Team }}"
```

When fulfilling a contract, Aptomi will process all contexts within that contract one by one and find the first matching context. Once context is selected, labels will be changed according to the `change-labels` section and service allocation will be done according to the corresponding "allocation" section within the context.

## Cluster

[Cluster](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Cluster) is an entity which defines a cluster in Aptomi where containers can be deployed. Even though Aptomi is focused on k8s, it's designed to support
multiple cluster types (e.g. Docker Swarm, Apache Mesos and others). Cluster type is defined via `type` attribute.

Clusters are global to Aptomi and must be always defined in `system` namespace.

A typical definition of k8s cluster looks like:
```yaml
- kind: cluster
  metadata:
    namespace: system
    name: cluster-us-east
  type: kubernetes
  config:
    kubeconfig:
      # put your kubeconfig for the cluster here
```

## Dependency

Defining a service and a contract only publishes a service into Aptomi, but it does not trigger instantiation/deployment of that service.
In order to request an instance, one must create a [Dependency](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Dependency) object in Aptomi.

When creating a dependency, one must specify `who` (or `what`) is making a request, which `contract` is being requested, and what set of initial labels is passed to Aptomi.

For example, here is how **Alice** would request a **wordpress**:
```yaml
- kind: dependency
  metadata:
    namespace: main
    name: alice_uses_wordpress
  user: Alice
  contract: wordpress
  labels:
      label1: value1
```

Since Aptomi rules all label-based, you can create a policy to make intelligent decisions based on the initial set of labels being passed, as well as transform those labels.

## Rule

One of the most powerful features of Aptomi is ability to define [rules](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Rule), which get evaluated in runtime during state enforcement.

Rules can be **global** (defined in `system` namespace) as well as **local** (defined within a given namespace).

The order in which the rules are being processed and executed is:
* local rules - one by one for a given namespace, sorted by weight
* global rules - one by one, sorted by weight

A rule has a criteria and an action. If a criteria evaluates to true, then an action is executed. The list of supported actions is:
* change-labels - change one or more labels
* dependency - reject dependency and not allow instantiation

A typical and most commonly used rule action in Aptomi is to change a label. For example, by changing a system-level label called `cluster`, you can control into which cluster the code will get deployed to. Deploying
code without setting `cluster` label will result in an error, because Aptomi won't have a way of knowing where the code should be deployed.

For example, this rule will tell Aptomi to always deploy services with **blog** label to cluster **cluster-us-east**:
```yaml
- kind: rule
  metadata:
    namespace: main
    name: blog_services_get_deployed_to_us_east
  weight: 10
  criteria:
    require-all:
      - service.Labels.blog == true
  actions:
    change-labels:
      set:
        cluster: cluster-us-east
```

Here is another rule that will not allow users from **dev** team to instantiate any **blog** services. Even if they try to declare a dependency, it will not be fulfilled by Aptomi:
```yaml
- kind: rule
  metadata:
    namespace: main
    name: dev_teams_cannot_instantiate_blog_services
  weight: 20
  criteria:
    require-all:
      - team == 'dev'
      - service.Labels.blog == true
  actions:
    dependency: reject
```

# Common constructs
## Labels
Aptomi policy processing is based entirely on labels. When a dependency is requested, an initial set of labels is formed by combining labels of the requester (e.g. user labels) and a given dependency. Throughout processing,
these labels can be transformed using `change-labels` directives and accessed at any time in expressions and templates.

One way of leveraging this, for example, would be to:
* define a service, contract and two contexts
* set label 'replicas' to value '1' in dev context
* set label 'replicas' to value '3' in production context
* use label 'replicas' in the service to configure it appropriately
* that way Aptomi would be able to provision different instances of the same service with different settings

Here is how one would use [change-labels](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#LabelOperations) to achieve that:
```yaml
- kind: contract
  metadata:
    namespace: main
    name: myservice

  contexts:
    - name: dev
      criteria:
        require-all:
          - team == 'dev'
      change-labels:
        set:
          replicas: 1
      allocation:
        service: myservice

    - name: prod
      criteria:
        require-none:
          - team == 'dev'
      change-labels:
        set:
          replicas: 3
      allocation:
        service: myservice
```

## Expressions
All expressions used in Aptomi should follow [Knetic/govaluate](https://github.com/Knetic/govaluate) syntax and must evaluate to bool.

You can reference the following variables in expressions:
* labels - you can reference any label by specifying its name, e.g. `team` will return a value of a label with name 'team'
* service - you can reference a service which is currently being processed. it's an object, so you can go down and look into its properties, e.g. `service.Name` or `service.Labels.blog`

## Criteria
[Criteria](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Criteria) allows to define complex matching expressions in the policy.
It supports `require-all`, `require-any` and `require-none` sections, with a list of expressions under each section.

Criteria gets evaluated to true only when:
* All `require-all` expressions evaluate to true
* At least one of `require-any` expressions evaluates to true
* None of `require-none` expressions evaluate to true

If a section is absent, it will be skipped. So it's perfectly fine to have a criteria with fewer than 3 sections (e.g. with just `require-all`) or with no sections at all.

It's also possible to have an empty criteria without any clauses (or even omit `criteria` construct all together). In this case an empty criteria is always considered to be 'true'.

## Templates
All text templates used in Aptomi should follow [text/template](https://golang.org/pkg/text/template/) syntax and must evaluate to string.

The most common use of text templates in Aptomi is code & discovery parameters inside a service:
* **code parameters** - allow to pass parameters into Helm charts, substituting variables with calculated label values
* **discovery parameters** - allow components to expose their discovery parameters to other services/components, substituting variables with calculated label values

You can reference the following variables in text templates:

* `{{ .Labels }}` - the current set of labels, e.g.:
  * `{{ .Labels.labelName }}` will return the value of label with name `labelName`
  * `{{ .Labels.cluster }}` will return the special `cluster` label, which will indicate the name of the cluster in `system` namespace to which the code will get deployed to
* `{{ .User}}` - the current user who requested a dependency
  * `{{ .User.Name }}` - name of the user
  * `{{ .User.Secrets }}` - a map of user secrets
  * `{{ .User.Labels }}` - a map of user labels
* `{{ .Discovery }}` - a set of discovery parameters
  * `{{ .Discovery.instance }}` - a unique human-readable deployment name of the current component instance to be deployed
  * `{{ .Discovery.instanceid }}` - a unique hash of the current component instance to be deployed
  * `{{ .Discovery.service.instanceid }}` - a unique hash of the current service instance to be deployed
  * `{{ .Discovery.component1.[...].componentN.propertyName }}` - you can traverse component graph to get the value of 'propertyName' from discovery properties exposed by an particular component

## Namespace references
Sometimes you will want to specify an absolute path to an object located in a different namespace.

For example, this would mean that Aptomi will look for wordpress contract defined in the same namespace (i.e. 'main'):
```yaml
- kind: dependency
  metadata:
    namespace: main
    name: alice_uses_wordpress
  user: Alice
  contract: wordpress
```

While this would mean that Aptomi will look for wordpress contract in a given namespace (i.e. 'specialns'):
```yaml
- kind: dependency
  metadata:
    namespace: main
    name: alice_uses_wordpress
  user: Alice
  contract: specialns/wordpress
```

You can use the same syntax to create a service, which depends on a contract published in a different namespace (i.e. 'dbns'), by other team:
```yaml
- kind: service
  metadata:
    namespace: main
    name: wordpress

  labels:
    blog: true

  components:
    - name: wordpress_component
      code:
        ...
      dependencies:
        - db_component

    - name: db_component
      contract: dbns/sql-database
```
