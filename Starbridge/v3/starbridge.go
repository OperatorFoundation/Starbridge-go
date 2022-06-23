package Starbridge

import (
	"encoding/hex"
	"errors"
	"net"
	"strconv"

	replicant "github.com/OperatorFoundation/Replicant-go/Replicant/v3"
	"github.com/OperatorFoundation/Replicant-go/Replicant/v3/polish"
	"github.com/OperatorFoundation/Replicant-go/Replicant/v3/toneburst"
	pt "github.com/OperatorFoundation/shapeshifter-ipc/v3"
	"golang.org/x/net/proxy"
)

type TransportClient struct {
	Config  ClientConfig
	Address string
	Dialer  proxy.Dialer
}

type TransportServer struct {
	Config  ServerConfig
	Address string
	Dialer  proxy.Dialer
}

type ClientConfig struct {
	Address                   string `json:"serverAddress"`
	ServerPersistentPublicKey string `json:"serverPersistentPublicKey"`
}

type ServerConfig struct {
	ServerPersistentPrivateKey string `json:"serverPersistentPrivateKey"`
}

type starbridgeTransportListener struct {
	address  string
	listener *net.TCPListener
	config   ServerConfig
}

func newStarbridgeTransportListener(address string, listener *net.TCPListener, config ServerConfig) *starbridgeTransportListener {
	return &starbridgeTransportListener{address: address, listener: listener, config: config}
}

func (listener *starbridgeTransportListener) Addr() net.Addr {
	interfaces, _ := net.Interfaces()
	addrs, _ := interfaces[0].Addrs()
	return addrs[0]
}

// Accept waits for and returns the next connection to the listener.
func (listener *starbridgeTransportListener) Accept() (net.Conn, error) {
	conn, err := listener.listener.Accept()
	if err != nil {
		return nil, err
	}

	host, portString, splitError := net.SplitHostPort(listener.address)
	if splitError != nil {
		return nil, splitError
	}

	port, intError := strconv.Atoi(portString)
	if intError != nil {
		return nil, intError
	}

	if len(listener.config.ServerPersistentPrivateKey) != 64 {
		return nil, errors.New("incorrect key size")
	}

	keyBytes, keyError := hex.DecodeString(listener.config.ServerPersistentPrivateKey)
	if keyError != nil {
		return nil, keyError
	}

	replicantConfig := getServerConfig(host, port, keyBytes)

	return NewServerConnection(replicantConfig, conn)
}

// Close closes the transport listener.
// Any blocked Accept operations will be unblocked and return errors.
func (listener *starbridgeTransportListener) Close() error {
	return listener.listener.Close()
}

// Listen checks for a working connection
func (config ServerConfig) Listen(address string) (net.Listener, error) {
	addr, resolveErr := pt.ResolveAddr(address)
	if resolveErr != nil {
		return nil, resolveErr
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newStarbridgeTransportListener(address, ln, config), nil
}

// Dial connects to the address on the named network
func (config ClientConfig) Dial(address string) (net.Conn, error) {
	conn, dialErr := net.Dial("tcp", address)
	if dialErr != nil {
		return nil, dialErr
	}

	host, portString, splitError := net.SplitHostPort(config.Address)
	if splitError != nil {
		return nil, splitError
	}

	port, intError := strconv.Atoi(portString)
	if intError != nil {
		return nil, intError
	}

	if len(config.ServerPersistentPublicKey) != 64 {
		return nil, errors.New("incorrect key size")
	}

	keyBytes, keyError := hex.DecodeString(config.ServerPersistentPublicKey)
	if keyError != nil {
		return nil, keyError
	}

	replicantConfig := getClientConfig(host, port, keyBytes)
	transportConn, err := NewClientConnection(replicantConfig, conn)

	if err != nil {
		if conn != nil {
			_ = conn.Close()
		}
		return nil, err
	}

	return transportConn, nil
}

func NewClient(config ClientConfig, dialer proxy.Dialer) TransportClient {
	return TransportClient{
		Config:  config,
		Address: config.Address,
		Dialer:  dialer,
	}
}

func NewServer(config ServerConfig, address string, dialer proxy.Dialer) TransportServer {
	return TransportServer{
		Config:  config,
		Address: address,
		Dialer:  dialer,
	}
}

// Dial creates outgoing transport connection
func (transport *TransportClient) Dial() (net.Conn, error) {
	conn, dialErr := transport.Dialer.Dial("tcp", transport.Address)
	if dialErr != nil {
		return nil, dialErr
	}

	dialConn := conn
	host, portString, splitError := net.SplitHostPort(transport.Address)
	if splitError != nil {
		return nil, splitError
	}

	port, intError := strconv.Atoi(portString)
	if intError != nil {
		return nil, intError
	}

	if len(transport.Config.ServerPersistentPublicKey) != 64 {
		return nil, errors.New("incorrect key size")
	}

	keyBytes, keyError := hex.DecodeString(transport.Config.ServerPersistentPublicKey)
	if keyError != nil {
		return nil, keyError
	}

	replicantConfig := getClientConfig(host, port, keyBytes)
	transportConn, err := NewClientConnection(replicantConfig, conn)

	if err != nil {
		_ = dialConn.Close()
		return nil, err
	}

	return transportConn, nil
}

func (transport *TransportServer) Listen() (net.Listener, error) {
	addr, resolveErr := pt.ResolveAddr(transport.Address)
	if resolveErr != nil {
		return nil, resolveErr
	}

	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newStarbridgeTransportListener(transport.Address, ln, transport.Config), nil
}

func NewClientConnection(config replicant.ClientConfig, conn net.Conn) (net.Conn, error) {
	return replicant.NewClientConnection(conn, config)
}

func NewServerConnection(config replicant.ServerConfig, conn net.Conn) (net.Conn, error) {
	return replicant.NewServerConnection(conn, config)
}

func NewReplicantClientConnectionState(config replicant.ClientConfig) (*replicant.ConnectionState, error) {
	return replicant.NewReplicantClientConnectionState(config)
}

func NewReplicantServerConnectionState(config replicant.ServerConfig, polishServer polish.Server, conn net.Conn) (*replicant.ConnectionState, error) {
	return replicant.NewReplicantServerConnectionState(config, polishServer, conn)
}

func getClientConfig(host string, port int, serverPublicKey []byte) replicant.ClientConfig {
	polishClientConfig := polish.DarkStarPolishClientConfig{
		Host:            host,
		Port:            port,
		ServerPublicKey: serverPublicKey,
	}

	toneburstClientConfig := toneburst.StarburstConfig{
		FunctionName: "SMTPClient",
	}

	clientConfig := replicant.ClientConfig{
		Toneburst: toneburstClientConfig,
		Polish:    polishClientConfig,
	}

	return clientConfig
}

func getServerConfig(host string, port int, serverPrivateKey []byte) replicant.ServerConfig {
	polishServerConfig := polish.DarkStarPolishServerConfig{
		Host:             host,
		Port:             port,
		ServerPrivateKey: serverPrivateKey,
	}

	toneburstServerConfig := toneburst.StarburstConfig{
		FunctionName: "SMTPServer",
	}

	serverConfig := replicant.ServerConfig{
		Toneburst: toneburstServerConfig,
		Polish:    polishServerConfig,
	}

	return serverConfig
}
