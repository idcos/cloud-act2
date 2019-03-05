//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package define

const (
	MysqlDriver = "mysql"

	// logger lever
	Trace = "TRACE"
	Debug = "DEBUG"
	Info  = "INFO"

	Esc = 0x1b

	//Redis Command
	Publish   = "PUBLISH"
	Subscribe = "SUBSCRIBE"
)

const (
	ApplicationJSON = "application/json"
)

const (
	ProxyDefaultPort  = "5555"
	MasterDefaultPort = "6868"
	/* kernels >= 2.6.36 */
	// https://lwn.net/Articles/391226/
	OomScoreAdj = -1000
)
