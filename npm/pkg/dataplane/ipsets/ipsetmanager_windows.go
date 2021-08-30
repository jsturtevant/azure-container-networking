package ipsets

// SetPolicyTypes associated with SetPolicy. Value is IPSET.
type SetPolicyType string

const (
	SetPolicyTypeIpSet       SetPolicyType = "IPSET"
	SetPolicyTypeNestedIpSet SetPolicyType = "NESTEDIPSET"
)

// SetPolicySetting creates IPSets on network
type SetPolicySetting struct {
	Id     string
	Name   string
	Type   SetPolicyType
	Values string
}

func isValidIPSet(set *IPSet) error {
	if set.Name == "" {
		return fmt.Errorf("IPSet " + set.Name + " is missing Name")
	}

	if set.Type == "" {
		return fmt.Errorf("IPSet " + set.Type + " is missing Type")
	}

	if set.HashedName == "" {
		return fmt.Errorf("IPSet " + set.HashedName + " is missing HashedName")
	}

	return nil
}

func getSetPolicyType(set *IPSet) SetPolicyType {
	setKind := getSetKind(set)
	switch setKind {
	case ListSet:
		return SetPolicyTypeNestedIpSet
	case HashSet:
		return SetPolicyTypeIpSet
	default:
		return "Unknown"
	}
}

func convertToSetPolicy(set *IPSet) (*SetPolicySetting, error) {
	err := isValidIPSet(set)
	if err != nil {
		return err
	}

	return &SetPolicySetting{
		Id:     set.HashedName,
		Name:   set.Name,
		Type:   getSetPolicyType(set),
		Values: set.GetSetContents(),
	}
}
