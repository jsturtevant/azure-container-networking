package policies

import "sync"

type PolicyMap struct {
	sync.Mutex
	cache map[string]*NPMNetworkPolicy
}

type PolicyManager struct {
	policyMap *PolicyMap
}

func NewPolicyManager() PolicyManager {
	return PolicyManager{
		policyMap: &PolicyMap{
			cache: make(map[string]*NPMNetworkPolicy),
		},
	}
}

func (pMgr *PolicyManager) GetPolicy(name string) (*NPMNetworkPolicy, error) {
	pMgr.policyMap.Lock()
	defer pMgr.policyMap.Unlock()

	if policy, ok := pMgr.policyMap.cache[name]; ok {
		return policy, nil
	}

	return nil, nil
}
