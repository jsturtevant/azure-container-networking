package policies

import (
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ipsets"
	networkingv1 "k8s.io/api/networking/v1"
)

type NPMNetworkPolicy struct {
	Name string
	// PodSelectorIPSets holds all the IPSets generated from Pod Selector
	PodSelectorIPSets []*ipsets.IPSet
	// OtherIPSets holds all IPSets generated from policy
	// except for pod selector IPSets
	OtherIPSets []*ipsets.IPSet
	ACLs        []*ACLPolicy
	// Making this a podKey instead should be
	// use NPMPod obj
	Pods  []string
	RawNP *networkingv1.NetworkPolicy
}

type ACLPolicy struct { // Iptable rules
	PolicyID  string
	Comment   string
	SrcList   []SetInfo
	DstList   []SetInfo
	Target    Verdict
	Direction Direction
	SrcPorts  []Ports
	DstPorts  []Ports
	Protocol  string
}

// SetInfo helps capture additional details in a matchSet
// exmaple match set in linux:
//             ! azure-npm-123 src,src
// "!" this indicates a negative match of an IPset for src,src
// Included flag captures the negative or positive match
// MatchType captures match flags
type SetInfo struct {
	IPSet     *ipsets.IPSet
	Included  bool
	MatchType string // match type can be “src”, “src,dst” or “dst,dst” etc
}

type Ports struct {
	Port    int64
	EndPort int64
}

type Verdict string
type Direction string

const (
	Ingress Direction = "IN"
	Egress  Direction = "OUT"
	Both    Direction = "BOTH"

	Allowed Verdict = "ALLOW"
	Dropped Verdict = "DROP"
)
