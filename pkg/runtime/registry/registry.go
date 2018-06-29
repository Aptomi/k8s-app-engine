package registry

import (
	"github.com/Aptomi/aptomi/pkg/engine"
	"github.com/Aptomi/aptomi/pkg/engine/actual"
	"github.com/Aptomi/aptomi/pkg/engine/apply/action"
	"github.com/Aptomi/aptomi/pkg/engine/resolve"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
)

// Interface represents main object registry interface that covers database operations for all objects
type Interface interface {
	PolicyRegistry
	RevisionRegistry
	ActualStateRegistry
}

// PolicyRegistry represents database operations for Policy object
type PolicyRegistry interface {
	GetPolicy(runtime.Generation) (*lang.Policy, runtime.Generation, error)
	GetPolicyData(runtime.Generation) (*engine.PolicyData, error)
	InitPolicy() error
	UpdatePolicy(updated []lang.Base, performedBy string) (changed bool, data *engine.PolicyData, err error)
	DeleteFromPolicy(deleted []lang.Base, performedBy string) (changed bool, data *engine.PolicyData, err error)
}

// RevisionRegistry represents database operations for Revision object
type RevisionRegistry interface {
	NewRevision(policyGen runtime.Generation, desiredState *resolve.PolicyResolution, recalculateAll bool) (*engine.Revision, error)
	GetDesiredState(*engine.Revision) (*resolve.PolicyResolution, error)
	GetRevision(gen runtime.Generation) (*engine.Revision, error)
	UpdateRevision(revision *engine.Revision) error
	NewRevisionResultUpdater(revision *engine.Revision) action.ApplyResultUpdater
	GetFirstUnprocessedRevision() (*engine.Revision, error)
	GetLastRevisionForPolicy(policyGen runtime.Generation) (*engine.Revision, error)
	GetAllRevisionsForPolicy(policyGen runtime.Generation) ([]*engine.Revision, error)
}

// ActualStateRegistry represents database operations for the actual state handling
type ActualStateRegistry interface {
	GetActualState() (*resolve.PolicyResolution, error)
	NewActualStateUpdater(*resolve.PolicyResolution) actual.StateUpdater
}
