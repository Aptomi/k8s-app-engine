Items to complete:

1. Support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references must include "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace
  - move istio into a 'system' namespace
  - rules in namespaces? global rules in system namespace?
  - rules to reference namespaces?

2. Figure out what to do with logging... it's messed up right now
  - RIGHT NOW IT'S VERY HARD TO DEBUG POLICY / UNDERSTAND WHAT'S GOING ON
  - do we want to show users a full log for policy evaluation?
  - only show a particular namespace?
  - e.g. what if criteria expression failed to compile, or evaluation fails (we are comparing integer to a string), how do we propagate this to the user?
  - debug log vs. rule log
  - get rid of all debug in language
  - use logging only in engine
  - implement event log (filterable by ns, obj type, etc)

3. Error handing in engine & improved engine unit test coverage for corner cases

4. Get rid of dependency ID

5. Do we need to move Dependency.Resolved into resolved usage state?

6. Get rid of service.Metadata.Name == 'istio'

7. Reformat, deal with code style and missing comments

8. Implement polling for external entities & storing objects in DB


Questions:
1. Shall we consider renaming .User -> .Consumer?


Done:
* Flexible contexts (==inheritance, ==more powerful expressions)

* Check security issues with Knetic (possible to call methods on objects from the policy expressions)

* Change criteria to <RequireAll>, <RequireAny> and <RequireNone>

* Separate labels from label operations
  - so that services can have labels (and we can refer to them from criteria expressions)
  - and services can have labels operations (set/remove, etc)
  - just rename existing "labels" (ops) to "change-labels" in yaml to avoid confusion

* Add unit test to verify label calculation in the engine while policy is being resolved

* Expose contextual data to templates as well
  - same way, as we are doing for expressions

* Add caching for expressions and templates

* Add unit tests for allocation name resolution with variables (referring with .User and .Labels)

* Removed change-labels for allocation

* Remove custom JSON serializers

* Component keys
  - permanent key: service, context, allocation, component
  - additional keys defined by the context:
        e.g. "stage" allocation can define additional key
        and say that "User.ID" gets inserted into allocation name
  - get rid of allocationNameResolved -> allocation key resolution
  - fix graphviz, strings.Split(key, "#")
  - get rid of ComponentRootName everywhere

* Remove allocation name all together and leave only keys (engine + UI)

* Separate progress calculation code from progress bar

* Make target for unit test coverage
  - Go tooling for code coverage works for individual packages
  - So if we have 'language' pkg at 50% and the other half is covered by tests from 'engine', it will be impossible to calculate
  - We need to have packages completely independent. With their own independent code coverage. Can't rely on cross-package tests
  - Also see https://github.com/pierrre/gotestcover
