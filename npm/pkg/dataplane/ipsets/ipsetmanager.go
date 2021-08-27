package ipsets

import (
	"sync"

	"github.com/Azure/azure-container-networking/npm/util/errors"
)

type IPSetMap struct {
	cache map[string]*IPSet
	sync.Mutex
}
type IPSetManager struct {
	listMap *IPSetMap
	setMap  *IPSetMap
}

func newIPSetMap() *IPSetMap {
	return &IPSetMap{
		cache: make(map[string]*IPSet),
	}
}

func (m *IPSetMap) exists(name string) bool {
	m.Lock()
	defer m.Unlock()
	_, ok := m.cache[name]
	return ok
}

func NewIPSetManager(os string) IPSetManager {
	return IPSetManager{
		listMap: newIPSetMap(),
		setMap:  newIPSetMap(),
	}
}

func (iMgr *IPSetManager) getSetCache(set *IPSet) (*IPSetMap, error) {
	kind := getSetKind(set)

	var m *IPSetMap
	switch kind {
	case ListSet:
		m = iMgr.listMap
	case HashSet:
		m = iMgr.setMap
	default:
		return nil, errors.Errorf(errors.CreateIPSet, false, "unknown Set kind")
	}
	return m, nil
}
