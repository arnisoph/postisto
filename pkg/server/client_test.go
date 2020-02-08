package server_test

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	require := require.New(t)

	nocacert := ""
	badcacert := "../../test/data/certs/bad-ca.pem"
	badcacertpath := "ca-doesnotexist.pem"

	testContainer := integration.NewTestContainer()
	accs := map[string]*config.Account{
		"starttls":           integration.NewAccount(t, testContainer.IP, "", "", testContainer.Imap, true, false, true, nil, testContainer.Redis),
		"starttls_wrongport": integration.NewAccount(t, testContainer.IP, "", "", 42, true, false, true, nil, testContainer.Redis),
		"imaps":              integration.NewAccount(t, testContainer.IP, "", "", testContainer.Imaps, false, true, true, nil, testContainer.Redis),
		"imaps_wrongport":    integration.NewAccount(t, testContainer.IP, "", "", 42, false, true, true, nil, testContainer.Redis),
		"nocacert":           integration.NewAccount(t, testContainer.IP, "", "", testContainer.Imap, true, false, true, &nocacert, testContainer.Redis),
		"badcacert":          integration.NewAccount(t, testContainer.IP, "", "", testContainer.Imap, true, false, true, &badcacert, testContainer.Redis),
		"badcacertpath":      integration.NewAccount(t, testContainer.IP, "", "", testContainer.Imap, true, false, true, &badcacertpath, testContainer.Redis),
	}

	defer func() {
		for _, acc := range accs {
			require.Nil(acc.Connection.Disconnect())
		}

		//require.NoError(integration.DeleteContainer(testContainer))
	}()

	// ACTUAL TESTS BELOW
	var acc config.Account

	acc = *accs["starttls"]

	// Test re-login of a completly new account
	require.NoError(acc.Connection.DeleteMsgs("INBOX", []uint32{42}, false))

	require.NotNil(acc)
	require.NoError(acc.Connection.Connect())

	// Test normal re-login
	require.NoError(acc.Connection.Disconnect())
	require.NoError(acc.Connection.DeleteMsgs("INBOX", []uint32{42}, false))

	acc.Connection.Password = "wrongpass"
	require.EqualError(acc.Connection.Connect(), "Authentication failed.")

	acc = *accs["starttls_wrongport"]
	require.Error(acc.Connection.Connect())

	acc = *accs["imaps"]
	require.NoError(acc.Connection.Connect())

	acc = *accs["imaps_wrongport"]
	require.Error(acc.Connection.Connect())

	if os.Getenv("USER") != "ab" {
		acc = *accs["nocacert"]
		require.EqualError(acc.Connection.Connect(), "x509: certificate signed by unknown authority")
	}

	acc = *accs["badcacert"]
	require.EqualError(acc.Connection.Connect(), "x509: certificate signed by unknown authority")

	acc = *accs["badcacertpath"]
	require.EqualError(acc.Connection.Connect(), "open ca-doesnotexist.pem: no such file or directory")
}
