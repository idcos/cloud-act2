//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package encoding

import (
	"bytes"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"encoding/base64"
	"fmt"
)

func newEncoderTransformer(encodingName string) (transform.Transformer, error) {
	encodingName = strings.ToLower(encodingName)
	switch encodingName {
	case "cp936":
		fallthrough
	case "gbk":
		return simplifiedchinese.GBK.NewEncoder(), nil
	case "gb18030":
		return simplifiedchinese.GB18030.NewEncoder(), nil
	case "jis":
		return japanese.ShiftJIS.NewEncoder(), nil
	case "utf8":
		fallthrough
	case "utf-8":
		return unicode.UTF8.NewEncoder(), nil
	default:
		return nil, errors.New("unsupported encoding name " + encodingName)
	}
	return nil, errors.New("unsupported encoding name " + encodingName)
}

func newDecoderTransformer(decodingName string) (transform.Transformer, error) {
	decodingName = strings.ToLower(decodingName)
	switch decodingName {
	case "cp936":
		fallthrough
	case "gbk":
		return simplifiedchinese.GBK.NewDecoder(), nil
	case "gb18030":
		return simplifiedchinese.GB18030.NewDecoder(), nil
	case "jis":
		return japanese.ShiftJIS.NewDecoder(), nil
	case "utf8":
		fallthrough
	case "utf-8":
		return unicode.UTF8.NewDecoder(), nil
	default:
		return nil, errors.New("unsupported decoding name" + decodingName)
	}
	return nil, errors.New("unsupported decoding name" + decodingName)
}

func EncodingTo(value []byte, encoding string) ([]byte, error) {
	var b bytes.Buffer

	encoding = strings.TrimSpace(strings.ToLower(encoding))

	transformer, err := newEncoderTransformer(encoding)
	if err != nil {
		return nil, err
	}

	writer := transform.NewWriter(&b, transformer)
	writer.Write(value)
	writer.Close()

	return b.Bytes(), nil
}

func GB18030ToUTF8(value []byte) ([]byte, error) {
	return DecodingTo(value, "utf-8")
}

func UTF8ToGB18030(value []byte) ([]byte, error) {
	return EncodingTo(value, "gb18030")
}

func DecodingTo(value []byte, decoding string) ([]byte, error) {
	var b bytes.Buffer

	decoding = strings.TrimSpace(strings.ToLower(decoding))

	transformer, err := newDecoderTransformer(decoding)
	if err != nil {
		return nil, err
	}

	writer := transform.NewWriter(&b, transformer)
	writer.Write(value)
	writer.Close()

	return b.Bytes(), nil

}

func DataURIEncode(contentType, charset string, content []byte) string {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if charset == "" {
		charset = "base64"
	}

	base64Encode := base64.StdEncoding.EncodeToString(content)

	return fmt.Sprintf("data:%s;%s,%s", contentType, charset, base64Encode)
}
