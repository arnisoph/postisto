package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/arnisoph/postisto/pkg/log"
	imapClientPkg "github.com/emersion/go-imap/client"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

type Connection struct {
	Server        string `yaml:"server"`
	Port          int    `yaml:"port"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	IMAPS         bool   `yaml:"imaps"`
	Starttls      *bool  `yaml:"starttls"`
	TLSVerify     *bool  `yaml:"tlsverify"`
	TLSCACertFile string `yaml:"cacertfile"`

	imapClient *imapClientPkg.Client
}

func (conn *Connection) Connect() error {
	var imapClient *imapClientPkg.Client
	var err error

	// validate config
	if err := conn.validate(); err != nil {
		return err
	}

	// When not using IMAPS, enable STARTTLS by default
	if !conn.IMAPS && conn.Starttls == nil {
		var b bool
		conn.Starttls = &b
		*conn.Starttls = true
	}

	certPool := x509.NewCertPool()
	if conn.TLSCACertFile != "" {
		pemBytes, err := ioutil.ReadFile(conn.TLSCACertFile)
		if err != nil {
			log.Errorw("Failed to load CA cert file", err, "TLSCACertFile", conn.TLSCACertFile)
			return err
		}

		certPool.AppendCertsFromPEM(pemBytes)

	} else {
		certPool = nil
	}

	tlsConfig := &tls.Config{
		ServerName:         conn.Server,
		InsecureSkipVerify: !*conn.TLSVerify,
		MinVersion:         tls.VersionTLS12,
		RootCAs:            certPool,
	}

	if conn.IMAPS {
		if imapClient, err = imapClientPkg.DialTLS(fmt.Sprintf("%v:%v", conn.Server, conn.Port), tlsConfig); err != nil {
			log.Errorw("Failed to connect to server", err, "server", conn.Server)
			return err
		}
	} else {
		if imapClient, err = imapClientPkg.Dial(fmt.Sprintf("%v:%v", conn.Server, conn.Port)); err != nil {
			log.Errorw("Failed to connect to server", err, "server", conn.Server)
			return err
		}

		if *conn.Starttls {
			if err = imapClient.StartTLS(tlsConfig); err != nil {
				log.Errorw("Failed to initiate TLS session after connecting to server (STARTTLS)", err, "server", conn.Server)
				return err
			}
		}
	}

	if log.GetLogLevel() == "trace" {
		imapClient.SetDebug(os.Stderr)
	}

	if err = imapClient.Login(conn.Username, conn.Password); err != nil {
		log.Errorw("Failed to login to server", err, "server", conn.Server, "username", conn.Username)
		return err
	}

	conn.imapClient = imapClient
	return err
}

func (conn *Connection) Disconnect() error {

	if conn.imapClient == nil {
		// no connection
		return nil
	}
	return conn.imapClient.Logout()
}

func (conn *Connection) validate() error {
	if conn.Server == "" {
		return errors.Errorf("server not set in account config")
	}

	if conn.Port == 0 {
		return errors.Errorf("port is not set in account config")
	}

	if conn.Username == "" {
		return errors.Errorf("username not set in account config")
	}

	if conn.TLSVerify == nil {
		var b bool
		conn.TLSVerify = &b
		*conn.TLSVerify = true
	}

	return nil
}
