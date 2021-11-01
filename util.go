package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

func addTrustedHost(name string, k ssh.PublicKey) error {
	log.Printf("WARNING: SSH-key verification is *NOT* in effect: to fix, add this trustedKey: %q", keyString(k))
	fmt.Println("Do you want to add the key to the trusted list? [yes]/no")
	var answer string
	fmt.Scanln(&answer)
	if answer == "" || answer == "yes" {
		homeDir, _ := os.UserHomeDir()
		confDir := filepath.Join(homeDir, ".config", "deployment")
		confPath := filepath.Join(confDir, "trusted_hosts.txt")
		os.MkdirAll(confDir, 0700)
		os.WriteFile(confPath, k.Marshal(), 0600)
		f, err := os.OpenFile(confPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		f.WriteString(name + "::")
		f.WriteString(keyString(k))
		f.WriteString("\n")
	} else {
		fmt.Println("You will be asked again when you connect next time.")
	}
	return nil
}

func readTrustedHost() (map[string]string, error) {
	homeDir, _ := os.UserHomeDir()
	confPath := filepath.Join(homeDir, ".config", "deployment", "trusted_hosts.txt")
	f, err := os.Open(confPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hosts := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		item := strings.TrimSpace(scanner.Text())
		items := strings.Split(item, "::")
		n, k := items[0], items[1]
		hosts[n] = k
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return hosts, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
