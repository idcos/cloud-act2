//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

//FileParam
type FileParam struct {
	//Content，文件内容，由contentType来解释
	Content string `json:"content"`
	//ContentType，内容类型，url：content是json数组数据，元素为url，conf:content为内容
	ContentType string `json:"contentType"`
	//
	TargetPath  string `json:"targetPath"`
	Owner       string `json:"owner,omitempty"`
	Group       string `json:"group,omitempty"`
	Perm        string `json:"perm"`
}
