package sftp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/sftp"
	"go.k6.io/k6/js/modules"
	"golang.org/x/crypto/ssh"
)

type SFTP struct {
	clients []*client
}

type client struct {
	id      int
	conn    *ssh.Client
	session *sftp.Client
	err     error
}

func New() *SFTP {
	return &SFTP{}
}

func init() {
	modules.Register("k6/x/sftp", New())
}

func (s *SFTP) ConnectVus(numVus int, host, port, user, pemFile, passphrase string) error {

	s.clients = make([]*client, numVus)
	cfg, err := config(user, pemFile, passphrase)

	t1 := time.Now()
	if err != nil {
		e := fmt.Errorf("failed to configure sftp conneciton: %v", err)
		log.Println(e)
		return e
	}

	ch := make(chan *client)
	// fan out
	for id := 0; id < numVus; id++ {
		go func() {
			c := client{
				id: id,
			}
			c.conn, c.err = connect(id, cfg, host, port)
			ch <- &c
		}()
	}

	// collect
	for id := 0; id < numVus; id++ {
		c := <-ch
		if c.err != nil {
			e := fmt.Errorf("failed to connect VU[%03d]: %v", c.id+1, c.err)
			log.Println(e)
			return e
		}

		// // Start an SFTP session
		c.session, c.err = sftp.NewClient(c.conn)
		if c.err != nil {
			e := fmt.Errorf("failed to start SFTP session: %v", c.err)
			log.Println(e)
			return e
		}
		s.clients[c.id] = c
		log.Printf("VU[%03d]: sftp session started\n", c.id+1)
	}

	log.Printf("SFTP[0]: connectVus took %v", time.Since(t1))

	return nil
}

func config(sftpUser, pemFile, passphrase string) (*ssh.ClientConfig, error) {

	log.Printf("SFTP[0]: configuring connection for user `%s`\n", sftpUser)

	// Read the PEM private key file
	pemBytes, err := os.ReadFile(pemFile)
	if err != nil {
		e := fmt.Errorf("failed to read PEM file: %v", err)
		log.Println(e)
		return nil, e
	}

	// Parse the PEM file with the passphrase
	privateKey, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(passphrase))
	if err != nil {
		e := fmt.Errorf("failed to parse private key: %v", err)
		log.Println(e)
		return nil, e
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}

func connect(vuIdInTest int, config *ssh.ClientConfig, host string, port string) (*ssh.Client, error) {

	log.Printf("VU[%03d]: connecting to %s:%s\n", vuIdInTest+1, host, port)

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		e := fmt.Errorf("failed to dial VU[%03d]: %v", vuIdInTest+1, err)
		log.Println(e)
		return nil, e
	}

	// A new ssh connection
	return conn, nil
}

func (s *SFTP) DisconnectVus() error {

	for i := 0; i < len(s.clients); i++ {
		err := s.clients[i].session.Close()
		if err != nil {
			e := fmt.Errorf("error closing session VU[%03d]: %v", i+1, err)
			log.Println(e)
			return e
		}

		err = s.clients[i].conn.Close()
		if err != nil {
			e := fmt.Errorf("error closing connection VU[%03d]: %v", i+1, err)
			log.Println(e)
			return e
		}

		log.Printf("VU[%03d]: disconnected from SFTP server\n", i+1)
	}

	return nil
}

func (s *SFTP) Upload(vuIdInTest int, localDir string, fileName string, remoteDir string) error {

	if vuIdInTest < 1 || vuIdInTest > len(s.clients) {
		e := fmt.Errorf("failed to upload file %s: vuIdInTest=%d does not exist", fileName, vuIdInTest)
		log.Println(e)
		return nil
	}

	log.Printf("VU[%03d]: uploading file %s\n", vuIdInTest, fileName)

	// Define the remote directory
	remoteFilePath := filepath.Join(remoteDir, fileName)

	// Open the local file
	localFilePath := filepath.Join(localDir, fileName)
	localFile, err := os.Open(localFilePath)
	if err != nil {
		e := fmt.Errorf("failed to open local file VU[%03d] %v", vuIdInTest, err)
		log.Println(e)
		return e
	}
	defer localFile.Close()

	// Create the remote file on the SFTP server
	remoteFile, err := s.clients[vuIdInTest-1].session.Create(remoteFilePath)
	if err != nil {
		e := fmt.Errorf("failed to create remote file VU[%03d]: %v", vuIdInTest, err)
		log.Println(e)
		return e
	}
	defer remoteFile.Close()

	// Copy the local file to the remote file
	_, err = remoteFile.ReadFrom(localFile)
	if err != nil {
		e := fmt.Errorf("failed to upload file VU[%03d]: %v", vuIdInTest, err)
		log.Println(e)
		return e
	}

	return nil
}
