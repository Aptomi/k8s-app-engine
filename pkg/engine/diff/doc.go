// Package diff allows Aptomi to determine the difference between actual state (running on the cloud) and desired
// state (that Aptomi wants to enforce), generating a list of actions to reconcile the difference. The generated
// list of actions will be passed to applier for execution.
package diff
