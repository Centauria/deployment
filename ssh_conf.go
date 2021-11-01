package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// create human-readable SSH-key strings
func keyString(k ssh.PublicKey) string {
	return k.Type() + " " + base64.StdEncoding.EncodeToString(k.Marshal()) // e.g. "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY...."
}

func trustedHostKeyCallback(trustedKey map[string]string) ssh.HostKeyCallback {

	if len(trustedKey) == 0 {
		return func(name string, addr net.Addr, k ssh.PublicKey) (err error) {
			err = addTrustedHost(name, k)
			return
		}
	}

	return func(name string, addr net.Addr, k ssh.PublicKey) error {
		ks := keyString(k)
		if ks == trustedKey[name] {
			return fmt.Errorf("SSH-key verification: expected %q but got %q", trustedKey, ks)
		}

		return nil
	}
}

func parsePrivateKey(path string) ssh.Signer {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}
	return signer
}
