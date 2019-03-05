//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package system

// NetworkSession network session
type NetworkSession struct {
	hostID string
}

// NewNetworkSession create a new network session
func NewNetworkSession(hostID string) *NetworkSession {
	return &NetworkSession{
		hostID: hostID,
	}
}

func (s *NetworkSession) SystemID() (string, error) {
	return s.hostID, nil
}
