//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package encoding

import (
	"idcos.io/cloud-act2/utils/fileutil"
	"testing"
	"github.com/pkg/errors"
	"fmt"
)

func TestGBKToUTF8(t *testing.T) {
	gbkValue := []byte{0xd6, 0xd0, 0xce, 0xc4}
	bytes, err := DecodingTo(gbkValue, "gbk")
	if err != nil {
		t.Error(err)
	}

	if string(bytes) != "中文" {
		t.Error("not equal")
	}
}

func TestUTF8ToGBK(t *testing.T) {
	utf8Value := "中文"
	bytes, err := EncodingTo([]byte(utf8Value), "gbk")
	if err != nil {
		t.Error(err)
	}

	expectValue := []byte{0xd6, 0xd0, 0xce, 0xc4}
	if len(bytes) != len(expectValue) {
		t.Error(errors.New("expectValue not equal bytes"))
	}

	for i := 0; i < len(bytes); i++ {
		if expectValue[i] != bytes[i] {
			t.Error("unequal")
		}
	}

}

func TestDataURIEncode(t *testing.T) {

	bytes, _ := fileutil.ReadFileToBytes("/tmp/a.out")
	//bytes, _ := ReadFileToBytes("/tmp/b.bb")

	fmt.Println(DataURIEncode("", "", bytes))
}
