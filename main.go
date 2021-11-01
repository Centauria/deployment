package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/akamensky/argparse"
	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	parser := argparse.NewParser("deployment", "Deploy project folder to another node via SSH")
	projectPath := parser.String("p", "project", &argparse.Options{
		Required: false,
		Default:  ".",
		Help:     `Project path`,
	})
	privateKeyPath := parser.String("k", "private-key", &argparse.Options{
		Required: false,
		Default:  filepath.Join(homeDir, ".ssh/id_rsa"),
		Help:     `Path to private key file`,
	})
	verbose := parser.Flag("v", "verbose", &argparse.Options{
		Required: false,
		Default:  false,
		Help:     `Print more info when working`,
	})
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		return
	}
	var trustedHosts map[string]string
	trustedHosts, err = readTrustedHost()
	if err != nil {
		fmt.Println("Error while reading trusted hosts")
		fmt.Println(err)
	}
	deploymentFilePath := filepath.Join(*projectPath, "deployment.txt")
	if _, err := os.Stat(deploymentFilePath); err == nil {
		dep, _ := ioutil.ReadFile(deploymentFilePath)
		deploymentPath := strings.TrimSpace(string(dep))
		conf := strings.Split(deploymentPath, ":")
		var ip, remotePath string
		var port int = 22
		if len(conf) == 2 {
			ip, remotePath = conf[0], conf[1]
		} else if len(conf) == 3 {
			ip, remotePath = conf[0], conf[2]
			port, _ = strconv.Atoi(conf[1])
		}
		currentUser, _ := user.Current()
		var sshClient *ssh.Client
		signer := parsePrivateKey(*privateKeyPath)
		sshConfig := &ssh.ClientConfig{
			User:            currentUser.Username,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: trustedHostKeyCallback(trustedHosts),
		}
		sshClient, err := ssh.Dial("tcp", ip+":"+strconv.Itoa(port), sshConfig)
		if err != nil {
			fmt.Println("Couldn't establish ssh connection to the remote server")
			fmt.Println(err)
			return
		}
		defer sshClient.Close()
		sftpClient, err := sftp.NewClient(sshClient)
		if err != nil {
			log.Fatal(err)
			return
		}
		defer sftpClient.Close()
		err = filepath.WalkDir(*projectPath,
			func(path string, info fs.DirEntry, err error) error {
				scpClient, err := scp.NewClientBySSH(sshClient)
				if err != nil {
					fmt.Println("Couldn't establish scp connection to the remote server")
					fmt.Println(err)
					return err
				}
				defer scpClient.Close()

				if path == deploymentFilePath {
					return nil
				}
				var dest string
				if *projectPath == "." {
					dest = filepath.Join(remotePath, path)
				} else {
					dest = strings.Replace(path, *projectPath, remotePath, 1)
				}
				if *verbose {
					fmt.Println(path, "->", dest)
				}
				if info.IsDir() {
					if _, err := sftpClient.Lstat(dest); err != nil {
						sftpClient.MkdirAll(dest)
					}
				} else {
					f, err := os.Open(path)
					if err != nil {
						fmt.Println("Error while opening file ", path)
						fmt.Println(err)
						return err
					}
					err = scpClient.CopyFile(f, dest, "0655")
					if err != nil {
						fmt.Println("Error while copying file", dest)
						fmt.Println(err)
						return err
					}
				}
				return nil
			})
	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println("Couldn't find private key file")
		fmt.Println(err)
		return
	}
}
