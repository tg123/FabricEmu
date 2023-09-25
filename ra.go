package FabricEmu

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/tg123/phabrik/federation"
	"github.com/tg123/phabrik/transport"
)

type ReplicaAgent struct {
	server       *transport.Server
	listenaddr   string
	serverCertTp string
	clientCert   *tls.Certificate

	hostid string
	nodeid string

	deploymentDirectory string

	replicas sync.Map

	stateful           bool
	onNewReplicaOpened func(replica *Replica)
}

type ReplicaAgentConfig struct {
	OnNewReplicaOpened func(replica *Replica)
	Stateful           bool
}

func NewReplicaAgent(config ReplicaAgentConfig) (*ReplicaAgent, error) {
	ra := &ReplicaAgent{
		onNewReplicaOpened: config.OnNewReplicaOpened,
	}

	if err := ra.createServer(); err != nil {
		return nil, err
	}

	ra.hostid = uuid.New().String()
	ra.nodeid = federation.NodeIDFromMD5(ra.hostid).String()

	dir, err := os.MkdirTemp("", "sfemudeploy")
	if err != nil {
		return nil, err
	}

	ra.deploymentDirectory = dir
	ra.stateful = config.Stateful

	return ra, nil
}

func (h *ReplicaAgent) createServer() error {
	serverCert, err := generateCert()
	if err != nil {
		return err
	}

	clientCert, err := generateCert()
	if err != nil {
		return err
	}

	tlsconf := &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientAuth:   tls.RequestClientCert,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if !bytes.Equal(rawCerts[0], clientCert.Certificate[0]) {
				return fmt.Errorf("bad client cert")
			}
			return nil
		},
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	s, err := transport.Listen(l, transport.ServerConfig{
		Config: transport.Config{
			TLS: tlsconf,
		},
		MessageCallback: h.handler,
	})
	if err != nil {
		return err
	}

	h.server = s
	h.listenaddr = l.Addr().String()
	h.serverCertTp = fmt.Sprintf("%x", sha1.Sum(serverCert.Certificate[0]))
	h.clientCert = clientCert

	return nil
}

func (h *ReplicaAgent) Wait() error {
	return h.server.Serve()
}
