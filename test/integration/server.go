package integration

import (
	"context"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gopkg.in/redis.v4"
	"math/rand"
	"testing"
	"time"
	"unsafe"
)

const MaxTestMailCount = 17

func NewAccount(t *testing.T, host string, username string, password string, port int, starttls bool, imaps bool, tlsverify bool, cacertfile *string, redisPort int) *config.Account {
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
			Server:   host,
			Port:     port,
			Username: username,
			Password: password,

			IMAPS:         imaps,
			Starttls:      &starttls,
			TLSVerify:     &tlsverify,
			TLSCACertFile: *cacertfile,
		},
		InputMailbox:    &inputMailbox,
		FallbackMailbox: &fallbackMailbox,
	}

	var err error
	acc.Connection.Username, err = NewIMAPUser(acc.Connection.Server, acc.Connection.Username, acc.Connection.Password, redisPort)
	require.NoError(err)

	return &acc
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

func NewIMAPUser(host string, username string, password string, redisPort int) (string, error) {
	if username == "" {
		username = fmt.Sprintf("test-%v@example.com", RandString(5))
	}

	redisClient, err := newRedisClient(host, redisPort)
	if err != nil {
		return username, err
	}

	dbs := [2]string{"userdb", "passdb"}
	for _, db := range dbs {
		key := fmt.Sprintf("dovecot/%v/%v", db, username)
		value := fmt.Sprintf(`{"uid":"65534","gid":"65534","home":"/tmp/%[1]v","username":"%[1]v","password":"%[2]v"}`, username, password)

		if err := redisClient.Set(key, value, 0).Err(); err != nil {
			return username, err
		}
	}

	return username, nil
}

func newRedisClient(host string, port int) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", host, port),
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping().Result()

	return redisClient, err
}

type TestContainer struct {
	IP    string
	Imap  int
	Imaps int
	Redis int

	Context   context.Context
	Container testcontainers.Container
}

func NewTestContainer() TestContainer {
	ctx := context.Background()
	waitPort, _ := nat.NewPort("tcp", "143")

	req := testcontainers.ContainerRequest{
		Image:        "bechtoldt/tabellarius_tests-docker",
		ExposedPorts: []string{"143/tcp", "993/tcp", "6379/tcp"},
		WaitingFor:   wait.ForListeningPort(waitPort),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		log.Fatal("Failed to start test Container", err)
	}

	//defer container.Terminate(ctx)
	ip, err := container.Host(ctx)
	if err != nil {
		log.Fatal("Failed to get test Container IP addr", err)
	}

	imap, err := container.MappedPort(ctx, "143")
	if err != nil {
		log.Fatal("Failed to get test Container port", err)
	}

	imaps, err := container.MappedPort(ctx, "993")
	if err != nil {
		log.Fatal("Failed to get test Container port", err)
	}

	kv, err := container.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatal("Failed to get test Container port", err)
	}

	return TestContainer{
		IP:        ip,
		Imap:      imap.Int(),
		Imaps:     imaps.Int(),
		Redis:     kv.Int(),
		Context:   ctx,
		Container: container}
}

func DeleteContainer(testContainer TestContainer) error {
	return testContainer.Container.Terminate(testContainer.Context)
}
