package server_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/arnisoph/postisto/test/integration"
	imapUtil "github.com/emersion/go-imap"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestUploadMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	// ACTUAL TESTS BELOW

	require.EqualError(acc.Connection.Upload("does-not-exit.txt", "INBOX", []string{}), "open does-not-exit.txt: no such file or directory")
	require.Error(acc.Connection.Upload("../../test/data/mails/empty-mail.txt", "INBOX", []string{})) //TODO flags/gmail?
}

func TestSearchAndFetchMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 3

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Test searching only
	uids, err := acc.Connection.Search("INBOX", nil, nil)
	require.NoError(err)
	require.Equal([]uint32{1, 2, 3}, uids)

	// Search in non-existing mailbox
	fetchedMails, err := acc.Connection.SearchAndFetch("non-existent", nil, nil)
	require.Error(err)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: non-existent"))
	require.Equal(0, len(fetchedMails))

	// Search in correct Mailbox now
	fetchedMails, err = acc.Connection.SearchAndFetch("INBOX", nil, nil)
	require.NoError(err)
	require.Equal(numTestmails, len(fetchedMails))
}

func TestSetMailFlags(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 1

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := acc.Connection.SearchAndFetch("INBOX", nil, nil)
	require.NoError(err)
	require.Equal(numTestmails, len(fetchedMails))

	// Test failed GetFlags (because of non-existing mailbox)
	_, err = acc.Connection.GetFlags("non-existing-mailbox", 0)
	require.Error(err)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: non-existing-mailbox"))

	// Set custom flags
	var flags []string

	// Add flags
	require.Nil(acc.Connection.SetFlags("INBOX", []uint32{fetchedMails[0].RawMessage.Uid}, "+FLAGS", []interface{}{"fooooooo", "asdasd", "$MailFlagBit0", server.FlaggedFlag}, false))
	flags, err = acc.Connection.GetFlags("INBOX", fetchedMails[0].RawMessage.Uid)
	require.NoError(err)
	require.ElementsMatch([]string{"fooooooo", "asdasd", "$mailflagbit0", server.FlaggedFlag}, flags)

	// Remove flags
	require.Nil(acc.Connection.SetFlags("INBOX", []uint32{fetchedMails[0].RawMessage.Uid}, "-FLAGS", []interface{}{"fooooooo", "asdasd"}, false))
	flags, err = acc.Connection.GetFlags("INBOX", fetchedMails[0].RawMessage.Uid)
	require.NoError(err)
	require.ElementsMatch([]string{"$mailflagbit0", server.FlaggedFlag}, flags)

	// Replace all flags with new list
	require.Nil(acc.Connection.SetFlags("INBOX", []uint32{fetchedMails[0].RawMessage.Uid}, "FLAGS", []interface{}{"123", "forty-two"}, false))
	flags, err = acc.Connection.GetFlags("INBOX", fetchedMails[0].RawMessage.Uid)
	require.NoError(err)
	require.ElementsMatch([]string{"123", "forty-two"}, flags)
}

func TestMoveMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 5

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), *acc.InputMailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := acc.Connection.SearchAndFetch(*acc.InputMailbox, nil, nil)
	require.Equal(numTestmails, len(fetchedMails))
	require.NoError(err)

	// Move mails arround
	err = acc.Connection.Move([]uint32{fetchedMails[0].RawMessage.Uid}, "INBOX", "MyTarget42")
	require.NoError(err)

	err = acc.Connection.Move([]uint32{fetchedMails[1].RawMessage.Uid}, "INBOX", "INBOX")
	require.NoError(err)

	err = acc.Connection.Move([]uint32{fetchedMails[2].RawMessage.Uid}, "INBOX", "MyTarget!!!")
	require.NoError(err)

	err = acc.Connection.Move([]uint32{fetchedMails[3].RawMessage.Uid}, "wrong-source", "MyTarget!!!")
	require.Error(err)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: wrong-source"))

	err = acc.Connection.Move([]uint32{fetchedMails[4].RawMessage.Uid}, "INBOX", "ütf-8 & 梦龙周")
	require.NoError(err)

	var uids []uint32
	uids, err = acc.Connection.Search("INBOX", nil, nil)
	require.NoError(err)
	require.EqualValues([]uint32{4, 6}, uids) // UID 1 moved, UID 2 became 6, UID 3 moved, UID 4 kept untouched, UID 5 moved
}

func TestDeleteMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 3

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), *acc.InputMailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := acc.Connection.SearchAndFetch(*acc.InputMailbox, nil, nil)
	require.Equal(numTestmails, len(fetchedMails))
	require.NoError(err)

	// Delete one mail
	err = acc.Connection.DeleteMsgs("does-not-exist", []uint32{fetchedMails[0].RawMessage.Uid}, true) // mailbox doesn't exist, can't be deleted
	require.Error(err)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: does-not-exist"))

	err = acc.Connection.DeleteMsgs("INBOX", []uint32{fetchedMails[1].RawMessage.Uid}, false) // not moved yet, flag, don't expunge yet
	require.NoError(err)
	flags, err := acc.Connection.GetFlags("INBOX", fetchedMails[1].RawMessage.Uid)
	require.NoError(err)
	require.EqualValues([]string{server.DeletedFlag}, flags)
	err = acc.Connection.DeleteMsgs("INBOX", []uint32{fetchedMails[1].RawMessage.Uid}, true) // not moved yet, flag & expunge
	require.NoError(err)

	var uids []uint32
	uids, err = acc.Connection.Search("INBOX", nil, nil)
	require.NoError(err)
	require.EqualValues([]uint32{1, 3}, uids) // UID 1 kept untouched, UID 2 deleted, UID 3 kept untouched
}

