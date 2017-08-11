What is going on:
* introducing universal objects
* dealing with context inheritance (== contexts not tied to services)

Items to complete:
1. Support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references must include "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace

2. Context inheritance (when we go into dependent services, service.Name won't work; need to figure out label-based thing)
     - e.g. list of services to which the context applies to
     - or do "and" condition (e.g. org == dev and services in [...])

3. Add support not only for policy objects, but for generated objects as well

4. Implement a file db loader/writer
  - get rid of revision.yaml across the code and move into that implementation

5. Get rid of dependency ID

6. Deal with external entities (charts, secrets, ldap configuration)

7. Add unit tests to verify label calculation in the engine while policy is being resolved

8. Move istio into a 'system' namespace

9. Check security issue with Knetic (possible to call methods on objects from the policy)
