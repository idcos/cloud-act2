//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package arrays

import (
	"testing"
)

func TestArrayTrimSpace(t *testing.T) {
	v := []string{"\taaa", "bbb        ", " ccc ", "ddd\t"}
	newV := ArrayTrimSpace(v)

	if newV[0] != "aaa" {
		t.Error("not valid aaa")
	}

	if newV[1] != "bbb" {
		t.Error("not valid bbb")
	}

	if newV[2] != "ccc" {
		t.Error("not valid ccc")
	}

	if newV[3] != "ddd" {
		t.Error("not valid ddd")
	}
}

func TestSkipOnlySpaceString(t *testing.T) {
	v := []string{"\taaa", "bbb        ", "       ", "\n", "\t\t\t\t", " ccc ", "ddd\t"}
	newV := SkipOnlySpaceString(v)

	if newV[0] != "\taaa" {
		t.Error("not valid aaa")
	}

	if newV[1] != "bbb        " {
		t.Error("not valid bbb")
	}

	if newV[2] != " ccc " {
		t.Error("not valid ccc")
	}

	if newV[3] != "ddd\t" {
		t.Error("not valid ddd")
	}
}