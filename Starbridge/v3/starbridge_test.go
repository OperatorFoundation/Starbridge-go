package Starbridge

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/aead/ecdh"
)

func TestStarbridge(t *testing.T) {
	addr := "127.0.0.1:1234"

	serverConfig, clientConfig, configError := GenerateNewConfigPair("127.0.0.1", 1234)
	if configError != nil {
		t.Fail()
		return
	}

	listener, listenError := serverConfig.Listen(addr)
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

	clientConn, clientConnError := clientConfig.Dial(addr)
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
	configError := GenerateConfigFiles("127.0.0.1", 1234)
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
	if keyError != nil {
		t.Fail()
	}
}

func TestKeyVerificationBadPrivateKey(t *testing.T) {
	keyExchange := ecdh.Generic(elliptic.P521()) 
	privateKey, _, keyError := keyExchange.GenerateKey(rand.Reader)
	if keyError != nil {
		t.Fail()
	}

	result := CheckPrivateKey(privateKey)

	if !result {
		t.Fail()
	}
}