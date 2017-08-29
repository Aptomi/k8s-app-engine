Items to complete:

1. Support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references must include "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace
  - move istio into a 'system' namespace
  - rules in namespaces? global rules in system namespace?
  - rules to reference namespaces?

2. Improve debugging of the policy (for successful & unsuccessful processing)

  - Which event logs can we store
      -> policy evaluation log
           * successful == gets tied to a revision
           * error = stays in the journal orphan (not tied to a revision)
      -> policy apply log
           * gets tied to a revision
      -> external services, e.g. LDAP sync service (gets tied to a separate revision of users, and the service keeps N last revisions)

  - Journal ID
    - event log gets written to a generic event log store and gets tied to a "Journal ID"
    - for successful policy execution, a revision will be associated with "Journal ID"
    - if error happened during policy execution, no revision will be associated with "Journal ID" (but you can still pull log and look at what happened)

  - Index/Filter
    - namespace (e.g. show only logs for namespaces user has access to based on ACL)
    - dependency
    - component instance key

  - Params
    - map of freeform params (e.g. for policy evaluation - things we show on UI)

  - Misc
    - rule log must be replaced with event log, filtered by componentKey
    - all Debug.* calls must be removed. Likely, create a logger in engine
    - all panic() calls must be removed. E.g. getCluster
    - all engine events should be propagated to a user
      - notices: e.g. context match
      - errors: if criteria expression failed to compile, or evaluation fails (we are comparing integer to a string)
    - if possible, attach event by context (not by passing a bunch of variables)

3. Error handing in engine & improved engine unit test coverage for corner cases

4. Separate apply() from diff calculation
   - Deal with component create/update times (calculated in diff)

5. Plugins: engine -> ComponentInstancePlugin() -> istio

6. Get rid of dependency ID

7. Do we need to move Dependency.Resolved into resolved usage state?

8. Get rid of service.Metadata.Name == 'istio'

9. Reformat, deal with code style and missing comments

10. Implement polling for external entities & storing objects in DB

11. Cross-service cycle in engine

12. Secrets being printed to event log left and right. Who should be able to see them, if anyone?


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
