//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package common

//Pagination 分页对象
type Pagination struct {
	PageNo     int64
	PageSize   int64
	PageCount  int64
	TotalCount int64
	List       interface{}
}
//GetPageCount 根据分页大小和记录总数计算分页总数
func GetPageCount(pageSize, totalCount int64) (pageCount int64) {
	pageCount = totalCount / pageSize
	if totalCount%pageSize > 0 {
		pageCount++
	}
	return
}
