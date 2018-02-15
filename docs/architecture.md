# Aptomi Architecture

## Components
Aptomi consists of two main binaries:
* Aptomi Server - `aptomi` is a server, which serves API & UI as well as does deployment and continuous state enforcement of applications/components running on Kubernetes
* Aptomi Client - `aptomictl` is a client, which talks to Aptomi Server over API. It allows end-users to interact with Aptomi 

![Aptomi Components](../images/aptomi-components.png) 

Aptomi server has the following main internal components:
* **UI and API** - served over HTTP
* **Policy Engine** - engine to process the uploaded "policy" (app definitions, cluster definitions, rules) and translate it into a `Desired State`
* **State Enforcer** - applies `Desired State`, creating/updating/deleting containers in Kubernetes and applying configs/rules
* **Database** - uses [Bolt](https://github.com/boltdb/bolt) as a database to persist its data

## State Enforcement
Aptomi has a notion of `Desired State` and `Actual State`:
* `Desired State` basically defines which components should be running, where, how they should be configured, and how the insfrastructure components around them should be set up
* `Actual State` is a result of applying changes. It may not be equal to the desired state (e.g. when Aptomi tried to create/update/delete a container in Kubernetes, but was unable to)

![Aptomi Enforcement](../images/aptomi-enforcement.png)

Aptomi is continuously validating/enforcing the state, always trying to reconcile `Desired State` and `Actual State` 

![Aptomi Engine Architecture](../images/aptomi-engine-architecture.png)

