package filter_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestApplyCommands(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 2

	require.NoError(acc.Connection.Connect())
	defer func() {
		require.Nil(acc.Connection.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	testMails, err := acc.Connection.SearchAndFetch("INBOX", nil, nil)
	require.Equal(numTestmails, len(testMails))
	require.NoError(err)

	// Apply commands
	cmds := make(filter.FilterOps)
	cmds["move"] = "MyTarget"
	cmds["add_flags"] = []interface{}{"add_foobar", "Bar", "$MailFlagBit0", server.FlaggedFlag}
	cmds["remove_flags"] = []interface{}{"set_foobar", "bar"}

	// Message 1
	require.Nil(filter.RunCommands(&acc.Connection, "INBOX", testMails[0].RawMessage.Uid, cmds))
	flags, err := acc.Connection.GetFlags("MyTarget", testMails[0].RawMessage.Uid)
	require.NoError(err)
	require.ElementsMatch([]string{"add_foobar", "$mailflagbit0", server.FlaggedFlag}, flags)

	// Message 2: replace all flags
	cmds["replace_all_flags"] = []interface{}{"42", "bar", "oO", "$MailFlagBit0", server.FlaggedFlag}
	require.Nil(filter.RunCommands(&acc.Connection, "INBOX", testMails[1].RawMessage.Uid, cmds))
	flags, err = acc.Connection.GetFlags("MyTarget", testMails[1].RawMessage.Uid)
	require.NoError(err)
	require.ElementsMatch([]string{"42", "bar", "oo", "$mailflagbit0", server.FlaggedFlag}, flags)

	// Upload fresh mail
	require.Nil(acc.Connection.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", 1), "INBOX", []string{}))

	// Load newly uploaded mail
	testMails, err = acc.Connection.SearchAndFetch("INBOX", nil, nil)
	require.Equal(1, len(testMails))
	require.NoError(err)

	// Apply cmd to this new mail 3 too
	cmds["replace_all_flags"] = []interface{}{"completly", "different"}
	require.Nil(filter.RunCommands(&acc.Connection, "INBOX", testMails[0].RawMessage.Uid, cmds))
	flags, err = acc.Connection.GetFlags("MyTarget", testMails[0].RawMessage.Uid)
	require.NoError(err)
	require.ElementsMatch([]string{"completly", "different"}, flags)

	// Verify resulting INBOX
	uids, err := acc.Connection.Search("INBOX", nil, nil)
	require.NoError(err)
	require.Empty(uids)

	// Verify resulting MyTarget
	uids, err = acc.Connection.Search("MyTarget", nil, nil)
	require.NoError(err)
	require.ElementsMatch([]uint32{1, 2, 3}, uids)
}
