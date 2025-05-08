package sftp_test

import (
	"os"
	"path"
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

func TestConnect(t *testing.T) {
	t.Log("Given the need for verifying the correctness of xk6-sftp connect")
	{
		t.Logf("\tTest 0:\tWhen connecting to host=%q on port=%q\n\t\t\tusing credentials user=%q, pemFile=%q and passphrase=%q", host, port, user, pemFile, passphrase)
		{
			ext := sftp.New()
			_, err := ext.Connect(host, port, user, pemFile, passphrase)
			if err != nil {
				t.Errorf("\t%s:\tShould establish connection without errors, but got: %v", failed, err)
			} else {
				t.Logf("\t%s:\tShould establish connection without errors", succeed)
				defer ext.Disconnect(1)
			}
		}
	}
}

func TestUpload(t *testing.T) {

	t.Log("Given the need for verifying the correctness of xk6-sftp upload")
	{
		t.Logf("\tTest 1:\tWhen uploading the file=%q from localDir=%q to remoteDir=%q\n\t\t\ton host=%q and port=%q with user=%q, pemFile=%q and passphrase=%q", filename, localDir, remoteDir, host, port, user, pemFile, passphrase)
		{
			ext := sftp.New()

			_, err := ext.Connect(host, port, user, pemFile, passphrase)

			if err != nil {
				t.Fatalf("\t%s:\tShould establish connection without errors, but got: %v", failed, err)
			}

			defer ext.Disconnect(1)

			if err := ext.Upload(1, localDir, filename, remoteDir); err != nil {
				t.Errorf("\t%s:\tShould complete without errors, but got: %v", failed, err)
			} else {
				t.Logf("\t%s:\tShould complete without errors", succeed)
			}
		}
	}
}

func TestDownload(t *testing.T) {

	t.Log("Given the need for verifying the correctness of xk6-sftp download")
	{
		ext := sftp.New()

		_, err := ext.Connect(host, port, user, pemFile, passphrase)

		if err != nil {
			t.Fatalf("\t%s:\tShould establish connection without errors, but got: %v", failed, err)
		}

		defer ext.Disconnect(1)

		if err := ext.Upload(1, localDir, filename, remoteDir); err != nil {
			t.Fatalf("\t%s:\tShould upload without errors, but got: %v", failed, err)
		}

		t.Logf("\tTest 2:\tWhen downloading the file=%q from remoteDir=%q to the localDir=%q\n\t\t\ton host=%q and port=%q with user=%q, pemFile=%q and passphrase=%q", filename, remoteDir, localDir, host, port, user, pemFile, passphrase)
		{
			tmpLocalDir := t.TempDir()
			if err := ext.Download(1, remoteDir, filename, tmpLocalDir); err != nil {
				t.Errorf("\t%s:\tShould complete without errors, but got: %v", failed, err)
			} else {
				t.Logf("\t%s:\tShould complete without errors", succeed)
				assert.FileExistsf(t, path.Join(tmpLocalDir, filename), "%s\tShould exist after download %q", failed, localDir+filename)
			}
		}
	}
}

func TestConnectVus(t *testing.T) {
	t.Log("Given the need for verifying the correctness of xk6-sftp connect clients")
	{
		count := 3
		t.Logf("\tTest 4:\tWhen connecting %d clients to host=%q on port=%q\n\t\t\tusing credentials user=%q, pemFile=%q and passphrase=%q", count, host, port, user, pemFile, passphrase)
		{
			ext := sftp.New()
			err := ext.ConnectVus(count, host, port, user, pemFile, passphrase)
			if err != nil {
				t.Errorf("\t%s:\tShould establish connection without errors, but got: %v", failed, err)
			} else {
				t.Logf("\t%s:\tShould establish connection without errors", succeed)
				defer ext.DisconnectVus()
			}
		}
	}
}
