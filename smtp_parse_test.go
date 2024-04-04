package main

import (
	"fmt"
	"bytes"
	"io"
	"net"
	"strings"
	"testing"
)

type MockConn struct {
	net.Conn
	buf *bytes.Buffer
}

func (conn *MockConn) Read(dst []byte) (int, error) {
	n, e := conn.buf.Read(dst)
	return n, e
}

func (conn *MockConn) Write(src []byte) (int, error) {
	return len(src), nil
}

func (conn *MockConn) Close() error {
	return nil
}

type SmtpParseTest struct {
	rawData    []byte
	expectMail *Mail
	expectErr  error
}

func newSmtpParseTestWithLines(lines []string, expectMail *Mail, expectErr error) SmtpParseTest {
	lines = append(lines, "")
	rawData := []byte(strings.Join(lines, "\r\n"))
	return SmtpParseTest{
		rawData:    rawData,
		expectMail: expectMail,
		expectErr:  expectErr,
	}
}

func TestSmtpParse(t *testing.T) {
	tests := []SmtpParseTest{
		newSmtpParseTestWithLines(
			[]string{
				"HELO localhost",
				"MAIL FROM:<send@example.com>",
				"RCPT TO:<recv@example.com>",
				"DATA",
				"From: Sender Name <send@example.com>",
				"Content-Type: text/html; charset=UTF-8",
				"Subject: Plain ASCII Subject Test",
				"",
				"This is a test",
				".",
			},
			&Mail{
				Recver:      "recv@example.com",
				ContentType: "text/html; charset=UTF-8",
				SenderName:  "Sender Name",
				Subject:     "Plain ASCII Subject Test",
				Content:     "This is a test",
			},
			nil,
		),
		newSmtpParseTestWithLines(
			[]string{
				"HELO localhost",
				"MAIL FROM:<send2@example.com>",
				"RCPT TO:<recv2@example.com>",
				"DATA",
				"From: =?UTF-8?B?5Y+R6YCB6ICF?= <send2@example.com>",
				"Content-Type: text/html; charset=UTF-8",
				"Subject: =?UTF-8?B?VVRGLTgg5rWL6K+V?=",
				"",
				"这是一条带有中文 UTF-8 邮件正文测试",
				"并且它有多行",
				".",
			},
			&Mail{
				Recver:      "recv2@example.com",
				ContentType: "text/html; charset=UTF-8",
				SenderName:  "发送者",
				Subject:     "UTF-8 测试",
				Content:     "这是一条带有中文 UTF-8 邮件正文测试\r\n并且它有多行",
			},
			nil,
		),
		newSmtpParseTestWithLines(
			[]string{
				"HELO",
			},
			nil,
			io.EOF,
		),

	}
	for i, test := range tests {
		succ, errMsg := testSmtpParseSingle(test)
		if !succ {
			t.Errorf("#%d test failed: %s", i, errMsg)
		}
	}
}

func testSmtpParseSingle(test SmtpParseTest) (bool, string) {
	conn := &MockConn{
		buf: bytes.NewBuffer(bytes.Clone(test.rawData)),
	}
	gotMail, gotErr := ParseSmtp(conn)

	if gotErr != test.expectErr {
		return false, fmt.Sprintf("expect err: %+v; got err: %+v", test.expectErr, gotErr)
	}

	if gotErr == nil && *gotMail != *test.expectMail {
		return false, fmt.Sprintf("expect mail: %+v; parsed mail: %+v", test.expectMail, gotMail)
	}

	return true, ""
}
