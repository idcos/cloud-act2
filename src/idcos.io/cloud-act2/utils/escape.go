//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package utils

func EscapeBashArgs(args string) string {
	return args
}

func EscapeArgs(args string, windows bool) string {
	if windows {
		return args
	}
	return EscapeBashArgs(args)
}
