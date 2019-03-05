//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package aux

import (
	"fmt"
	"idcos.io/cloud-act2/define"
	"strings"
)

type Builder struct {
	strings.Builder
}

func (b *Builder) Green(value string) *Builder {
	fmt.Fprintf(b, "%c[32m%s%c[0m", define.Esc, value, define.Esc)
	return b
}

func (b *Builder) Red(value string) *Builder {
	fmt.Fprintf(b, "%c[31m%s%c[0m", define.Esc, value, define.Esc)
	return b
}

func (b *Builder) Append(value string) *Builder {
	fmt.Fprint(b, value)
	return b
}
