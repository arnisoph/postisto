package integration

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/stretchr/testify/require"
	"gopkg.in/redis.v4"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

const MaxTestMailCount = 17

func NewAccount(t *testing.T, username string, password string, port int, starttls bool, imaps bool, tlsverify bool, cacertfile *string) *config.Account {

	require := require.New(t)

	require.NoError(log.InitWithConfig("debug", false))

	if cacertfile == nil {
		defaultcacert := "../../test/data/certs/ca.pem"
		cacertfile = &defaultcacert
	}

	inputMailbox := "INBOX"
	fallbackMailbox := "INBOX"
	acc := config.Account{
		Enable: true,
		Connection: server.Connection{
			Server:   "localhost",
			Port:     port,
			Username: NewUsername(username),
			Password: password,

			IMAPS:         imaps,
			Starttls:      &starttls,
			TLSVerify:     &tlsverify,
			TLSCACertFile: *cacertfile,
		},
		InputMailbox:    &inputMailbox,
		FallbackMailbox: &fallbackMailbox,
	}

	redisClient, err := newRedisClient()
	require.Nil(err)

	err = newIMAPUser(&acc, redisClient)
	require.Nil(err)

	return &acc
}

func NewStandardAccount(t *testing.T) *config.Account {
	return NewAccount(t, "", "test", 10143, true, false, true, nil)
}

func RandString(n int) string { // https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func NewUsername(u string) string {
	if u != "" {
		return u
	}

	return fmt.Sprintf("test-%v@example.com", RandString(5))
}

func newIMAPUser(acc *config.Account, redisClient *redis.Client) error {
	dbs := [2]string{"userdb", "passdb"}
	for _, db := range dbs {
		key := fmt.Sprintf("dovecot/%v/%v", db, acc.Connection.Username)
		value := fmt.Sprintf(`{"uid":"65534","gid":"65534","home":"/tmp/%[1]v","username":"%[1]v","password":"%[2]v"}`, acc.Connection.Username, acc.Connection.Password)

		if err := redisClient.Set(key, value, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func newRedisClient() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping().Result()

	return redisClient, err
}
