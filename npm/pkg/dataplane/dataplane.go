package dataplane

import (
	"runtime"

	"github.com/Azure/azure-container-networking/npm"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ipsets"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/policies"
)

type DataPlane struct {
	policyMgr policies.PolicyManager
	ipsetMgr  ipsets.IPSetManager
	OsType    OsType
	networkID string
	// key is PodKey
	endpointCache map[string]*NPMEndpoint
}

type NPMEndpoint struct {
	Name      string
	ID        string
	NetpolRef []string
}

func NewDataPlane() *DataPlane {
	return &DataPlane{
		OsType:        detectOsType(),
		policyMgr:     policies.NewPolicyManager(),
		ipsetMgr:      ipsets.NewIPSetManager(string(detectOsType())),
		endpointCache: make(map[string]*NPMEndpoint),
	}
}

type OsType string

const (
	Windows OsType = "windows"
	Linux   OsType = "linux"
)

// Detects the OS type
func detectOsType() OsType {
	os := runtime.GOOS
	switch os {
	case "linux":
		return Linux
	case "windows":
		return Windows
	default:
		panic("Unsupported OS type: " + os)
	}
}

func (dp *DataPlane) InitializeDataPlane() error {
	return dp.initializeDataPlane()
}

func (dp *DataPlane) ResetDataPlane() error {
	return dp.resetDataPlane()
}

func (dp *DataPlane) CreateIPSet(set *ipsets.IPSet) error {
	return dp.ipsetMgr.CreateIPSet(set)
}

func (dp *DataPlane) DeleteSet(name string) error {
	return dp.ipsetMgr.DeleteSet(name)
}

func (dp *DataPlane) DeleteList(name string) error {
	return dp.ipsetMgr.DeleteList(name)
}

func (dp *DataPlane) AddToSet(setNames []*ipsets.IPSet, ip, podKey string) error {
	return dp.ipsetMgr.AddToSet(setNames, ip, podKey)
}

func (dp *DataPlane) RemoveFromSet(setNames []string, ip, podKey string) error {
	return dp.ipsetMgr.RemoveFromSet(setNames, ip, podKey)
}

func (dp *DataPlane) AddToList(listName string, setNames []string) error {
	return dp.ipsetMgr.AddToList(listName, setNames)
}

func (dp *DataPlane) RemoveFromList(listName string, setNames []string) error {
	return dp.ipsetMgr.RemoveFromList(listName, setNames)
}

func (dp *DataPlane) UpdatePod(pod *npm.NpmPod) error {
	return nil
}

func (dp *DataPlane) ApplyDataPlane() error {
	return nil
}

func (dp *DataPlane) AddPolicies(policies *policies.NPMNetworkPolicy) error {
	return dp.policyMgr.AddPolicies(policies)
}

func (dp *DataPlane) RemovePolicies(policyName string) error {
	return dp.policyMgr.RemovePolicies(policyName)
}

func (dp *DataPlane) UpdatePolicies(policies *policies.NPMNetworkPolicy) error {
	return dp.policyMgr.UpdatePolicies(policies)
}
