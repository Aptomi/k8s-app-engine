Items to complete:

3. Support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references must include "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace
  - move istio into a 'system' namespace

4. Add support not only for policy objects, but for generated objects as well

5. Implement a file db loader/writer
  - get rid of revision.yaml across the code and move into that implementation

6. Get rid of dependency ID

7. Deal with external entities (charts, secrets, ldap configuration)

9. Figure out what to do with logging... it's messed up right now
  - RIGHT NOW IT'S VERY HARD TO DEBUG POLICY / UNDERSTAND WHAT'S GOING ON
  - do we want to show users a full log for policy evaluation?
  - only show a particular namespace?
  - e.g. what if criteria expression failed to compile, or evaluation fails (we are comparing integer to a string), how do we propagate this to the user?

11. Remove custom JSON serializers

12. Code coverage

Done:
* Flexible contexts (==inheritance, ==more powerful expressions)

* Check security issues with Knetic (possible to call methods on objects from the policy expressions)

* Change criteria to <RequireAll>, <RequireAny> and <RequireNone>

* Separate labels from label operations
- so that services can have labels (and we can refer to them from criteria expressions)
- and services can have labels operations (set/remove, etc)
- just rename existing "labels" (ops) to "change-labels" in yaml to avoid confusion

* add unit test to verify label calculation in the engine while policy is being resolved

* expose contextual data to templates as well
  - same way, as we are doing for expressions

* add caching for expressions and templates

* add unit tests for allocation name resolution with variables (referring with .User and .Labels)

* removed change-labels for allocation