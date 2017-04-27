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

## Algorithm

We treat services and contexts as policy definition files.

Actions happen when:
  * Consumer asks for a service. Then we resolve everything and provide service instance to a consumer (resolving and instantiating all dependencies if/as needed)
  * Team makes a change to policy files. Or consumer labels get changed. We would need to re-evaluate and resolve everything once again, because e.g.
    * Consumer may lose access to an allocated service
    * Consumer may get directed to a different service instance

High-level description of how we process everything in the proof of concept:

1. When a consumer requests a service
  1. Add a placeholder record into a global "who is using what" list, in the form of
    * [`consumer X` is using `service Y`] -> (`empty`)
       * `empty` means that a specific instance has not been allocated yet for `consumer X`
  2. Call our algorithm to enforce the policy. As a result of policy enforcement, all empty objects will be resolved into specific instances

2. When a change to policy files is made
  1. Call our algorithm to enforce the policy. As a result of policy enforcement, some service allocations entries will change and we will need to deal with this

3. When a change to consumer attributes is made
  1. We have no way of knowing when external attributes in LDAP get changed. So we may want to call policy/state enforcement on a regular basis (every 1 minute)

4. How state enforcement is implemented
  1. Load all services and contexts
  2. Process all entries "who is using what" list, one by one
  3. For a given entry
    1. Take a service
    2. Topologically sort all of its components
    3. Resolve every component entry to either code or service instance
    4. For services
        1. Match context by criteria
        1. Match allocation by criteria

... to be continued

... we also need to store what has been allocated (allocation name is a unique key)
