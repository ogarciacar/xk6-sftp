package sftp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"go.k6.io/k6/js/modules"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// This will allow operations in the current directory and its subdirectories
// while maintaining a secure fallback.
var allowedBasePath = func() string {
	wd, err := os.Getwd()
	if err != nil {
		return "/tmp" // fallback to /tmp if can't get working directory
	}
	return wd
}()

type SFTP struct {
	connectedVUs int
	clients      map[int]*SftpVu
	logger       *log.Logger
}

func New() *SFTP {
	l := log.New(log.Writer(), "xk6-sftp: ", log.Flags()|log.Lmsgprefix)
	l.Println("New")
	return &SFTP{
		connectedVUs: 0,
		clients:      make(map[int]*SftpVu),
		logger:       l,
	}
}

func init() {
	modules.Register("k6/x/sftp", New())
}

func (s *SFTP) Connect(host string, port string, user string, pemFile string, passphrase string) (*SftpVu, error) {

	cfg, err := config(user, pemFile, passphrase)

	if err != nil {
		return nil, fmt.Errorf("failed to configure SFTP connection: %v", err)
	}

	conn, err := ssh.Dial("tcp", host+":"+port, cfg)

	if err != nil {
		return nil, fmt.Errorf("failed to dial SFTP host: %v", err)
	}

	session, err := sftp.NewClient(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to start SFTP session: %v", err)
	}

	s.connectedVUs = s.connectedVUs + 1
	s.clients[s.connectedVUs] = &SftpVu{
		Id:      s.connectedVUs,
		Conn:    conn,
		Session: session,
	}

	s.logger.Printf("VU[%05d] Connect", s.connectedVUs)

	// A new ssh connection
	return s.clients[s.connectedVUs], nil
}

type SftpVu struct {
	Id      int
	Conn    *ssh.Client
	Session *sftp.Client
}

func (s *SFTP) Disconnect(key int) {
	s.logger.Printf("VU[%05d] Disconnect", key)
	err := s.clients[key].disconnect()
	if err != nil {
		s.logger.Printf("VU[%05d] Disconnect error: %v", key, err)
	}
	delete(s.clients, key)
}

func (v *SftpVu) disconnect() error {

	err := v.Session.Close()
	if err != nil {
		return fmt.Errorf("error closing SFTP session: %v", err)
	}

	err = v.Conn.Close()
	if err != nil {
		return fmt.Errorf("error disconnecting from SFTP host: %v", err)
	}

	return nil
}

func (s *SFTP) Upload(key int, localDir string, fileName string, remoteDir string) error {

	remoteFilePath := filepath.Join(remoteDir, fileName)
	localFilePath := filepath.Join(localDir, fileName)
	s.logger.Printf("VU[%05d] Upload %q to remote %q", key, localFilePath, remoteFilePath)
	return s.clients[key].upload(localDir, fileName, remoteDir)
}

func (v *SftpVu) upload(localDir string, fileName string, remoteDir string) error {

	remoteFilePath := filepath.Join(remoteDir, fileName)

	// Open the local file
	localFilePath := filepath.Join(localDir, fileName)

	safePath := filepath.Clean(localFilePath)

	if !strings.HasPrefix(safePath, allowedBasePath) {
		return fmt.Errorf("invalid path: %s", safePath)
	}

	localFile, err := os.Open(safePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// Create the remote file on the SFTP server
	remoteFile, err := v.Session.Create(remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %v", err)
	}
	defer remoteFile.Close()

	// Copy the local file to the remote file
	_, err = remoteFile.ReadFrom(localFile)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

func (s *SFTP) Download(key int, remoteDir string, fileName string, localDir string) error {
	remoteFilePath := filepath.Join(remoteDir, fileName)
	localFilePath := filepath.Join(localDir, fileName)
	log.Printf("VU[%05d] Download %q to local %q", key, remoteFilePath, localFilePath)
	return s.clients[key].download(remoteDir, fileName, localDir)
}

func (v *SftpVu) download(remoteDir string, fileName string, localDir string) error {

	remoteFilePath := filepath.Join(remoteDir, fileName)

	// Open the remote file on the SFTP server
	remoteFile, err := v.Session.Open(remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %v", err)
	}
	defer remoteFile.Close()

	// Define the local file path
	localFilePath := filepath.Join(localDir, fileName)

	safePath := filepath.Clean(localFilePath)

	if !strings.HasPrefix(safePath, allowedBasePath) {
		return fmt.Errorf("invalid path: %s", safePath)
	}

	// Create the local file
	localFile, err := os.Create(safePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	// Copy the remote file to the local file
	_, err = localFile.ReadFrom(remoteFile)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}

	return nil
}

func (s *SFTP) ConnectVus(numVus int, host, port, user, pemFile, passphrase string) error {
	for i := 0; i < numVus; i++ {
		_, err := s.Connect(host, port, user, pemFile, passphrase)
		if err != nil {
			return fmt.Errorf("failed to connect VU[%d]: %v", i+1, err)
		}
	}

	return nil
}

func (s *SFTP) DisconnectVus() {
	for key := range s.clients {
		s.Disconnect(key)
	}
}

func config(sftpUser, pemFile, passphrase string) (*ssh.ClientConfig, error) {
	safePath := filepath.Clean(pemFile)
	if !strings.HasPrefix(safePath, allowedBasePath) {
		return nil, fmt.Errorf("invalid path: %s", safePath)
	}

	// Read the PEM private key file
	pemBytes, err := os.ReadFile(safePath)
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}
	knownHostsPath := filepath.Join(homeDir, ".ssh", "known_hosts")

	knownHostsCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, fmt.Errorf("could not create hostkeycallback: %w", err)
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: sftpUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyCallback: knownHostsCallback,
	}

	return config, nil
}
