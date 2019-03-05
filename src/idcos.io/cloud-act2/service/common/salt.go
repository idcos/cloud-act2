//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

// SaltEvent salt event data
// my have the following format
// {u'tag': 'salt/auth', u'data': {u'_stamp': u'2018-11-06T06:45:33.566637', u'act': u'accept', u'id': u'6C12A913-756C-4A6B-B149-35E6351BA939', u'pub': u'-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzvvKeF+topbBMjIXaq2L\nQ+sdOgbcCgkyzPkJ8cqSCVEeVBU6wf3rDJ3AvYj0xNFwxY7/78riBKaC38XE62lw\nejeFFF45+qtiY1WwzjaNCncUI/gSp9xjMX1+VouaV+XHsvMt5XQ65yMLbWZoBl5G\na/L3GzdWehEi+DcDUOjo+oxBuXgX5a5EO0HTVPiqqJKGlQ9w56pOK7+rbeqh+8yK\n6nr9sYTcE4EtJd1tdjo/v2nVd5H8+yy5zfEt/eZvPT1+0K4wwDwkq5p4DEH1hdfi\n0UxqAd9fbqZ90NGScdExfijqaaR2wKhVN8bHI6OsrSNVId9ItEgZDqyuCrpLh8K4\nKwIDAQAB\n-----END PUBLIC KEY-----', u'result': True}}
// {u'tag': 'minion/refresh/6C12A913-756C-4A6B-B149-35E6351BA939', u'data': {u'Minion data cache refresh': u'6C12A913-756C-4A6B-B149-35E6351BA939', u'_stamp': u'2018-11-06T06:45:34.009732'}}
// {u'tag': 'minion_start', u'data': {u'_stamp': u'2018-11-06T06:45:35.574612', u'pretag': None, u'cmd': u'_minion_event', u'tag': u'minion_start', u'data': u'Minion 6C12A913-756C-4A6B-B149-35E6351BA939 started at Tue Nov  6 14:45:35 2018', u'id': u'6C12A913-756C-4A6B-B149-35E6351BA939'}}
// {u'tag': 'salt/minion/6C12A913-756C-4A6B-B149-35E6351BA939/start', u'data': {u'_stamp': u'2018-11-06T06:45:35.720646', u'pretag': None, u'cmd': u'_minion_event', u'tag': u'salt/minion/6C12A913-756C-4A6B-B149-35E6351BA939/start', u'data': u'Minion 6C12A913-756C-4A6B-B149-35E6351BA939 started at Tue Nov  6 14:45:35 2018', u'id': u'6C12A913-756C-4A6B-B149-35E6351BA939'}}
type SaltEvent struct {
	Tag  string                 `json:"tag"`
	Data map[string]interface{} `json:"data"`
}
