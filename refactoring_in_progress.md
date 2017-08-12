What is going on:
* introducing universal objects
* dealing with context inheritance (== contexts not tied to services)

Items to complete:
1. Support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references must include "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace

3. Add support not only for policy objects, but for generated objects as well

4. Implement a file db loader/writer
  - get rid of revision.yaml across the code and move into that implementation

5. Get rid of dependency ID

6. Deal with external entities (charts, secrets, ldap configuration)

7. Add unit tests to verify label calculation in the engine while policy is being resolved

8. Move istio into a 'system' namespace

10. Figure out what to do with logging... it's messed up right now
  - RIGHT NOW IT'S VERY HARD TO DEBUG POLICY / UNDERSTAND WHAT'S GOING ON
  - do we want to show users a full log for policy evaluation?
  - only show a particular namespace?
  - e.g. what if criteria expression failed to compile, or evaluation fails (we are comparing integer to a string), how do we propagate this to the user?

11. Expose contextual data to templates as well
  - same way, as we are doing for expressions

12. We likely need to separate labels from label operations
  - so that services can have labels (and we can refer to them from criteria expressions)
  - and services can have labels ops (set/remove, etc)
  - just rename existing "labels" (ops) to change-labels in yaml


Done:

2. Context inheritance

9. Check security issues with Knetic (possible to call methods on objects from the policy expressions)
