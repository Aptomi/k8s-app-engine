# Table of contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Concepts](#concepts)
- [Objects](#objects)
  - [ACL](#acl)
  - [Bundle](#bundle)
  - [Service](#service)
  - [Cluster](#cluster)
  - [Claim](#claim)
  - [Rule](#rule)
- [Common constructs](#common-constructs)
  - [Labels](#labels)
  - [Expressions](#expressions)
  - [Criteria](#criteria)
  - [Templates](#templates)
  - [Namespace references](#namespace-references)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Concepts
* A single Aptomi instance is called an Aptomi **domain**.
* There can be multiple **namespaces** within an Aptomi domain, including a special `system` namespace.
* Namespaces can have different **access rights** for different groups of users. A common use of namespaces is to
  allocate one namespace (or set of namespaces) for each team within an organization, where the corresponding team has
  full control over how it defines its applications.
* Every **object** defined in Aptomi must belong to a certain namespace.

When defining any object in Aptomi, you must specify the following essential metadata:
* `kind` - The type of object
* `metadata.namespace` - Which namespace the object is defined in
* `metadata.name` - The name of the object, which must be unique within the given namespace and the given object kind

The following example illustrates a simple definition of a wordpress `bundle` object in the `main` namespace.
```yaml
- kind: bundle
  metadata:
    namespace: main
    name: wordpress
```

# Objects

## ACL

Before using Aptomi for the first time, you must set up [access control rights](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#ACLRule) for the Aptomi **domain**, and **namespaces** for different users/teams.

There are three built-in user roles in Aptomi:
* **domain admin** - Has full access rights to all Aptomi namespaces, including the `system` namespace
  * Domain admins can change the global ACL, the global list of rules, and the global list of clusters, which all reside in the `system` namespace.
  * Domain admins can define and publish bundles, services, claims, and rules in any namespace.
* **namespace admin** - Has full access rights for a given list of Aptomi namespaces
  * Namespace admins can view, but not manage objects in the `system` namespace.
  * Namespace admins can define and publish bundles, services, claims, and rules in a given set of Aptomi namespaces.
* **service consumer** - Can only consume services in a given list of Aptomi namespaces
  * Service consumers have view-only access for all namespaces.
  * Service consumers can only define and declare claims, and therefore consume services, in a given list of Aptomi namespaces.

For example, the following YAML block would promote all users with the `global_ops == true` label into domain admins:
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

This example would promote all users with the `is_operator == true` label into namespace admins for the `main` namespace:
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

Finally, this example would make all users with the `org == dev` label into service consumers for the `main` namespace:
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

## Bundle

A [Bundle](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Bundle) is an entity that you would use to define the structure of your application and its dependencies.

Bundles consist of one of more [components](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#BundleComponent) which correspond to the components of your application.
Every component can be either a piece of code that needs to be deployed/instantiated or a reference to another service. This way, you can construct your application from
the code that you own, and leverage services which may be owned by someone else in the organization.

A Bundle can have also have labels attached to it. You can refer to those labels with expressions, which is a practice typically employed when writing rules.

When defining a **code component**, you must define the following fields:
* `name` - The component name, which must be unique within the bundle
* `code` - The section which describes the application component that needs to be instantiated and managed
    * `type` - Right now, the only supported code type is [helm](https://helm.sh/), which is the Helm package manager for k8s. However, Aptomi is completely
      pluggable, and allows developers to use their favorite framework for packaging applications. Aptomi can support applications manifests defined via ksonnet, k8s YAMLs, and more!
* `discovery` - Every component can expose arbitrary discovery information about itself, in the form of labels, to other components.
* `dependencies` - Other components within the bundle which the current component depends on. Defining dependencies helps Aptomi process discovery information and propagate parameters
  in the right order, also allowing you to control the correct instantiation/destruction order of application components.

For example, here is how you would define an application which consists of Wordpress and MySQL database components:
```yaml
- kind: bundle
  metadata:
    namespace: main
    name: wordpress

  labels:
    blog: true

  components:
    - name: wordpress_component
      code:
        type: helm
        params:
          chartRepo: https://myhelmcharts.com/repo
          chartName: wordpress
          chartVersion: 4.9.1
          db_url: "{{ .Discovery.mysql_component.url }}"
          db_name: "wp"
      dependencies:
        - mysql_component

    - name: mysql_component
      code:
        type: helm
        params:
          chartRepo: https://myhelmcharts.com/repo
          chartName: mysql
          chartVersion: 5.7.21
      discovery:
        url: "mysql-{{ .Discovery.instance }}:3306"
```

For Helm plugin, you need to provide the following parameters under the `params` section in `code`, while the rest of the parameters will be passed "as is" to the instantiated Helm chart:
* `chartRepo` - The **URL** of the repository with your Helm charts
* `chartName` - The **name** of the Helm chart
* `chartVersion` *(Optional)* - The **version** of the Helm chart. If the chart version is not specified, the latest version will be used
* `target` - The name of the **cluster** and, optionally, **k8s namespace** to which the code will be deployed

Every parameter under the `params` section can be either a fixed value or an expression that refers to various labels.

Components can also have custom criteria defined and associated with them. If a specified criterion evaluates to true, the component is then included into a bundle. Otherwise, it will be excluded from processing. For example:
```yaml
- kind: bundle
  metadata:
    namespace: main
    name: wordpress

  components:
    - name: wordpress_component
      ...
    - name: mysql_component
      criteria:
        require-all:
          - db_enabled
      ...
```

## Service
Once a bundle is defined, it has to be exposed through a [service](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Service).

Services represent a consumable bundle. When someone wants to consume an instance of a bundle, it has to be done through a service. When
a bundle depends on another instance of a different bundle, it has to have a dependency on the corresponding service as one of its components.

A Service can have a number of contexts, which define different implementations of that service. Each context allows you to define the specific implementation and criteria under which it will be picked.
This means that a service will be fulfilled via one of the defined contexts at runtime, based on the defined criteria. Different parameters/bundles can be picked based on the type of environment (e.g. `dev` vs. `prod`), the properties of a consumer (e.g. which 'team' a consumer belongs to),
the time of day, or any other user-defined labels/properties. A service can also control whether the bundle instance will be dedicated or shared.

For example, a team responsible for running databases can define a service for `sql-database` and its specific implementations. In the code-snippet below, we see two contexts defined for the `sql-database` context:
* `dev` - For people from the 'dev' team, which will be implemented via instantiating the `sqlite` bundle.
* `prod` - For people outside of the 'dev' team, which will be implemented via instantiating the `mysql` bundle.
```yaml
- kind: service
  metadata:
    namespace: main
    name: sql-database

  contexts:
    - name: dev
      criteria:
        require-all:
          - team == 'dev'
      allocation:
        bundle: sqlite

    - name: prod
      criteria:
        require-none:
          - team == 'dev'
      allocation:
        bundle: mysql
```

In the example shown below, the team who runs the wordpress application wants to rely on an external database, and have modified their bundle definition to
use a special **service component**, which points to a service instead of the code:
```yaml
- kind: bundle
  metadata:
    namespace: main
    name: wordpress

  labels:
    blog: true

  components:
    - name: wordpress_component
      code:
        type: helm
        params:
          chartRepo: https://myhelmcharts.com/repo
          chartName: wordpress
          chartVersion: 4.9.1
          db_url: "{{ .Discovery.db_component.database.url }}"
          db_name: "wp"
      dependencies:
        - db_component

    - name: db_component
      service: sql-database
```

If you don't need the power of services and want your bundle to be instantiated directly, you can create a service which maps 1-to-1 to
the corresponding bundle, while keeping the criteria empty (i.e. always true). An example of this is shown below:
```yaml
- kind: service
  metadata:
    namespace: main
    name: mysql

  contexts:
    - name: primary
      allocation:
        bundle: mysql
```

In order to control service dedication/sharing, you can use a special `keys` field inside an allocation.
For example, the following YAML block specifies that each team will receive its own dedicated instance of the `mysql` bundle:
```yaml
- kind: service
  metadata:
    namespace: main
    name: mysql

  contexts:
    - name: primary
      allocation:
        bundle: mysql
        keys:
          - "{{ .User.Labels.Team }}"
```

When fulfilling a service, Aptomi will process all contexts within that service one-by-one, and find the first matching context. Once a context is selected, labels will be changed according to the `change-labels` section, and bundle allocation will be done according to the corresponding `allocation` section within the selected context.

## Cluster

A [Cluster](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Cluster) is an entity which defines a cluster in Aptomi where containers can be deployed. Even though Aptomi is focused on k8s, it is designed to support
multiple cluster types, such as Docker Swarm, Apache Mesos, and others. Cluster type is defined via the `type` attribute.

Clusters are global to Aptomi and must be always defined in the `system` namespace.

A typical definition of a k8s cluster looks like this:
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

## Claim

Defining a bundle and a service only publishes a service into Aptomi, and does not trigger instantiation/deployment of that service.
In order to request an instance, you must create a [Claim](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Claim) object in Aptomi.

When creating a claim, you must specify a `service` to instantiate, as well as the initial labels to pass to it.

For example, here is how **Alice** would request a **wordpress**:
```yaml
- kind: claim
  metadata:
    namespace: main
    name: alice_uses_wordpress
  user: Alice
  service: wordpress
  labels:
      label1: value1
```

Since Aptomi rules are all label-based, you can create a policy to make intelligent decisions based on the initial set of labels being passed, as well as transform those labels according to your needs.

## Rule

One of the most powerful features of Aptomi is the ability to define [rules](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Rule), which get evaluated at runtime during state enforcement.

Rules can be **global** (defined in `system` namespace) as well as **local** (defined within a given namespace).

The order in which the rules are processed and executed is as follows:
* local rules - one by one for a given namespace, sorted by weight
* global rules - one by one, sorted by weight

A rule can have user-defined criteria and associated actions. If the criterion evaluates to true, then an action is executed. The list of supported actions is:
* change-labels - change one or more labels
* claim - reject claim and not allow instantiation

The most commonly used rule action in Aptomi is to change a label. For example, by changing a system-level label called `target`, you can control which cluster and namespace the code will get deployed to. Deploying
code without setting the `target` label will result in an error, because Aptomi won't have a way of knowing where the code should be deployed.

For example, the following rule will tell Aptomi to always deploy bundles with the `blog` label to the cluster named `cluster-us-east`:
```yaml
- kind: rule
  metadata:
    namespace: main
    name: blog_bundles_get_deployed_to_us_east
  weight: 10
  criteria:
    require-all:
      - bundle.Labels.blog == true
  actions:
    change-labels:
      set:
        target: cluster-us-east
```

Here is another example of a rule, which prohibits users from the `dev` team from instantiating any `blog` bundles. Even if those users try to declare a claim, it will not be fulfilled by Aptomi:
```yaml
- kind: rule
  metadata:
    namespace: main
    name: dev_teams_cannot_instantiate_blog_bundles
  weight: 20
  criteria:
    require-all:
      - team == 'dev'
      - bundle.Labels.blog == true
  actions:
    claim: reject
```

# Common constructs
## Labels
Policy processing in Aptomi is based entirely on labels. When a claim is defined, an initial set of labels is formed by combining the labels of the requester (e.g. user labels) and a given claim. Throughout processing,
these labels can be transformed using `change-labels` directives, and accessed at any time in expressions and templates.

One way of leveraging this, for example, would be to:
* Define a bundle, a service, and two contexts.
* Set the `replicas` label to have a value of `1` in the dev context.
* Set the `replicas` label to have a value of `3` in the production context.
* Use the `replicas` label in the service to configure it appropriately.
* This way, Aptomi would be able to provision different instances of the same bundle with different settings.

The following code illustrates how one would use [change-labels](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#LabelOperations) to achieve the above objectives:
```yaml
- kind: service
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
        bundle: mybundle

    - name: prod
      criteria:
        require-none:
          - team == 'dev'
      change-labels:
        set:
          replicas: 3
      allocation:
        bundle: mybundle
```

## Expressions
All expressions used in Aptomi should follow the [Knetic/govaluate](https://github.com/Knetic/govaluate) syntax guidelines and must evaluate to a bool.

You can reference the following variables in expressions:
* labels - You can reference any label by specifying its name, e.g. `team` will return the value of a label with the name 'team'.
* bundles - You can reference a bundle which is currently being processed. Since it's an object, you can go down and look into its properties, e.g. `bundle.Name` or `bundle.Labels.blog`

## Criteria
[Criteria](https://godoc.org/github.com/Aptomi/aptomi/pkg/lang#Criteria) allow you to define complex matching expressions in your policy.
Criteria constructs in Aptomi support `require-all`, `require-any` and `require-none` sections, with a list of expressions under each section.

Criteria evaluate to true only when:
* All `require-all` expressions evaluate to true
* At least one of the `require-any` expressions evaluates to true
* None of the `require-none` expressions evaluate to true

If a section is absent, it will be skipped. So it's perfectly fine to have criteria with fewer than 3 sections (e.g. with just `require-all`) or with no sections at all.

It's also possible to have empty criteria without any clauses (or even omit the `criteria` construct all together). In this case, empty criteria are always considered to be 'true'.

## Templates
All text templates used in Aptomi should follow the [text/template](https://golang.org/pkg/text/template/) syntax guidelines, and must evaluate to a string.

The most common use of text templates in Aptomi is to define the code & discovery parameters inside a bundle:
* **code parameters** - Pass parameters into Helm charts, substituting variables with calculated label values
* **discovery parameters** - Components can expose their discovery parameters to other bundles/components, substituting variables with calculated label values

You can reference the following variables in your text templates:

* `{{ .Labels }}` - the current set of labels, e.g.:
  * `{{ .Labels.labelName }}` will return the value of label with name `labelName`
* `{{ .User}}` - the current user who defined a claim
  * `{{ .User.Name }}` - name of the user
  * `{{ .User.Secrets }}` - a map of user secrets
  * `{{ .User.Labels }}` - a map of user labels
* `{{ .Discovery }}` - a set of discovery parameters
  * `{{ .Discovery.instance }}` - a unique human-readable deployment name of the current component instance to be deployed
  * `{{ .Discovery.instanceid }}` - a unique hash of the current component instance to be deployed
  * `{{ .Discovery.Bundle.instanceid }}` - a unique hash of the current bundle instance to be deployed
  * `{{ .Discovery.component1.[...].componentN.propertyName }}` - you can traverse component graph to get the value of 'propertyName' from discovery properties exposed by an particular component

## Namespace references
Sometimes you will want to specify an absolute path to an object located in a different namespace.

For example, the following code would tells Aptomi to look for a wordpress service defined in the **same** namespace (in this case, `main`):
```yaml
- kind: claim
  metadata:
    namespace: main
    name: alice_uses_wordpress
  user: Alice
  service: wordpress
```

While this code would tell Aptomi to look for a wordpress service in a **given** namespace (in this case, `specialns`):
```yaml
- kind: claim
  metadata:
    namespace: main
    name: alice_uses_wordpress
  user: Alice
  service: specialns/wordpress
```

You can also use the same syntax to create a bundle, which depends on a service published in a different namespace (i.e. `dbns`) by another team:
```yaml
- kind: bundle
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
      service: dbns/sql-database
```
