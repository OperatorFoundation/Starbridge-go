# Operator Foundation

[Operator](https://operatorfoundation.org) makes usable tools to help people around the world with censorship, security, and privacy.

# Starbridge-go

Starbridge is a Pluggable Transport that requires only minimal configuration information from the user. Under the hood, it uses the [Replicant](https://github.com/OperatorFoundation/Replicant-go) Pluggable Transport technology for network protocol obfuscation. Replicant is more complex to configure, so Starbridge is a good starting point for those wanting to use the technology to circumvent Internet cenorship, but wanting a minimal amount of setup.

## Shapeshifter

The Shapeshifter project provides network protocol shapeshifting technology (also sometimes referred to as obfuscation). The purpose of this technology is to change the characteristics of network traffic so that it is not identified and subsequently blocked by network filtering devices.

There are two components to Shapeshifter: transports and the dispatcher. Each transport provides a different approach to shapeshifting. These transports are provided as a Go library which can be integrated directly into applications. The dispatcher is a command line tool which provides a proxy that wraps the transport library. It has several different proxy modes and can proxy both TCP and UDP network traffic.

If you are an application developer working in the Go programming language, then you probably want to use the transports library directly in your application. If you are an end user that is trying to circumvent filtering on your network or you are a developer that wants to add pluggable transports to an existing tool that is not written in the Go programming language, then you probably want the dispatcher. Please note that familiarity with executing programs on the command line is necessary to use this tool. You can find Shapeshifter Dispatcher here: <https://github.com/OperatorFoundation/shapeshifter-dispatcher>

If you are looking for a complete, easy-to-use VPN that incorporates shapeshifting technology and has a graphical user interface, consider [Moonbounce](https://github.com/OperatorFoundation/Moonbounce), an application for macOS which incorporates Shapeshifter without the need to write code or use the command line.

### Shapeshifter Transports
The transports implement the Pluggable Transports version 2.1 specification, which is available here: <https://www.pluggabletransports.info/spec/#build> Specifically, they implement the Go Transports API v2.1.

The purpose of the transport library is to provide a set of different transports. Each transport implements a different method of shapeshifting network traffic. The goal is for application traffic to be sent over the network in a shapeshifted form that bypasses network filtering, allowing the application to work on networks where it would otherwise be blocked or heavily throttled.

## Installation
Starbridge is written in the Go programming language. To compile it you need
to install Go:

<https://golang.org/doc/install>

If you already have Go installed, make sure it is a compatible version:

    go version

The version should be 1.17 or higher.

If you get the error "go: command not found", then trying exiting your terminal
and starting a new one.

In order to use Starbridge in your project, you must have Go modules enabled in your project. How to do this is
beyond the scope of this document. You can find more information about Go modules here: <https://blog.golang.org/using-go-modules>

To use in your project, simply import:

    import "github.com/OperatorFoundation/Starbridge-go/Starbridge/v3"
    
Your go build tools should automatically add this module to your go.mod and go.sum files. Otherwise, you can add it to the go.mod file directly. See the official Go modules guide for more information on this.    

Please note that the import path includes "/v3" to indicate that you want to use the version of the module compatible with the PT v3.0 specification. This is required by the Go modules guide.

When you build your project, it should automatically fetch the correct version of the transport module.

## Testing

To run the existing Starbridge test, cd into the directory containing the test file and run it
```
cd Starbridge-go/Starbridge/v3
go test
```

## Usage

Starbridge-go is written in the go programming language

Create a Starbridge client config

```
    clientConfig := ClientConfig{
		  Address:                   <address>,
		  ServerPersistentPublicKey: <ServerPersistentPublicKey>,
	  }
```

Create a Starbridge server config

```
serverConfig := ServerConfig{
		ServerPersistentPrivateKey: <ServerPersistentPrivateKey>,
	}
```

Create a listener with the server config

```
	listener, listenError := serverConfig.Listen(<address>)
```

Create a go func that accepts the server connection, reads into a buffer, writes to the client, then closes the server connection

```
	go func() {
		serverConn, serverConnError := listener.Accept()

		buffer := make([]byte, 4)
		numBytesRead, readError := serverConn.Read(buffer)

		numBytesWritten, writeError := serverConn.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})

		_ = listener.Close()
	}()
```

Dial the server address

```
	clientConn, clientConnError := clientConfig.Dial(<address>)
```

Write some bytes to the server

```
	writeBytes := []byte{0x0A, 0x11, 0xB0, 0xB1}
	bytesWritten, writeError := clientConn.Write(writeBytes)
```

Read the bytes sent from the server

```
	readBuffer := make([]byte, 4)
	bytesRead, readError := clientConn.Read(readBuffer)
```

Close the client connection

```
	_ = clientConn.Close()
 ```
  
The success of the test can be further verified by using a program like Wireshark to check the bytes received during the test.  If the test was successful, there will be an SMTP conversation, followed by some encrypted data on the data stream
