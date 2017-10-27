// Package apply implements state enforcer, which executes a list of actions to move from actual state to desired
// state, performing actual deployment of services and configuration of the underlying cloud components. As state
// gets enforced, it will configure the cloud to run new services/components, update existing services/components
// and delete services/components which are no longer needed.
package apply
