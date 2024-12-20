package filter_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestGetUnsortedMails(t *testing.T) {
	require := require.New(t)

	testContainer := integration.NewTestContainer()
	acc := integration.NewAccount(t, testContainer.IP, "", "test", testContainer.Imap, true, false, true, nil, testContainer.Redis)
	const numTestmails = 2

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW
	testMessages, err := filter.GetUnsortedMsgs(&acc.Connection, *acc.InputMailbox, []string{imap.SeenFlag, imap.FlaggedFlag})
	require.NoError(err)
	require.Equal(2, len(testMessages))
}

func TestEvaluateFilterSetsOnMails(t *testing.T) {
	require := require.New(t)
	testContainer := integration.NewTestContainer()

	type targetStruct struct {
		name  string
		num   int
		flags []string
	}
	type parserTest struct {
		fallbackMsgNum int
		mailsToUpload  []int
		targets        []targetStruct
		skip           bool
	}
	tests := []parserTest{
		{ // #1
			fallbackMsgNum: 1,
			mailsToUpload:  []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
			},
		},
		{ // #2
			mailsToUpload: []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestAND", num: 1},
			},
		},
		{ // #3
			mailsToUpload: []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestRegex", num: 1},
			},
		},
		{ // #4
			mailsToUpload: []int{1, 2, 3, 8, 9, 14},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestMisc", num: 3},
			},
		},
		{ // #5
			mailsToUpload: []int{1, 2, 3, 13, 15, 16, 17},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestUnicodeFrom-梦龙周", num: 3},
				{name: "MailFilterTest-TestUnicodeSubject", num: 1},
			},
		},
		{ // #6
			mailsToUpload: []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-foo", num: 1},
				{name: "MailFilterTest-bar", num: 1},
			},
		},
		{ // #7
			fallbackMsgNum: 3,
			mailsToUpload:  []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "MailFilterTest-baz", num: 1},
				{name: "MailFilterTest-zab", num: 1},
			},
		},
		{ // #8
			fallbackMsgNum: 3,
			mailsToUpload:  []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "X-Postisto-MailFilterTest-lorem", num: 1},
				{name: "X-Postisto-MailFilterTest-ipsum", num: 1},
			},
			skip: true, // TODO gmail connection broken..
		},
		{ // #9
			mailsToUpload: []int{1, 2, 3},
			targets: []targetStruct{
				{name: "Trash", num: 3, flags: []string{imap.SeenFlag}},
			},
		},
	}

	for testNum, test := range tests {
		if !test.skip {
			log.Debug(fmt.Sprintf("Starting TestEvaluateFilterSetsOnMails #%v", testNum+1))
		} else {
			log.Debug(fmt.Sprintf("SKIPPING TestEvaluateFilterSetsOnMails #%v", testNum+1))
			continue
		}

		// Get config
		require.NoError(ioutil.WriteFile(fmt.Sprintf("../../test/data/configs/valid/local_imap_server/TestEvaluateFilterSetsOnMails-%v/.postisto.local_imap_server.pwd", testNum+1), []byte("test"), 0600))
		cfg, err := config.NewConfigFromFile(fmt.Sprintf("../../test/data/configs/valid/local_imap_server/TestEvaluateFilterSetsOnMails-%v/", testNum+1))
		require.NoError(err)

		acc := cfg.Accounts["local_imap_server"]
		filters := cfg.Filters["local_imap_server"]

		// Create new random user
		acc.Connection.Username, err = integration.NewIMAPUser(testContainer.IP, "", acc.Connection.Password, testContainer.Redis)
		require.NoError(err)
		require.Equal("test", cfg.Accounts["local_imap_server"].Connection.Password)

		if strings.Contains(acc.Connection.Server, "gmail") { //TODO tidy up
			acc.Connection.Username = os.Getenv("POSTISTO_GMAIL_TEST_ACC_USERNAME")
			acc.Connection.Password = os.Getenv("POSTISTO_GMAIL_TEST_ACC_PASSWORD")
		} else {
			acc.Connection.Server = testContainer.IP
			acc.Connection.Port = testContainer.Imap
		}

		// Set debug info for failed assertions
		debugInfo := map[string]string{"username": acc.Connection.Username, "testNum": fmt.Sprint(testNum + 1)}

		// Connect to IMAP server
		require.NoError(acc.Connection.Connect(), debugInfo)

		if strings.Contains(acc.Connection.Server, "gmail") {
			log.Debug("Detected gmail account. Going to cleanup...")
			uids, err := acc.Connection.Search("INBOX", nil, nil)
			require.NoError(err)
			if len(uids) > 0 {
				err = acc.Connection.DeleteMsgs("INBOX", uids, true)
				require.NoError(err)
			}

			mailBoxes, err := acc.Connection.List()
			require.NoError(err, debugInfo)
			for mailboxName, _ := range mailBoxes {
				if strings.Contains(strings.ToLower(mailboxName), "x-postisto") {
					require.NoError(acc.Connection.DeleteMailbox(mailboxName), debugInfo)
				}
			}
		}

		// Simulate new unsorted mails by uploading
		for i, mailNum := range test.mailsToUpload {
			require.NotNil(acc, debugInfo)
			require.NotNil(acc.Connection, debugInfo)
			require.NotNil(*acc.InputMailbox, debugInfo)
			require.NotEmpty(filters, debugInfo)
			require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", mailNum), *acc.InputMailbox, nil), debugInfo)

			var withoutFlags []string
			if !strings.Contains(acc.Connection.Server, "gmail") { // gmail does some extra magic, marking (some) new messages as "important"....
				withoutFlags = append(withoutFlags, server.FlaggedFlag)
			}

			// verify upload
			uploadedMails, err := acc.Connection.Search(*acc.InputMailbox, nil, withoutFlags)

			require.NoError(err, debugInfo)
			require.Len(uploadedMails, i+1, fmt.Sprintf("This (#%v) or one of the previous mail uploads failed!", i+1), debugInfo)

			if strings.Contains(acc.Connection.Server, "gmail") {
				//gmail flaggs APPENDed msgs. I don't know yet why.. //TODO
				require.NoError(acc.Connection.SetFlags(*acc.InputMailbox, uploadedMails, "-FLAGS", []interface{}{server.FlaggedFlag}, false), debugInfo)
			}
		}

		// ACTUAL TESTS BELOW

		// Baaaam
		require.NoError(filter.EvaluateFilterSetsOnMsgs(&acc.Connection, *acc.InputMailbox, []string{imap.SeenFlag, imap.FlaggedFlag}, *acc.FallbackMailbox, filters), debugInfo)

		fallbackMethod := "moving"
		if *acc.FallbackMailbox == *acc.InputMailbox || *acc.FallbackMailbox == "" {
			fallbackMethod = "flagging"
		}

		// Verify Source
		if fallbackMethod == "flagging" {
			fetchedMails, err := acc.Connection.Search(*acc.InputMailbox, nil, []string{server.FlaggedFlag})
			require.Nil(err, debugInfo)
			require.Equal(0, len(fetchedMails), "Unexpected num of mails in source %v", *acc.InputMailbox, debugInfo)
		} else {
			// fallback = moving
			fetchedMails, err := acc.Connection.Search(*acc.InputMailbox, nil, nil)
			require.Nil(err, debugInfo)
			require.Equal(0, len(fetchedMails), "Unexpected num of mails in source %v", *acc.InputMailbox, debugInfo)
		}

		// Verify Targets
		for _, target := range test.targets {
			// fallback = flagging
			fetchedMails, err := acc.Connection.Search(target.name, nil, nil)
			require.Nil(err, debugInfo)
			require.Equal(target.num, len(fetchedMails), "Unexpected num of mails in target %v", target.name, debugInfo)

			if len(target.flags) > 0 {
				for _, fetchedMail := range fetchedMails {
					flags, err := acc.Connection.GetFlags(target.name, fetchedMail)
					require.NoError(err, debugInfo)
					require.EqualValues(target.flags, flags, debugInfo)
				}
			}
		}

		// Verify fallback mailbox
		fallBackMsgs, err := acc.Connection.Search(*acc.FallbackMailbox, nil, nil)
		require.Nil(err, debugInfo)
		require.Equal(test.fallbackMsgNum, len(fallBackMsgs), debugInfo)

		// Disconnect - Hoooraaay!
		require.Nil(acc.Connection.Disconnect(), debugInfo)
	}
}
