package ssh

import (
	"fmt"
	gossh "golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
)

func publicKeyFile(file string) gossh.Signer {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH private key file %s", file))
		return nil
	}

	key, err := gossh.ParsePrivateKey(buffer)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}

	return key
}
