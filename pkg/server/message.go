package server

import (
	"fmt"
	imapUtil "github.com/emersion/go-imap"
	mailUtil "github.com/emersion/go-message/mail"
	"strings"
)

type RawMessage imapUtil.Message

type Message struct {
	RawMessage imapUtil.Message
	Headers    MessageHeaders
}
type MessageHeaders map[string]interface{}

func NewMessage(rawMail *imapUtil.Message, headers MessageHeaders) *Message {
	return &Message{RawMessage: *rawMail, Headers: headers}
}

func parseMessageHeaders(rawMessage *imapUtil.Message) (MessageHeaders, error) {
	headers := MessageHeaders{}
	var err error

	// Create for mail parsing
	var section imapUtil.BodySectionName
	section.Specifier = imapUtil.HeaderSpecifier // Loads all headers only (no body)

	msgBody := rawMessage.GetBody(&section)
	if msgBody == nil {
		return headers, fmt.Errorf("server didn't returned message body for mail")
	}

	mr, err := mailUtil.CreateReader(msgBody)
	if err != nil {
		return headers, err
	}

	// Address Lists in headers
	addrFields := []string{"from", "to", "cc", "reply-to"}
	for _, fieldName := range addrFields {
		parsedList, err := parseAddrList(mr, fieldName, mr.Header.Get(fieldName))

		if err != nil {
			return nil, err
		} else {
			if parsedList == "" {
				// no need to set non-existant fields
				continue
			}

			headers[fieldName] = parsedList
		}
	}

	// Some other standard envelope headers
	headers["subject"] = strings.ToLower(fmt.Sprintf("%v", rawMessage.Envelope.Subject))
	headers["date"] = strings.ToLower(fmt.Sprintf("%v", rawMessage.Envelope.Date))
	headers["message-id"] = strings.ToLower(fmt.Sprintf("%v", rawMessage.Envelope.MessageId))

	// All the other headers
	alreadyHandled := []string{"subject", "date", "message-id"}
	alreadyHandled = append(alreadyHandled, addrFields...)
	fields := mr.Header.Fields()
	for {
		next := fields.Next()
		if !next {
			break
		}

		fieldName := strings.ToLower(fields.Key())
		fieldValue := strings.ToLower(fields.Value())

		if contains(alreadyHandled, fieldName) {
			// we maintain these headers elsewhere
			continue
		}

		switch val := headers[fieldName].(type) {
		case nil:
			// detected new header
			headers[fieldName] = fieldValue
		case string:
			headerList := []string{val, fieldValue}
			headers[fieldName] = headerList
		case []string:
			headers[fieldName] = append(val, fieldValue)
		}
	}

	/*
		// Process each message's part
		for {
			p, err := m.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				log.Printf("Got text: %v", string(b))
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				log.Printf("Got attachment: %v", filename)
			}

		}
	*/

	return headers, err
}

func parseAddrList(mr *mailUtil.Reader, fieldName string, fallback string) (string, error) {
	var fieldValue string
	addrs, err := mr.Header.AddressList(fieldName)

	if addrs == nil {
		// parsing failed, so return own or externally set fallback
		f := mr.Header.FieldsByKey(fieldName)
		if !f.Next() {
			return "", err
		} else {
			return strings.TrimSpace(fallback), nil
		}
	}

	if err != nil && err.Error() != "mail: missing '@' or angle-addr" { //ignore bad formated addrs
		// oh, real error
		return "", err
	}

	for _, addr := range addrs {
		formattedAddr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v <%v>", addr.Name, addr.Address)))
		if fieldValue != "" {
			fieldValue += ", "
		}
		fieldValue += formattedAddr
	}

	return strings.TrimSpace(fieldValue), err
}

func contains(s []string, e string) bool { //TODO do we really need to implement this?
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
