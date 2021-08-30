package dataplane

import (
	"fmt"

	"github.com/Microsoft/hcsshim/hcn"
)

const (
	// Windows specific constants
	AZURENETWORKNAME = "azure"
)

func (dp *DataPlane) initializeDataplane() error {
	fmt.Printf("Initializing dataplane for windows")

	// Get Network ID
	network, err := hcn.GetNetworkByName(AZURENETWORKNAME)


	return nil
}