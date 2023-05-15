package Starbridge

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"net/smtp"
	"testing"

	"github.com/aead/ecdh"
)

func TestBadClientSMTP(t *testing.T) {
	_, priv, err := GenerateKeys()
	if err != nil {
		panic(err)
	}

	serverConfig := ServerConfig{
		ServerAddress:    "127.0.0.1:2525",
		ServerPrivateKey: *priv,
		Transport:        "Starbridge",
	}

	go func() {
		if err := listenAndServe(serverConfig); err != nil {
			fmt.Println("serve error: ", err)
		}
	}()

	c, err := smtp.Dial("127.0.0.1:2525")
	if err != nil {
		fmt.Println("SMTP Dial error")
		panic(err)
	}
	defer c.Close()

	if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		fmt.Println("StartTLS error")
		panic(err)
	}
}

func listenAndServe(serverConfig ServerConfig) error {
	l, err := serverConfig.Listen()
	if err != nil {
		fmt.Println("Server listen error")
		panic(err)
	}
	defer l.Close()

	acceptErr := make(chan error)
	go func() {
		for {
			conn, err := l.Accept()

			fmt.Println("Accepted a connection!")
			if err != nil {
				acceptErr <- err
				fmt.Println("Listener accept error: ", err)
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				buf := make([]byte, 1024)
				n, err := conn.Read(buf)
				if err != nil {
					fmt.Println("read error: ", err)
					return
				}

				fmt.Println("from client: ", base64.StdEncoding.EncodeToString(buf[:n]))
			}(conn)
		}
	}()

	return <-acceptErr
}

func TestStarbridge(t *testing.T) {
	serverConfig, clientConfig, configError := GenerateNewConfigPair("127.0.0.1:1234", nil)
	if configError != nil {
		t.Fail()
		return
	}

	listener, listenError := serverConfig.Listen()
	if listenError != nil {
		fmt.Println(listenError)
		t.Fail()
		return
	}

	go func() {
		fmt.Printf("listener type: %T\n", listener)
		serverConn, serverConnError := listener.Accept()
		if serverConnError != nil {
			fmt.Println(serverConnError)
			t.Fail()
			return
		}
		if serverConn == nil {
			fmt.Println("serverConn is nil")
			t.Fail()
			return
		}

		buffer := make([]byte, 4)
		numBytesRead, readError := serverConn.Read(buffer)
		if readError != nil {
			fmt.Println(readError)
			t.Fail()
			return
		}
		fmt.Printf("number of bytes read on server: %d\n", numBytesRead)
		fmt.Printf("serverConn type: %T\n", serverConn)

		// Send a response back to person contacting us.
		numBytesWritten, writeError := serverConn.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
		if writeError != nil {
			fmt.Println(writeError)
			t.Fail()
			return
		}
		fmt.Printf("number of bytes written on server: %d\n", numBytesWritten)

		_ = listener.Close()
	}()

	clientConn, clientConnError := clientConfig.Dial()
	if clientConnError != nil {
		fmt.Println(clientConnError)
		t.Fail()
		return
	}

	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	bytesWritten, writeError := clientConn.Write(writeBytes)
	if writeError != nil {
		fmt.Println(writeError)
		t.Fail()
		return
	}
	fmt.Printf("number of bytes written on client: %d\n", bytesWritten)

	readBuffer := make([]byte, 4)
	bytesRead, readError := clientConn.Read(readBuffer)
	if readError != nil {
		fmt.Println(readError)
		t.Fail()
		return
	}
	fmt.Printf("number of bytes read on client: %d\n", bytesRead)

	_ = clientConn.Close()
}

func TestStarbridgeCustomConfig(t *testing.T) {
	serverConfig := ServerConfig{
		ServerAddress:    "127.0.0.1:1234",
		ServerPrivateKey: "",
		Transport:        "Starbridge",
	}

	clientConfig := ClientConfig{
		ServerAddress:   "127.0.0.1:1234",
		ServerPublicKey: "",
		Transport:       "Starbridge",
	}

	listener, listenError := serverConfig.Listen()
	if listenError != nil {
		fmt.Println(listenError)
		t.Fail()
		return
	}

	go func() {
		fmt.Printf("listener type: %T\n", listener)
		serverConn, serverConnError := listener.Accept()
		if serverConnError != nil {
			fmt.Println(serverConnError)
			t.Fail()
			return
		}
		if serverConn == nil {
			fmt.Println("serverConn is nil")
			t.Fail()
			return
		}

		buffer := make([]byte, 4)
		numBytesRead, readError := serverConn.Read(buffer)
		if readError != nil {
			fmt.Println(readError)
			t.Fail()
			return
		}
		fmt.Printf("number of bytes read on server: %d\n", numBytesRead)
		fmt.Printf("serverConn type: %T\n", serverConn)

		// Send a response back to person contacting us.
		numBytesWritten, writeError := serverConn.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
		if writeError != nil {
			fmt.Println(writeError)
			t.Fail()
			return
		}
		fmt.Printf("number of bytes written on server: %d\n", numBytesWritten)

		_ = listener.Close()
	}()

	clientConn, clientConnError := clientConfig.Dial()
	if clientConnError != nil {
		fmt.Println(clientConnError)
		t.Fail()
		return
	}

	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	bytesWritten, writeError := clientConn.Write(writeBytes)
	if writeError != nil {
		fmt.Println(writeError)
		t.Fail()
		return
	}
	fmt.Printf("number of bytes written on client: %d\n", bytesWritten)

	readBuffer := make([]byte, 4)
	bytesRead, readError := clientConn.Read(readBuffer)
	if readError != nil {
		fmt.Println(readError)
		t.Fail()
		return
	}
	fmt.Printf("number of bytes read on client: %d\n", bytesRead)

	_ = clientConn.Close()
}

func TestConfigFileGenerate(t *testing.T) {
	configError := GenerateConfigFiles("127.0.0.1:1234", nil)
	if configError != nil {
		t.Fail()
	}
}

func TestKeyVerificationGoodKeys(t *testing.T) {
	keyExchange := ecdh.Generic(elliptic.P256())
	privateKey, publicKey, keyError := keyExchange.GenerateKey(rand.Reader)
	if keyError != nil {
		t.Fail()
	}

	if CheckPublicKey(publicKey) != nil {
		t.Fail()
	}

	if !CheckPrivateKey(privateKey) {
		t.Fail()
	}
}

func TestKeyVerificationBadPublicKey(t *testing.T) {
	publicKey := ecdh.Point{big.NewInt(9), big.NewInt(100)}

	keyError := CheckPublicKey(publicKey)
	if keyError == nil {
		t.Fail()
	}
}
