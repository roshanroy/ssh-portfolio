package sshsessionhandler

import (
	"os"
	"io"
	"log"
	"net"
	"sync"
	"golang.org/x/crypto/ssh"
)

func InitializeSSHConfig(sshHostPrivateKeyPath string) (*ssh.ServerConfig) {
	
	
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
	for request := range globalSessionRequests {
		switch request.Type {
		case "shell":
			request.Reply(true, nil)
			responseBytesLen, err := io.WriteString(reqResponseChannel, "\r\n Hello World, SSH Portfolio working lessgooooo")
			go handleUserInput(reqResponseChannel)
			if err != nil {
				log.Print("Error writing response to channel:", err)
			} else {
				log.Printf("Response Written: %d bytes", responseBytesLen)
			}
		case "pty-req":
			request.Reply(true, nil) //Temporary as this is the request we have to handle to get our client terminal configurations. This is crucial for our use case.
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