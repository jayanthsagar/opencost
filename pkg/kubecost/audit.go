package kubecost

import (
	"encoding"
	"golang.org/x/exp/slices"
	"sync"
	"time"
)

type AuditType string

const (
	AuditAllocationReconciliation AuditType = "AuditAllocationReconciliation"
	AuditAll                      AuditType = ""
	AuditInvalidType              AuditType = "InvalidType"
)

// ToAuditType converts a string to an Audit type
func ToAuditType(check string) AuditType {
	switch check {
	case string(AuditAllocationReconciliation):
		return AuditAllocationReconciliation
	case string(AuditAll):
		return AuditAll
	default:
		return AuditInvalidType
	}
}

type AuditStatus string

const (
	FailedStatus AuditStatus = "Failed"
	PassedStatus             = "Passed"
)

type AuditError error

// Audit contains the result of a process run by an Auditor
type Audit struct {
	AuditType   AuditType
	Status      AuditStatus
	Description string
	LastRun     time.Time
}

func (a *Audit) Clone() *Audit {
	return &Audit{
		AuditType:   a.AuditType,
		Status:      a.Status,
		Description: a.Description,
		LastRun:     a.LastRun,
	}
}

type AuditFloatResult struct {
	Expected float64
	Actual   float64
}

func (afr *AuditFloatResult) Clone() *AuditFloatResult {
	return &AuditFloatResult{
		Expected: afr.Expected,
		Actual:   afr.Actual,
	}
}

type AllocationReconciliationAudit struct {
	Status            AuditStatus
	Description       string
	LastRun           time.Time
	Resources         map[string]map[string]*AuditFloatResult
	AllocsBlankNode   []string
	NodesWithNoAllocs []string
	MissingNodes      []string
}

func (ara *AllocationReconciliationAudit) Clone() *AllocationReconciliationAudit {
	resources := make(map[string]map[string]*AuditFloatResult, len(ara.Resources))
	for node, resourceMap := range ara.Resources {
		copyResourceMap := make(map[string]*AuditFloatResult, len(resourceMap))
		for resourceName, val := range resourceMap {
			copyResourceMap[resourceName] = val.Clone()
		}
		resources[node] = copyResourceMap
	}
	return &AllocationReconciliationAudit{
		Status:            ara.Status,
		Description:       ara.Description,
		LastRun:           ara.LastRun,
		Resources:         resources,
		AllocsBlankNode:   slices.Clone(ara.AllocsBlankNode),
		NodesWithNoAllocs: slices.Clone(ara.NodesWithNoAllocs),
		MissingNodes:      slices.Clone(ara.MissingNodes),
	}
}

type AuditDetail interface {
	Clone() AuditDetail

	// Representations
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// AuditCoverage tracks coverage of each audit type
type AuditCoverage struct {
	sync.RWMutex
	AllocationReconciliation Window
}

func NewAuditCoverage() *AuditCoverage {
	return &AuditCoverage{}
}

// Update expands the coverage of each Window in the coverage that the given AuditSet's Window if the corresponding Audit is not nil
// Note: This means of determining coverage can lead to holes in the given window
func (ac *AuditCoverage) Update(as *AuditSet) {
	if as != nil && as.AllocationReconciliation != nil {
		ac.AllocationReconciliation.Expand(as.Window)
	}

}

type AuditSet struct {
	sync.RWMutex
	AllocationReconciliation *AllocationReconciliationAudit
	Window                   Window
}

// NewAuditSet creates an empty AuditSet with the given window
func NewAuditSet(start, end time.Time) *AuditSet {
	return &AuditSet{
		Window: NewWindow(&start, &end),
	}
}

// UpdateAuditSet overwrites any audit fields in the caller with those in the given AuditSet which are not nil
func (as *AuditSet) UpdateAuditSet(that *AuditSet) *AuditSet {
	if as == nil {
		return that
	}

	if that.AllocationReconciliation != nil {
		as.AllocationReconciliation = that.AllocationReconciliation
	}

	return as
}

func (as *AuditSet) IsEmpty() bool {
	return as == nil || (as.AllocationReconciliation == nil)
}

func (as *AuditSet) GetWindow() Window {
	return as.Window
}

func (as *AuditSet) Clone() *AuditSet {
	if as == nil {
		return nil
	}

	as.RLock()
	defer as.RUnlock()

	return &AuditSet{
		AllocationReconciliation: as.AllocationReconciliation.Clone(),
		Window:                   as.Window.Clone(),
	}
}

func (as *AuditSet) CloneSet() Set {
	return as.Clone()
}

type AuditSetRange struct {
	Range[*AuditSet]
}
