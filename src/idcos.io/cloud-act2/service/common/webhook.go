//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

//WebHookAddForm web钩子新增对象
type WebHookAddForm struct {
	Event         string            `json:"event"`
	Uri           string            `json:"uri"`
	Token         string            `json:"token"`
	CustomHeaders map[string]string `json:"customHeaders"`
	CustomData    interface{}       `json:"customData"`
}

//WebHookUpdateForm web钩子修改对象
type WebHookUpdateForm struct {
	ID            string            `json:"id"`
	Event         string            `json:"event"`
	Uri           string            `json:"uri"`
	Token         string            `json:"token"`
	CustomHeaders map[string]string `json:"customHeaders"`
	CustomData    interface{}       `json:"customData"`
}

func ParamToPayload(param ProxyJobExecParam) ParamPayload {
	hostPayloads := make([]HostPayload, len(param.ExecHosts))
	for i, execHost := range param.ExecHosts {
		hostPayloads[i] = HostPayload{
			IP:       execHost.HostIP,
			Port:     execHost.HostPort,
			EntityID: execHost.EntityID,
			IdcName:  execHost.IdcName,
			OsType:   execHost.OsType,
			Encoding: execHost.Encoding,
		}
	}

	return ParamPayload{
		ExecHosts: hostPayloads,
		Param:     param.ExecParam,
		Provider:  param.Provider,
	}
}
