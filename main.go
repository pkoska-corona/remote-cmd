package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/ghodss/yaml"

	"golang.org/x/crypto/ssh"
)

var keypath, passphrase, user, host, command, hostfile string
var remotehosts []string

func init() {
	flag.StringVar(&keypath, "keypath", "/my/path/to/privatekey", "Specify the path for the private key")
	flag.StringVar(&passphrase, "keypass", "key-p@ssw0rd", "Specify the passphrase for the private key")
	flag.StringVar(&user, "user", "myuser", "Specify the remote user for the connection")
	flag.StringVar(&host, "host", "1.2.3.4", "Specify the remote host for the connection")
	flag.StringVar(&command, "cmd", "ls -l", "Specify the command to run on the remote host")
	flag.StringVar(&hostfile, "hostfile", "hostfile.yaml", "Specify the file which contains the ip addresses of the remote hosts")

	flag.Parse()
}

func main() {
	h := &remotehosts

	if hostfile != "" {
		*h = ProcessRemoteIPs(hostfile)
		for _, host := range remotehosts {
			RunRemoteCommand(keypath, passphrase, host, command)
		}
	} else {
		RunRemoteCommand(keypath, passphrase, host, command)
	}
}

func PubKeyAuth(keypath string, passphrase string) ssh.AuthMethod {

	b, err := ioutil.ReadFile(keypath)
	Handle(err, "Failed to read private key file")

	key, err := ssh.ParsePrivateKeyWithPassphrase(b, []byte(passphrase))
	Handle(err, "Unable to parse private key with passphrase")

	return ssh.PublicKeys(key)

}

func ProcessRemoteIPs(hostfile string) []string {
	type IPS struct {
		Values []string `json:"ips"`
	}

	var ips IPS

	f, err := ioutil.ReadFile(hostfile)
	Handle(err, "Failed to open hostfile")

	yaml.Unmarshal(f, &ips)

	return ips.Values

}

func RunRemoteCommand(keypath string, passphrase string, host string, command string) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			PubKeyAuth(keypath, passphrase),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host+":10222", config)
	if err != nil {
		Handle(err, "unable to connect")
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		Handle(err, "Failed to create session to remote host")
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		Handle(err, "Failed to run command")
	}
	fmt.Println(b.String())
}

func Handle(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:\n", msg, err)
	}
}