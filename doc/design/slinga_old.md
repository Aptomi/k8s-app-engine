# Slinga - Service Control and Definition Language

Entities
* org
* scope
* service
* context
* component
* label
*

Questions:
* how one service will get some values from another? Like how Kafka will understand where to take ZooKeeper address?
  IMHO it should be done in the following way: ZooKeeper sets some labels while deploying and using that exposing all potentially needed information and then Kafka will use this labels


Entities / Terms:

* **Type** (*Object Type*) - type of the `Object` in system, e.g. `Service`,
  `Contract` and etc.
* **Object** - instance of any `Type` in system
* **Name** - DNS subdomain like string that defines name for the concrete
  instance of `Object`, unique inside the `Org`
* **Org** (*Organization*) - DNS subdomain like string that defines
  namespace any `Object` in system belongs to
  * `<some.subdomain.like.name>`
  * e.g. `super_company.platform_team.dev` or just `aptomi`
* **Path** - unique id (or address) of any `Object` in system, it consists of
  `Org`, `Type` and `Name` separated by slashes `/`, where `Type` and `Name`
  could be repeated to identify nested `Object`
  * `<some.subdomain.like.name>/<object_type>/<some.subdomain.like.object.name>[/<subtype>/<some.subname>]`
  * e.g. Service: `company_a.team_1.platform/service/kafka`
  * e.g. Contract Scope: `company_a.team_1.platform/contract/kafka/scope/scary-people`
* **Label** - key-value pair

Object Types:
* Service
* Contract

```yaml
---
company_a.team_1.platform/service/kafka:
  labels:
    app: analytics
    display_name: "Apache Kafka"
    description: "Super Fancy Kafka Service from Platform team"
    # for example: service/type: "analytics-iot"
  components:
    kafka:
      code:
        aptomi/orch/kubernetes-helm:
          - name: kafka
            version: 1.13.2
            repository: http://storage.googleapis.com/kubernetes-charts-incubator
            values:
              helm-value: some-expression-based-on-label
      dependencies:
        - company_a.team_1.platform/contract/zookeeper: context_name

---
company_a.team_1.platform/contract/kafka:
  services:
    - company_a.team_1.platform/service/kafka
  contexts:
    mock:
      services:
        - company_a.team_1.platform.qa/service/kafka-mock:
            a: b
      ownership: shared-with-all
    test:
      ownership : shared-within-all
    staging:
      ownership: shared-within-scope
      labels:
        a: b
    prod:
      ownership: shared-within-scope
    prod-high-prio:
      ownership: dedicated-per-consumer
  scopes:
    scary-people:
      contexts:
        - mock
        - test
      filters:
        - *.dev:
          - {stage/in-test && ...}
          - {stage/in-test2 && ...}
    crazy-people-with-guns:
      contexts:
        - staging
        - prod
      filters:
        - *.prod

---
company_a.team_1.platform/service/zookeeper:
  labels:
    aptomi/display_name: "Apache ZooKeeper"
    aptomi/description: "Super Fancy ZooKeeper Service from Platform team"
  components:
    zookeeper:
      labels: []
      code:
        aptomi/orch/kubernetes-helm:
          - name: zookeeper
            version: 1.42.5
            repository: http://storage.googleapis.com/kubernetes-charts-incubator
            values:
              a: b

---
company_a.team_1.platform/contract/zookeeper: {}

```
