package main

import (
	"log"
	"net"

	"ssh-portfolio/config"
	"ssh-portfolio/sshsessionhandler"

	
)

func main() {

	

	sshServerConfig := sshsessionhandler.InitializeSSHConfig(config.GCfg.App.PrivateSSHKeyFile)


	tcpListener, err := net.Listen("tcp", ":"+config.GCfg.App.Port)
	if err != nil {
		log.Fatalf("Error listening on port %s: %v", config.GCfg.App.Port, err)
	}

	log.Printf("Waiting for SSH client connections (%v)...", tcpListener.Addr().String())
	for {

		clientConnection, err := tcpListener.Accept()
		if err != nil {
			log.Print("Unable to accept client connection: ", err)
			continue
		}

		go sshsessionhandler.HandleConnection(clientConnection, sshServerConfig)
	}
}














