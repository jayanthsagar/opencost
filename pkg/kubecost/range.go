package kubecost

import (
	"fmt"
	"github.com/kubecost/cost-model/pkg/util/json"
	"sync"
)

// Range is a generic implementation
type Range[T Set] struct {
	lock sync.RWMutex
	sets []T
}

func (r *Range[T]) Append(that T) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.sets = append(r.sets, that)
}

// Each invokes the given function for each Set in the Range
func (r *Range[T]) Each(f func(int, T)) {
	if r == nil {
		return
	}

	for i, set := range r.sets {
		f(i, set)
	}
}

func (r *Range[T]) Get(i int) (T, error) {
	if i < 0 || i >= len(r.sets) {
		var set T
		return set, fmt.Errorf("range: index out of range: %d", i)
	}

	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.sets[i], nil
}

func (r *Range[T]) Length() int {
	if r == nil || r.sets == nil {
		return 0
	}

	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.sets)
}

// IsEmpty returns false if Range contains a single Set that is not empty
func (r *Range[T]) IsEmpty() bool {
	if r == nil || r.Length() == 0 {
		return true
	}
	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, set := range r.sets {
		if !set.IsEmpty() {
			return false
		}
	}
	return true
}

func (r *Range[T]) MarshalJSON() ([]byte, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return json.Marshal(r.sets)
}
