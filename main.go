package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ghodss/yaml"

	"golang.org/x/crypto/ssh"
)

var keypath, passphrase, user, host, command, hostfile, help string
var remotehosts []string

func init() {
	flag.StringVar(&keypath, "keypath", "/my/path/to/privatekey", "Specify the path for the private key")
	flag.StringVar(&passphrase, "keypass", "key-p@ssw0rd", "Specify the passphrase for the private key")
	flag.StringVar(&user, "user", "myuser", "Specify the remote user for the connection")
	flag.StringVar(&host, "host", "1.2.3.4", "Specify the remote host for the connection")
	flag.StringVar(&command, "cmd", "ls -l", "Specify the command to run on the remote host")
	flag.StringVar(&hostfile, "hostfile", "hostfile.yaml", "Specify the file which contains the ip addresses of the remote hosts")
	flag.StringVar(&help, "h", "", "Print help message (This message)")

	flag.Parse()
}

func main() {
	if (keypath == "") || (passphrase == "") || (user == "") || (host == "") || (command == "") || (help != "") {
		flag.PrintDefaults()
		os.Exit(1)
	}

	h := &remotehosts

	if hostfile != "" {
		*h = ProcessRemoteHostConfig(hostfile)
		for _, host := range remotehosts {
			RunRemoteCommand(keypath, passphrase, host, command)
		}
	} else {
		RunRemoteCommand(keypath, passphrase, host, command)
	}
}

func PubKeyAuth(keypath, passphrase string) ssh.AuthMethod {

	b, err := ioutil.ReadFile(keypath)
	Handle(err, "Failed to read private key file")

	key, err := ssh.ParsePrivateKeyWithPassphrase(b, []byte(passphrase))
	Handle(err, "Unable to parse private key with passphrase")

	return ssh.PublicKeys(key)

}

func ProcessRemoteHostConfig(hostfile string) []string {
	type IPS struct {
		Values []string `json:"ips"`
	}

	var ips IPS

	f, err := ioutil.ReadFile(hostfile)
	Handle(err, "Failed to open hostfile")

	yaml.Unmarshal(f, &ips)

	return ips.Values

}

func RunRemoteCommand(keypath, passphrase, host, command string) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			PubKeyAuth(keypath, passphrase),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host+":10222", config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		log.Fatal(err)
	}
	fmt.Println(b.String())
}

func Handle(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:\n", msg, err)
	}
}
