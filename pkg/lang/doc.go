// Package lang provides core constructs for describing Aptomi policy, as well as core structures for processing policy.
//
// Let's start with policy objects.
// Cluster - individual cluster where containers get deployed (e.g. k8s cluster).
// Service - service for a bundle (e.g. database).
// Context - a set of contexts, defining how service can be fulfilled (e.g. MariaDB, MySQL, SQLite).
// Bundle - specific bundle implementation (set of containers to run, and dependencies on other services).
// Claim - triggers instantiation of a service.
// Rule - rules which constitute policy, allowing to change labels and perform actions during policy resolution.
// ACLRule - rules which define user roles for accessing Aptomi namespaces.
//
// Now, core structures:
// LabelSet - set of labels that get processed and transformed
// LabelOperations - how to transform labels
// Criteria - bool-based lists of expressions for defining complex criteria
package lang
