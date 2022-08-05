package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"
)

type Mail struct {
	Recver, ContentType, Subject, Content string
}

func splitLines(data []byte, atEOF bool) (int, []byte, error) {
	for i := 0; i < len(data)-1; i++ {
		if bytes.Equal(data[i:i+2], []byte{'\r', '\n'}) {
			return i + 2, data[0:i], nil
		}
	}
	return 0, nil, nil
}

func ParseSmtp(conn net.Conn) (*Mail, error) {
	// make scanner
	scanner := bufio.NewScanner(conn)
	scanner.Split(splitLines)

	// welcome message
	io.WriteString(conn, "220 smtp2zoho\r\n")

	// recv HELO
	for {
		scanner.Scan()
		cmd := strings.Split(scanner.Text(), " ")
		major := strings.ToUpper(cmd[0])
		if len(cmd) < 2 || major != "HELO" {
			io.WriteString(conn, "503 Must issue a HELO command first.\r\n")
			continue
		}
		// resp Hello
		io.WriteString(conn, fmt.Sprintf("250 smtp2zoho Hello %s\r\n", cmd[1]))
		break
	}

	// recv MAIL FROM / RCPT TO
	var recver string
	mailFrom := false
	rcptTo := false
	for !(mailFrom && rcptTo) {
		scanner.Scan()
		cmd := strings.Split(scanner.Text(), ":")
		major := strings.ToUpper(cmd[0])
		if len(cmd) != 2 {
			io.WriteString(conn, "503 Must issue a \"MAIL FROM\" / \"RCPT TO\" command.\r\n")
			continue
		}

		switch major {
		case "MAIL FROM":
			io.WriteString(conn, fmt.Sprintf("250 OK %s\r\n", cmd[1]))
			mailFrom = true
		case "RCPT TO":
			recver = strings.Trim(cmd[1], " <>")
			io.WriteString(conn, fmt.Sprintf("250 OK %s\r\n", cmd[1]))
			rcptTo = true
		default:
			io.WriteString(conn, "503 Must issue a \"MAIL FROM\" / \"RCPT TO\" command.\r\n")
			continue
		}
	}

	// recv DATA command
	scanner.Scan()
	major := strings.ToUpper(scanner.Text())
	if major != "DATA" {
		io.WriteString(conn, "503 Must issue a DATA command.\r\n")
		conn.Close()
		return nil, ErrParsingFailed
	}
	// resp
	io.WriteString(conn, "354 Enter mail, end with \"<CR><LF>.<CR><LF>\".\r\n")

	// recv mail headers
	headers := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		header := strings.SplitN(line, ": ", 2)
		key := strings.ToLower(header[0])
		headers[key] = header[1]
	}

	// recv mail content
	lines := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "." {
			break
		}
		if len(line) >= 2 && line[0] == '.' {
			line = line[1:]
		}
		lines = append(lines, line)
	}
	content := strings.Join(lines, "\r\n")

	// parse mail info
	rawSubject := headers["subject"]
	var subject string
	if rawSubject[0:2] == "=?" {
		// accept UTF-8 with base64 encoding
		length := len(rawSubject)
		encoded := rawSubject[10 : length-2]
		b, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			io.WriteString(conn, "503 Invalid mail format\r\n")
			conn.Close()
			return nil, ErrParsingFailed
		}
		subject = string(b)
	} else {
		subject = rawSubject
	}

	contentType := headers["content-type"]

	// resp
	io.WriteString(conn, "250 Message sent\r\n")

	// end session
	scanner.Scan()
	io.WriteString(conn, "221 Bye\r\n")

	// conbine result
	mail := &Mail{
		Recver:      recver,
		ContentType: contentType,
		Subject:     subject,
		Content:     content,
	}
	return mail, nil
}
