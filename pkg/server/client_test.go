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

	accs := map[string]*config.Account{
		"starttls":           integration.NewAccount(t, "", "", 10143, true, false, true, nil),
		"starttls_wrongport": integration.NewAccount(t, "", "", 42, true, false, true, nil),
		"imaps":              integration.NewAccount(t, "", "", 10993, false, true, true, nil),
		"imaps_wrongport":    integration.NewAccount(t, "", "", 42, false, true, true, nil),
		"nocacert":           integration.NewAccount(t, "", "", 10143, true, false, true, &nocacert),
		"badcacert":          integration.NewAccount(t, "", "", 10143, true, false, true, &badcacert),
		"badcacertpath":      integration.NewAccount(t, "", "", 10143, true, false, true, &badcacertpath),
	}

	defer func() {
		for _, acc := range accs {
			require.Nil(acc.Connection.Disconnect())
		}
	}()

	// ACTUAL TESTS BELOW
	var acc config.Account

	acc = *accs["starttls"]
	require.NotNil(acc)
	require.NoError(acc.Connection.Connect())

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
