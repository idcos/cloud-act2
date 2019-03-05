//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package define

const (
	Deleted = "deleted"

	// host_exec_status
	Init  = "init"
	Doing = "doing"
	Done  = "done"

	// master status
	Running = "running"
	Stopped = "stopped"

	//Redis Channels
	RecordDone = "RECORD_DONE"
)

const (
	TokenExpired    = 0
	TokenWillExpire = 1
	TokenNotExpire  = 2
)

const (
	Yes = "Y"
	No  = "N"
)
