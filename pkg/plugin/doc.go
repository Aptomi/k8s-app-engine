// Package plugin introduces plugins for Aptomi engine such as cluster (kubernetes) and code (helm) plugins.
//
// Cluster plugins responsible for working with different cloud providers, such as kubernetes. Separated instance of
// the cluster plugins is created for each lang.Cluster, so, it could safely cache data.
//
// Code plugins responsible for handling deployment into the cloud providers, such as helm. Separated instance of the
// code plugin is created for each pair of the lang.Cluster and code type, for example, cluster "test-1" and code
// type "helm".
//
// All plugins created for single enforcement cycle or API call using plugin registry.
package plugin
