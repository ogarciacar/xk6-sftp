package sftp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"go.k6.io/k6/js/modules"
	"golang.org/x/crypto/ssh"
)

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
	s.clients[key].disconnect()
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
	localFile, err := os.Open(localFilePath)
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

	// Create the local file
	localFile, err := os.Create(localFilePath)
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

	//log.Printf("SFTP[0]: configuring connection for user %q\n", sftpUser)

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

// func connect(vuIdInTest int, config *ssh.ClientConfig, host string, port string) (*ssh.Client, error) {

// 	log.Printf("VU[%03d]: connecting to %q on port %q\n", vuIdInTest+1, host, port)

// 	// Connect to the SSH server
// 	conn, err := ssh.Dial("tcp", host+":"+port, config)
// 	if err != nil {
// 		e := fmt.Errorf("failed to dial VU[%03d]: %v", vuIdInTest+1, err)
// 		log.Println(e)
// 		return nil, e
// 	}

// 	// A new ssh connection
// 	return conn, nil
// }

// func (s *SFTP) DisconnectVus() error {

// 	for i := 0; i < len(s.clients); i++ {
// 		err := s.clients[i].session.Close()
// 		if err != nil {
// 			e := fmt.Errorf("error closing session VU[%03d]: %v", i+1, err)
// 			log.Println(e)
// 			return e
// 		}

// 		err = s.clients[i].conn.Close()
// 		if err != nil {
// 			e := fmt.Errorf("error closing connection VU[%03d]: %v", i+1, err)
// 			log.Println(e)
// 			return e
// 		}

// 		log.Printf("VU[%03d]: disconnected from %q\n", i+1, s.clients[i].conn.RemoteAddr().String())
// 	}

// 	log.Printf("SFTP[2]: disconnected %d VUs\n", len(s.clients))

// 	return nil
// }

// func (s *SFTP) Upload(vuIdInTest int, localDir string, fileName string, remoteDir string) error {

// 	if vuIdInTest < 1 || vuIdInTest > len(s.clients) {
// 		e := fmt.Errorf("failed to upload file %s: vuIdInTest=%d does not exist", fileName, vuIdInTest)
// 		log.Println(e)
// 		return e
// 	}

// 	//log.Printf("VU[%03d]: uploading file %q\n", vuIdInTest, fileName)

// 	// Define the remote directory
// 	remoteFilePath := filepath.Join(remoteDir, fileName)

// 	// Open the local file
// 	localFilePath := filepath.Join(localDir, fileName)
// 	localFile, err := os.Open(localFilePath)
// 	if err != nil {
// 		e := fmt.Errorf("failed to open local file VU[%03d] %v", vuIdInTest, err)
// 		log.Println(e)
// 		return e
// 	}
// 	defer localFile.Close()

// 	// Create the remote file on the SFTP server
// 	remoteFile, err := s.clients[vuIdInTest-1].session.Create(remoteFilePath)
// 	if err != nil {
// 		e := fmt.Errorf("failed to create remote file VU[%03d]: %v", vuIdInTest, err)
// 		log.Println(e)
// 		return e
// 	}
// 	defer remoteFile.Close()

// 	// Copy the local file to the remote file
// 	_, err = remoteFile.ReadFrom(localFile)
// 	if err != nil {
// 		e := fmt.Errorf("failed to upload file VU[%03d]: %v", vuIdInTest, err)
// 		log.Println(e)
// 		return e
// 	}

// 	return nil
// }

// func (s *SFTP) Download(vuIdInTest int, localDir string, filename string, remoteDir string) error {

// 	if vuIdInTest < 1 || vuIdInTest > len(s.clients) {
// 		e := fmt.Errorf("vuIdInTest=%d does not exist", vuIdInTest)
// 		log.Println(e)
// 		return e
// 	}

// 	remoteFilePath := filepath.Join(remoteDir, filename)
// 	log.Printf("VU[%03d]: downloading file %q\n", vuIdInTest, remoteFilePath)

// 	// Open the remote file on the SFTP server
// 	remoteFile, err := s.clients[vuIdInTest-1].session.Open(remoteFilePath)
// 	if err != nil {
// 		e := fmt.Errorf("failed to open remote file VU[%03d]: %v", vuIdInTest, err)
// 		log.Println(e)
// 		return e
// 	}
// 	defer remoteFile.Close()

// 	// Define the local file path
// 	localFilePath := filepath.Join(localDir, filename)

// 	// Create the local file
// 	localFile, err := os.Create(localFilePath)
// 	if err != nil {
// 		e := fmt.Errorf("failed to create local file VU[%03d]: %v", vuIdInTest, err)
// 		log.Println(e)
// 		return e
// 	}
// 	defer localFile.Close()

// 	// Copy the remote file to the local file
// 	_, err = localFile.ReadFrom(remoteFile)
// 	if err != nil {
// 		e := fmt.Errorf("failed to download file VU[%03d]: %v", vuIdInTest, err)
// 		log.Println(e)
// 		return e
// 	}

// 	return nil
// }
