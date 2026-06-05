package sshsessionhandler

import (
	"encoding/binary"
	"log"
	"net"
	"os"
	"sync"
	"ssh-portfolio/tui"

	"golang.org/x/crypto/ssh"
)

func InitializeSSHConfig(sshHostPrivateKeyPath string) *ssh.ServerConfig {

	parsedPrivateSSHKeyBytes, err := os.ReadFile(sshHostPrivateKeyPath)
	if err != nil {
		log.Fatal("Failed to read privateKey file: ", err)
	}

	privateKey, err := ssh.ParsePrivateKey(parsedPrivateSSHKeyBytes)
	if err != nil {
		log.Fatal("Failed to parse privateKey Contents: ", err)
	}

	sshServerConfig := &ssh.ServerConfig{

		//So that anyone could create an SSH session with the server without a password.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},
		//So that anyone could create an SSH session with any SSH Public Key
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, nil
		},
		NoClientAuth: true,
	}
	sshServerConfig.AddHostKey(privateKey)

	return sshServerConfig
}

func HandleConnection(connection net.Conn, sshServerConfig *ssh.ServerConfig) {
	if sshServerConfig == nil {
		return
	}

	sshConn, channels, GlobalRequests, err := ssh.NewServerConn(connection, sshServerConfig)
	if err != nil {
		log.Print("SSH Handshake failed : ", err)
		return
	}

	log.Printf("New SSH Connection accepted from %s (%s)", sshConn.RemoteAddr(), sshConn.User())

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ssh.DiscardRequests(GlobalRequests)
	}()

	for newChannel := range channels {

		if newChannel.ChannelType() != "session" {

			log.Printf("%s@%s sent channel request of type: %s. Discarding...", sshConn.RemoteAddr(), sshConn.User(), newChannel.ChannelType())
			newChannel.Reject(ssh.UnknownChannelType, "Unknown Channel Type.")
			continue
		}

		//Accept returns a bidirectional channel (reqResponseChannel) for reading and writing payload data, and a read-only control plane channel \
		//(the reason the type is a <-chan *ssh.Requests, instead of a standard ssh.Channel, since it is read only) for control requests like a pty-req

		reqResponseChannel, globalSessionRequests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Unable to accept SSH Session Request[%s@%s]:%v", sshConn.RemoteAddr(), sshConn.User(), err)
			return //Since if the SSH Channel.Accept() function fails, most likely the underlying SSH Connection is broken, hence no longer handle that TCP Connection.
		}

		go handleSessionChannelRequests(reqResponseChannel, globalSessionRequests)

	}

}

func handleSessionChannelRequests(reqResponseChannel ssh.Channel, globalSessionRequests <-chan *ssh.Request) {

	var termWidth, termHeight uint32
	for request := range globalSessionRequests {
		switch request.Type {
		case "shell":
			_, ok := tui.CreateModel("welcomeModel")
			if !ok {
				request.Reply(false,nil)
			}
			request.Reply(true, nil)
			
		case "pty-req":		// This is the request we have to handle to get our client terminal configurations. This is crucial for our use case.
			var ok bool
			termWidth, termHeight, ok = parsePtyRequest(request.Payload)
			if !ok {
				log.Printf("Received invalid dimensions (%d x %d) within session pty-req!", termWidth, termHeight)
				request.Reply(false, nil)
				continue
			}
			log.Printf("Dimensions Received: %d x %d", termWidth, termHeight)
			request.Reply(true, nil)

		default:
			log.Print("Received unsupported session channel request: ", request.Type)
			request.Reply(false, nil)
		}
	}
}

func handleUserInput(userRequests ssh.Channel) {
	for {
		inputBuffer := make([]byte, 256)
		inputLen, err := userRequests.Read(inputBuffer)
		if err == nil {
			if inputLen > 0 {
				log.Printf("User input(%d):%s", inputLen, inputBuffer)
			}
		} else {
			log.Printf("Error reading User Input.")
			return
		}
	}

}

func parsePtyRequest(payload []byte) (width uint32, height uint32, ok bool) {

	payloadLen := len(payload)
	if payloadLen < 16 {
		return 0, 0, false
	}

	termStringLen := binary.BigEndian.Uint32(payload[0:4])
	dimensionOffset := termStringLen + 4

	if (dimensionOffset + 16) > uint32(payloadLen) {
		return 0, 0, false
	}

	width = binary.BigEndian.Uint32(payload[dimensionOffset : dimensionOffset+4])
	height = binary.BigEndian.Uint32(payload[dimensionOffset+4 : dimensionOffset+8])
	if width > 0 && height > 0 {
		ok = true
		return
	}

	dimensionOffset += 8
	width = binary.BigEndian.Uint32(payload[dimensionOffset : dimensionOffset+4])
	height = binary.BigEndian.Uint32(payload[dimensionOffset+4 : dimensionOffset+8])
	if width > 0 && height > 0 {
		ok = true
		return
	}

	return 0, 0, false

}
