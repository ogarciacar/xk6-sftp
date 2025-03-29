package sftp_test

import (
	"testing"

	sftp "github.com/ogarciacar/xk6-sftp"

	"github.com/stretchr/testify/require"
)

func TestConnectClients(t *testing.T) {

	// SFTP server details
	host := "xxx"
	port := "xxx"
	user := "xxx"
	pemFile := "xxx"
	passphrase := "xxx"

	// file to upload
	localDir := "./"
	filename := "LICENSE"
	remoteDir := "uploadDir"

	vuIdInTest := 1
	s := sftp.New()
	err := s.ConnectVus(vuIdInTest, host, port, user, pemFile, passphrase)
	require.NoError(t, err, "ConnectVus should not return an error")

	err = s.Upload(vuIdInTest, localDir, filename, remoteDir)
	require.NoError(t, err, "Upload should not return an error")

	err = s.DisconnectVus()
	require.NoError(t, err, "DisconnectVus should not return an error")

}
