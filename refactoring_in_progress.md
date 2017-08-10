What is going on:
* introducing universal objects
* dealing with context inheritance (== contexts not tied to services)

Items to complete:
- get rid of all previous loading code
- we need to deal with context inheritance (when we go into dependent services, service.Name won't work; need to figure out label-based thing)
  - e.g. list of services to which the context applies to
  - or do "and" condition (e.g. org == dev and services in [...])
- inheritance / make sure we are looking for contexts in the right namespace
- add support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references must include "namespace/"
- add support not only for policy objects, but for generated objects as well
- get rid of dependency ID
- get rid of revision.yaml
- deal with external entities (charts, secrets, ldap configuration)
- we don't have any unit tests to verify label calculation in the engine while policy is being resolved
- move istio into a 'system' namespace
