Implementation:

1. Support for namespaces [ARCH DISCUSSION NEEDED]
  - we need to make sure everything that calls getName() is within a namespace
  - add checks for duplicate names in the same NS
  - add references to support "namespace/"
  - inheritance / make sure we are looking for contexts in the right namespace
  - move istio into a 'system' namespace
  - rules in namespaces? global rules in system namespace?
  - rules to reference namespaces?

2. ACL [ARCH DISCUSSION NEEDED]

3. Istio [ARCH DISCUSSION NEEDED]
  - Figure out a good model to fit services like istio into the engine
  - Without having user to create contexts for them

4. Implement LDAP sync as an external service
  - Generic system for aggregating and storing labels from different data sources
  - Log: gets tied to a separate revision of users, and the service keeps N last revisions)

5. Implement policy validation
  - e.g. compile all expressions, templates, etc

6. Attach policy apply log to component instances

7. Labels for services (to use in expressions)

8. Versions for services
   - Version is a special label, which can be compared

9. Aptomi quickstart
   - with sample app

10. Illustrate prod vs. stage contexts better in the demo (# of replicas, etc)

11. Cluster should be in component key


Minor issues:
- Get rid of dependency ID
- Deal with code style and missing comments
- Shall we consider renaming .User -> .Consumer?
- Plugins should support noop mode (if at all possible). I.e. noop should log Helm commands, but don't run them
- Unit tests are 50% using "testdata" and 50% using hand-created objects. Might make sense to use the latter everywhere
- Deal with EscapeName (it's Helm plugin specific, should not be present in engine)


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

* Wrap all external data into a single struct
  - decouple secrets from users/labels
  - load secrets normally, not in LDAPUserLoader

* Move Dependency.Resolved to policy resolution data

* Get rid of Resolved flag in component

* Get rid of Resolved/Unresolved in policy resolution, given that we changed the way how logs are stored

* Optimize policy resolution

* Multi-threaded policy resolution

* Implemented contracts
  * contexts are now local to contracts
  * contexts are evaluated in the order specified in the contract
  * removed change-labels from services and components

* Implemented rules
  * Ability to set cluster via rules vs in context
  * Reject everything by default
  * Have strict evaluation order
    - ordered by weight
  * Types of rules to support
    - expression (user/service/etc labels) => allow dependency
    - expression (user/service/etc labels) => reject dependency
    - expression (user/service/etc labels) => set labels (e.g. cluster)
  * Ability to do blacklist and whitelist rules
    - stop right away if encountered reject
  * in() function
