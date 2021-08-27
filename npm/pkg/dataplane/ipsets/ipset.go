package ipsets

import (
	"fmt"

	"github.com/Azure/azure-container-networking/npm/util"
)

type IPSet struct {
	Name       string
	HashedName string
	Type       SetType
	// IpPodKey is used for setMaps to store Ips and ports as keys
	// and podKey as value
	IpPodKey map[string]string
	// This is used for listMaps to store child IP Sets
	MemberIPSets     map[string]*IPSet
	NetPolReferCount int
	IpsetReferCount  int
}

type SetType int32

const (
	Unknown                  SetType = 0
	NameSpace                SetType = 1
	KeyLabelOfNameSpace      SetType = 2
	KeyValueLabelOfNameSpace SetType = 3
	KeyLabelOfPod            SetType = 4
	KeyValueLabelOfPod       SetType = 5
	NamedPorts               SetType = 6
	NestedLabelOfPod         SetType = 7
	CIDRBlocks               SetType = 8
)

var SetType_name = map[int32]string{
	0: "Unknown",
	1: "NameSpace",
	2: "KeyLabelOfNameSpace",
	3: "KeyValueLabelOfNameSpace",
	4: "KeyLabelOfPod",
	5: "KeyValueLabelOfPod",
	6: "NamedPorts",
	7: "NestedLabelOfPod",
	8: "CIDRBlocks",
}

var SetType_value = map[string]int32{
	"Unknown":                  0,
	"NameSpace":                1,
	"KeyLabelOfNameSpace":      2,
	"KeyValueLabelOfNameSpace": 3,
	"KeyLabelOfPod":            4,
	"KeyValueLabelOfPod":       5,
	"NamedPorts":               6,
	"NestedLabelOfPod":         7,
	"CIDRBlocks":               8,
}

func (x SetType) String() string {
	return SetType_name[int32(x)]
}

func GetSetType(x string) SetType {
	return SetType(SetType_value[x])
}

type SetKind string

const (
	ListSet SetKind = "list"
	HashSet SetKind = "set"
)

func NewIPSet(name string, setType SetType) *IPSet {
	return &IPSet{
		Name:             name,
		HashedName:       util.GetHashedName(name),
		IpPodKey:         make(map[string]string),
		Type:             setType,
		MemberIPSets:     make(map[string]*IPSet, 0),
		NetPolReferCount: 0,
		IpsetReferCount:  0,
	}
}

func GetSetContents(set *IPSet) ([]string, error) {
	contents := make([]string, 0)
	setType := getSetKind(set)
	switch setType {
	case HashSet:
		for podIp := range set.IpPodKey {
			contents = append(contents, podIp)
		}
		return contents, nil
	case ListSet:
		for _, memberSet := range set.MemberIPSets {
			contents = append(contents, memberSet.HashedName)
		}
		return contents, nil
	default:
		return contents, fmt.Errorf("Unknown set type %s", setType)
	}
}

func getSetKind(set *IPSet) SetKind {
	switch set.Type {
	case CIDRBlocks:
		return HashSet
	case NameSpace:
		return HashSet
	case NamedPorts:
		return HashSet
	case KeyLabelOfPod:
		return HashSet
	case KeyValueLabelOfPod:
		return HashSet
	case KeyLabelOfNameSpace:
		return ListSet
	case KeyValueLabelOfNameSpace:
		return ListSet
	case NestedLabelOfPod:
		return ListSet
	default:
		return "unknown"
	}
}
