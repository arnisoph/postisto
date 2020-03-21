package main

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestNewApp(t *testing.T) {
	app := newApp()
	require := require.New(t)

	// ACTUAL TESTS BELOW
	require.Equal("quite okay mail-sorting", app.Usage)
}

func TestRunApp(t *testing.T) {
	require := require.New(t)

	// Create local test IMAP server
	testContainer := integration.NewTestContainer()
	acc := integration.NewAccount(t, testContainer.IP, "test", "test", testContainer.Imap, true, false, true, nil, testContainer.Redis)

	// Simulate new unsorted mails by uploading
	for _, mailNum := range []int{1, 2, 3, 4} {
		require.NotNil(acc)
		require.NotNil(acc.Connection)
		require.NotNil(*acc.InputMailbox)
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", mailNum), *acc.InputMailbox, nil))
	}

	// Write exposed port to disk
	portConfig :=
		"accounts:\n" +
			"  local_imap_server:\n" +
			"    enable: true\n" +
			"    connection:\n" +
			"      server: localhost\n" +
			"      username: test\n" +
			"      password: test\n" +
			fmt.Sprintf("      port: %v\n", testContainer.Imap) +
			"      cacertfile: ../../test/data/certs/ca.pem"
	require.NoError(ioutil.WriteFile("../../test/data/configs/valid/local_imap_server/TestStartApp/.local.acc.yaml", []byte(portConfig), 0644))

	// ACTUAL TESTS BELOW
	require.EqualError(runApp("does-not exist", "debug", false, 42, true), "lstat does-not exist: no such file or directory")
	require.NoError(runApp("../../test/data/configs/valid/local_imap_server/TestStartApp/", "debug", false, 42, true))

	// Verify results
	fetchedMails, err := acc.Connection.Search(*acc.InputMailbox, nil, []string{server.FlaggedFlag})
	require.Nil(err)
	require.Equal(0, len(fetchedMails), "Unexpected num of mails in mailbox %v", *acc.InputMailbox)

	fetchedMails, err = acc.Connection.Search("MyTarget", nil, []string{server.FlaggedFlag})
	require.Nil(err)
	require.Equal(3, len(fetchedMails), "Unexpected num of mails in mailbox %v", "MyTarget")

	fetchedMails, err = acc.Connection.Search("MailFilterTest-TestRegex", nil, []string{server.FlaggedFlag})
	require.Nil(err)
	require.Equal(1, len(fetchedMails), "Unexpected num of mails in mailbox %v", "MailFilterTest-TestRegex")
}
