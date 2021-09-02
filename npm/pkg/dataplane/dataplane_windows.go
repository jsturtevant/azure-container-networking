package dataplane

import (
	"fmt"

	"github.com/Microsoft/hcsshim/hcn"
)

const (
	// Windows specific constants
	AzureNetworkName = "azure"
)

func (dp *DataPlane) initializeDataPlane() error {
	fmt.Printf("Initializing dataplane for windows")

	// Get Network ID
	network, err := hcn.GetNetworkByName(AzureNetworkName)
	if err != nil {
		return err
	}

	dp.networkID = network.Id

	endpoints, err := hcn.ListEndpointsOfNetwork(dp.networkID)
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {
		fmt.Println(endpoint.Policies)
		ep := &NPMEndpoint{
			Name:            endpoint.Name,
			ID:              endpoint.Id,
			NetPolReference: make(map[string]struct{}),
		}

		dp.endpointCache[ep.Name] = ep
	}

	return nil
}

func (dp *DataPlane) resetDataPlane() error {
	return nil
}
