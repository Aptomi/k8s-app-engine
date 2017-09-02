Items to complete:

1. Support for namespaces
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references to support "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace
  - move istio into a 'system' namespace
  - rules in namespaces? global rules in system namespace?
  - rules to reference namespaces?

2. Implement LDAP sync as an external service
  - Generic system for aggregating and storing labels from different data sources
  - Log: gets tied to a separate revision of users, and the service keeps N last revisions)

3. Figure out a good model to fit services like istio into the engine
  - without having user to create contexts for them

4. Contexts (and possibly other objects, such as rules) are evaluated in random order
  - If multiple contexts match a dependency, then the behavior will not be deterministic
  - Policy evaluation, when ran multiple times, can result in different outcomes

5. I, as an operator, can accidentally add a context (or rule), which can easily
   break all services or move them to another cluster, etc
  - Need to figure out how to prevent this

7. Implement policy validation
  - e.g. compile all expressions, templates, etc

8. RevisionSummary (moving away from files to boltdb)
  - Implement RevisionSummary object (wraps Revision and stores summary)
      - Implement diff inside RevisionSummary
          - structured (for all objects)
          - summary (as text/numbers)
      - Implement bool Changed() inside RevisionSummary

9. Attach apply log to component instances


Minor issues:
- Where/how to store text-based diff for revisions?
- Get rid of dependency ID
- Do we need to move Dependency.Resolved into resolved usage state?
- Deal with code style and missing comments
- Figure out a better way to deal with secrets in LabelSets. Check again how they are behing printed into event logs
- Shall we consider renaming .User -> .Consumer?
- Plugins should support noop mode (if at all possible). I.e. noop should log Helm commands, but don't run them
- Unit tests are 50% using "testdata" and 50% using hand-created objects. Might make sense to use the latter everywhere


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

* Improve debugging of the policy (for successful & unsuccessful processing)

  - Which event logs can we store
      -> policy evaluation log
           * successful == gets tied to a revision
           * error = stays in the journal orphan (not tied to a revision)
      -> policy apply log
           * gets tied to a revision

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

  - Error handling in the engine

* Separate apply() from diff() and resolve() in the engine
  - Outstanding big chunk of work

* Introduce engine plugins

* Detect cross-service cycle in engine to prevent infinite loops

* Improved engine unit test coverage for corner cases

* Modify code around charts to save them into tmp files

* Rename actions/plugins to desired & actual

* Actions must update actual state

* Deal with component create/update times (calculated in diff)
  - This logic should likely be moved to state reconciliation (updating component update/create times in Apply)

