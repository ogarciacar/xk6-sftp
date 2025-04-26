package sftp_test

import (
	"os"
	"testing"

	sftp "github.com/ogarciacar/xk6-sftp"

	"github.com/stretchr/testify/assert"
)

var (
	// SFTP server details
	host       = os.Getenv("SFTP_HOST")
	port       = os.Getenv("SFTP_PORT")
	user       = os.Getenv("SFTP_USER")
	pemFile    = os.Getenv("SFTP_PEMFILE")
	passphrase = os.Getenv("SFTP_PASSPHRASE")

	// file to upload/download
	localDir  = os.Getenv("LOCAL_DIR")
	filename  = os.Getenv("FILENAME")
	remoteDir = os.Getenv("REMOTE_DIR")
)

const succeed = "\u2713"
const failed = "\u2717"
const vuIdInTest = 1

var s *sftp.SFTP

func TestMain(m *testing.M) {

	// setup statements
	setup()

	// run the tests
	e := m.Run()

	// cleanup statements
	teardown()

	// report the exit code
	os.Exit(e)
}

func setup() {
	s = sftp.New()
	err := s.ConnectVus(vuIdInTest, host, port, user, pemFile, passphrase)
	if err != nil {
		panic(err)
	}
}

func teardown() {
	s.DisconnectVus()
}

func TestUpload(t *testing.T) {

	t.Parallel()

	t.Logf("Given SFTP server details: host=%s, port=%s, user=%s, pemFile=%s, passphrase=%s", host, port, user, pemFile, passphrase)
	{

		t.Logf("\tTest:\tWhen uploading filename=%q from localDir=%q to remoteDir=%q", filename, localDir, remoteDir)
		{
			err := s.Upload(vuIdInTest, localDir, filename, remoteDir)

			if err != nil {
				t.Errorf("\t%s:\tShould upload file, but got %v", failed, err)
			}

			t.Logf("\t%s:\tShould upload file", succeed)
		}
	}
}

func TestDownload(t *testing.T) {

	t.Parallel()

	t.Logf("Given SFTP server details: host=%s, port=%s, user=%s, pemFile=%s, passphrase=%s", host, port, user, pemFile, passphrase)
	{

		t.Logf("\tTest:\tWhen downloading filename=%q from remoteDir=%q to localDir=%q", filename, remoteDir, localDir)
		{
			err := s.Download(vuIdInTest, localDir, filename, remoteDir)

			if err != nil {
				t.Errorf("\t%s:\tShould download file, but got %v", failed, err)
			}

			t.Logf("\t%s:\tShould download file", succeed)

			assert.FileExistsf(t, localDir+filename, "%s\tShould exist after download %q", failed, localDir+filename)
		}
	}
}
