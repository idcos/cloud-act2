//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package define

const (
	WebHookEventJobCancel       = "job-cancel"
	WebHookEventJobRun          = "job-run"
	WebHookEventJobDone         = "job-done"
	WebHookEventJobSuccess      = "job-done-success"
	WebHookEventJobFail         = "job-done-fail"
	WebHookEventJobTimeout      = "job-done-timeout"
	WebHookEventProxyNew        = "proxy-new"
	WebHookEventProxyLose       = "proxy-lose"
	WebHookEventProxyReconnect  = "proxy-reconnect"
	WebHookEventHostChangeProxy = "host-change-proxy"
)

//WebHookEventList 所有的event列表
var WebHookEventList = []string{
	WebHookEventJobCancel,
	WebHookEventJobRun,
	WebHookEventJobDone,
	WebHookEventJobSuccess,
	WebHookEventJobFail,
	WebHookEventJobTimeout,
	WebHookEventProxyNew,
	WebHookEventProxyLose,
	WebHookEventProxyReconnect,
	WebHookEventHostChangeProxy,
}

const (
	WebHookHeaderToken = "X-Act2-Token"
	WebHookHeaderEvent = "X-Act2-Event"
	WebHookHeaderHost  = "host"
)
