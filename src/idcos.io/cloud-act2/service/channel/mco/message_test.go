//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package mco

import (
	"testing"
	"fmt"
)

func TestGetMessageBody(*testing.T) {
	value := `BAh7CzoNc2VuZGVyaWRJIg8xMC4wLjAuMTIzBjoGRVQ6DnJlcXVlc3RpZDA6EHNlbmRlcmFnZW50SSIKc2hlbGwGOwZUOgxtc2d0aW1lbCsHbs2PWzoJYm9k
eSIBqQQIewg6D3N0YXR1c2NvZGVpADoOc3RhdHVzbXNnSSIHT0sGOgZFVDoJZGF0YXsJOgtzdGRvdXRJIksgMjA6MzQ6NTQgdXAgNSBkYXlzLCAgNDoxMiwg
IDMgdXNlcnMsICBsb2FkIGF2ZXJhZ2U6IDAuMDgsIDAuMTEsIDAuMTcKBjsHVDoLc3RkZXJySSIABjsHVDoMc3VjY2Vzc1Q6DWV4aXRjb2RlaQA6CWhhc2hJ
IiUxYzg1MGQ5NWNlMTU1ZDEzNzhhM2I0MzY1MWI3YmY3NwY7BkY=`
	msg, err := GetMcoMessageResult([]byte(value))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(msg)
}
