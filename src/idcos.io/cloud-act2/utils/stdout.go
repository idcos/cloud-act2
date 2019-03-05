//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package utils

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"os"
)

func WriterToStdtab(vals []interface{}) {

	var valsBuf bytes.Buffer
	for _, v := range vals {
		valsBuf.WriteString(fmt.Sprintf("%v", v))
		valsBuf.WriteString("\t")
	}
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, valsBuf.String())
	w.Flush()
}

func ArrToStdtab(title []string, vals [][]interface{}) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 2, 8, 1, '\t', 0)

	var titleBuf bytes.Buffer

	for _, str := range title {
		titleBuf.WriteString(str)
		titleBuf.WriteString("\t")
	}

	fmt.Fprintln(w, titleBuf.String())

	for _, val := range vals {
		var valsBuf bytes.Buffer
		for _, v := range val {
			valsBuf.WriteString(fmt.Sprintf("%v", v))
			valsBuf.WriteString("\t")
		}
		fmt.Fprintln(w, valsBuf.String())
	}

	w.Flush()
}