func TestParseMailHeaders(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 5

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), *acc.InputMailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := acc.Connection.SearchAndFetch(*acc.InputMailbox, nil, nil)
	require.NoError(err)
	require.Equal(numTestmails, len(fetchedMails))

	// Verify parsed fields (header)
	parserTests := []struct {
		from     string
		to       string
		subject  string
		date     string
		received []string
	}{
		{ // #1
			from:    `youth4work <admin@youth4work.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "your registration at www.youth4work.com",
		},
		{ // #2
			from:    `rachit <rachit.jain@youth4work.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "cyberzonec",
		},
		{ // #3
			from:    `rachit <rachit.jain@youth4work.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "contact ranked talent for hire",
		},
		{ // #4
			from:    `bigrock.com sales team <automail@bigrock.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "customer sign up",
		},
		{ // #5
			from:    `invalid-address`,
			to:      `"mr. ütf-8" <foo@bar.net>`,
			subject: "ütf-8 💩",
			received: []string{
				strings.ToLower("from mail-storage-2.main-hosting.eu by mail-storage-2 (Dovecot) with LMTP id Dab9NZICjVWyQAAA7jq/7w for <shubham@cyberzonec.in>; Fri, 26 Jun 2015 07:43:20 +0000"),
				strings.ToLower("from mx2.main-hosting.eu (mx-mailgw [10.0.25.254]) by mail-storage-2.main-hosting.eu (Postfix) with ESMTP id 984D62096064 for <shubham@cyberzonec.in>; Fri, 26 Jun 2015 07:43:20 +0000 (UTC)"),
				strings.ToLower("from a10-20.smtp-out.amazonses.com (a10-20.smtp-out.amazonses.com [54.240.10.20]) by mx2.main-hosting.eu ([Main-Hosting.eu Mail System]) with ESMTPS id 4AF912D695A for <shubham@cyberzonec.in>; Fri, 26 Jun 2015 07:43:20 +0000 (UTC)"),
			},
		},
	}

	// Boring standard headers
	for i := 0; i < numTestmails; i++ {
		require.Equal(parserTests[i].from, fetchedMails[i].Headers["from"], "Failed in test #%v (FROM)", i+1)
		require.Equal(parserTests[i].to, fetchedMails[i].Headers["to"], "Failed in test #%v (TO)", i+1)
		require.Equal(parserTests[i].subject, fetchedMails[i].Headers["subject"], "Failed in test #%v (SUBJECT)", i+1)
	}

	// Exciting custom headers in #5
	require.Equal("<0000014e2ed21bf8-035d1578-a0ac-4afe-a3cb-ea4e65b92143-000000@email.amazonses.com>", fetchedMails[4].Headers["message-id"])
	require.Equal("<http://emailsparrow.com/unsubscribe.php?m=2392014&c=6508a8072980153786bbab4969679c2a&l=19&n=57154>", fetchedMails[4].Headers["list-unsubscribe"])
	require.Equal("2392014", fetchedMails[4].Headers["x-mailer-recptid"])
	require.Equal("foo <a@b.c>, baz <d@e.f>", fetchedMails[4].Headers["cc"])
	require.ElementsMatch(parserTests[4].received, fetchedMails[4].Headers["received"])
}

func TestConnection_List(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	// ACTUAL TESTS BELOW
	require.NoError(acc.Connection.CreateMailbox("foo"))
	require.NoError(acc.Connection.CreateMailbox("bar"))

	mailboxes, err := acc.Connection.List()
	require.NoError(err)
	mailboxesExpected := map[string]imapUtil.MailboxInfo{
		"Drafts": {Attributes: []string{"\\HasNoChildren", "\\Drafts"}, Delimiter: "/", Name: "Drafts"},
		"INBOX":  {Attributes: []string{"\\HasNoChildren"}, Delimiter: "/", Name: "INBOX"},
		"Junk":   {Attributes: []string{"\\HasNoChildren", "\\Junk"}, Delimiter: "/", Name: "Junk"},
		"Sent":   {Attributes: []string{"\\HasNoChildren", "\\Sent"}, Delimiter: "/", Name: "Sent"},
		"Trash":  {Attributes: []string{"\\HasNoChildren", "\\Trash"}, Delimiter: "/", Name: "Trash"},
		"bar":    {Attributes: []string{"\\HasNoChildren"}, Delimiter: "/", Name: "bar"},
		"foo":    {Attributes: []string{"\\HasNoChildren"}, Delimiter: "/", Name: "foo"},
	}
	require.Equal(mailboxesExpected, mailboxes)
}
