package ui

import (
	"go.bryk.io/pkg/did"
)

// Return the list of ALGO addresses linked to the provided identifier.
func getAlgoAddress(id *did.Identifier) []addressEntry {
	var result = []addressEntry{}
	svc := id.Service("algo-connect")
	if svc == nil {
		return result
	}
	var addresses []algoDestination
	ext := did.Extension{
		ID:      "algo-address",
		Version: "0.1.0",
	}
	if err := svc.GetExtension(ext.ID, ext.Version, &addresses); err != nil {
		return result
	}
	for _, entry := range addresses {
		result = append(result, addressEntry{
			Address: entry.Address,
			Network: entry.Network,
			Enabled: true,
		})
	}
	return result
}

// Return an empty service entry for `algo-connect`.
func newServiceEntry() *did.ServiceEndpoint {
	return &did.ServiceEndpoint{
		ID:       "algo-connect",
		Type:     "AlgorandExternalService",
		Endpoint: "https://did.algorand.foundation",
		Extensions: []did.Extension{
			{
				ID:      "algo-address",
				Version: "0.1.0",
				Data:    []algoDestination{},
			},
		},
	}
}
