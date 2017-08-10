What is going on:
* introducing universal objects
* dealing with context inheritance (== contexts not tied to services)

Items to complete:
- get rid of all previous loading code
- rename "unittests_new" back to "unittests"
- remove enabled flag
- inheritance / make sure we are looking for contexts in the right namespace
- add support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
- add support not only for policy objects, but for generated objects as well
- get rid of dependency ID
- get rid of revision.yaml
- deal with external entities (charts, secrets, ldap configuration)
- we don't have any unit tests to verify label calculation in the engine while policy is being resolved
