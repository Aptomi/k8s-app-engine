# Slinga - Service Control and Definition Language

## Entities

* **Service** - object which defines a service (e.g. kafka). Each service has
  * `name` - service name
  * `labels` - set of labels assigned to this service
  * `components` - set of components the service consists of. Each component has
    * `name` - component name
    * it can be either `code` (e.g. executable template for `helm`) or `service` (dependency on instance of another service)
    * `labels` - how labels get changed before passing further to the dependencies
    * `dependencies` - what other components this component depends on (within a given service)

* **Context** - defines a service context & specific rules how to allocate service instances. Each context has
  * `name` - context name
  * `criteria` - list of expression (if as least one evaluates to true, then the context would be matched)
  * `labels` - set of labels assigned to this service
  * `components` - set of components the service consists of. Each component has
    * `name` - component name
    * it can be either `code` (e.g. executable template for `helm`) or `service` (dependency on instance of another service)
    * `labels` - how labels get changed before passing further to the dependencies
    * `dependencies` - what other components this component depends on (within a given service)

## Diagram

To be added

## Examples

To be added