package config_test

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	require := require.New(t)

	var cfg *config.Config
	var err error

	// ACTUAL TESTS BELOW

	// NewConfigFromFile single file
	require.FileExists("../../test/data/configs/valid/accounts.yaml")

	cfg, err = config.NewConfigFromFile("../../test/data/configs/valid/accounts.yaml")
	require.NoError(err)
	require.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// NewConfigFromFile full config dir
	require.DirExists("../../test/data/configs/valid/")
	cfg, err = config.NewConfigFromFile("../../test/data/configs/valid/")
	require.NoError(err)
	require.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Test for readPasswordEnvFile
	require.NoError(ioutil.WriteFile("../../test/data/configs/valid/.postisto.readenv1.pwd", []byte("wh00pWh00p!"), 0600))
	require.NoError(ioutil.WriteFile("../../test/data/configs/valid/.postisto.readenv2.pwd", []byte("ütf-8 💩"), 0600))
	cfg, err = config.NewConfigFromFile("../../test/data/configs/valid/")
	require.Equal("wh00pWh00p!", cfg.Accounts["readenv1"].Connection.Password)
	require.Equal("ütf-8 💩", cfg.Accounts["readenv2"].Connection.Password)
	_, err = os.Stat("../../test/data/configs/valid/.postisto.readenv1.pwd")
	require.True(os.IsNotExist(err))
	_, err = os.Stat("../../test/data/configs/valid/.postisto.readenv2.pwd")
	require.True(os.IsNotExist(err))

	// Failed file/dir loading
	cfg, err = config.NewConfigFromFile("../../test/data/configs/does-not-exist")
	require.EqualError(err, "stat ../../test/data/configs/does-not-exist: no such file or directory")

	// Reading broken file
	cfg, err = config.NewConfigFromFile("../../test/data/configs/invalid/zero-file.yaml")
	require.EqualError(err, "yaml: control characters are not allowed")

	// Reading unaccessible file
	_, _ = os.Create("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml")
	require.Nil(os.Chmod("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml", 0000))
	cfg, err = config.NewConfigFromFile("../../test/data/configs/invalid-unreadable-file/")
	require.EqualError(err, "open ../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml: permission denied")
}
