package policies

import (
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ipsets"
	networkingv1 "k8s.io/api/networking/v1"
)

type NPMNetworkPolicy struct {
	Name              string
	PodSelectorIPSets []*ipsets.IPSet
	OtherIPSets       []*ipsets.IPSet
	ACLs              []*ACLPolicy
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
