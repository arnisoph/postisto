package log_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/log"

	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitWithConfig(t *testing.T) {
	require := require.New(t)

	// ACTUAL TESTS BELOW

	// Prepare some test log events
	testLogging := func() {
		log.Debug("Testing a debug log event.")
		log.Debugw("Testing a debug log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
		log.Info("Testing an info log event.")
		log.Infow("Testing an info log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
		log.Error("Testing an error log event.", fmt.Errorf("test error"))
		log.Errorw("Testing an error log event.", fmt.Errorf("test error"), "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
	}

	log.Info("log in TRACE")
	require.NoError(log.InitWithConfig("trace", false))
	testLogging()

	log.Info("log in DEBUG")
	require.NoError(log.InitWithConfig("debug", false))
	testLogging()

	log.Info("log in INFO")
	require.NoError(log.InitWithConfig("info", false))
	testLogging()

	log.Info("log in INFO and with JSON")
	require.NoError(log.InitWithConfig("info", true))
	testLogging()

	log.Info("log in ERROR")
	require.NoError(log.InitWithConfig("error", false))
	testLogging()
}
