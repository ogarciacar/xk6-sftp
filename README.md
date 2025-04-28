# xk6-sftp
xk6 extension for load testing SFTP servers

## Quick Start Guide: Uploading and Downloading Files

This guide explains how to use the `xk6-sftp` extension to upload and download files using environment variables.

### 1. Set Up Environment Variables
Before running the script, set the following environment variables to configure the SFTP connection and file paths:

- `SFTP_HOST`: The hostname or IP address of the SFTP server.
- `SFTP_PORT`: The port number for the SFTP server (default is usually 22).
- `SFTP_USER`: The username for authentication.
- `SFTP_PEMFILE`: Path to the private key file for authentication.
- `SFTP_PASSPHRASE`: Passphrase for the private key (if applicable).
- `LOCAL_DIR`: Local directory where files will be downloaded or uploaded.
- `FILENAME`: Name of the file to download or upload.
- `REMOTE_DIR`: Remote directory on the SFTP server.

Example:
```shell
export SFTP_HOST="example.com"
export SFTP_PORT="22"
export SFTP_USER="username"
export SFTP_PEMFILE="/path/to/private/key.pem"
export SFTP_PASSPHRASE="your-passphrase"
export LOCAL_DIR="/path/to/local/dir"
export FILENAME="example.txt"
export REMOTE_DIR="/path/to/remote/dir"
```
The `setup()` function in the scripts ensures all SFTP connections are stablished before execution:

- There will be one client connection per VU defined in the script.
- Ensure the number of iterations is equal or greather than the number of VUs.

### 2. Download Files
The `download.js` script is pre-configured to download a file from the SFTP server. It uses the `sftp.download()` function to download the file specified by `FILENAME` from `REMOTE_DIR` to `LOCAL_DIR`.

Run the script:
```shell
./k6 run scripts/download.js
```

### 3. Upload Files
The `upload.js` script is pre-configured to upload a file from the local host to the SFTP server. It uses the `sftp.upload()` function to upload the file specified by `FILENAME` from `LOCAL_DIR` to `REMOTE_DIR`.

Run the script:
```shell
./k6 run scripts/upload.js
```

### 4. Teardown
The `teardown()` function in the scripts ensures all SFTP connections are closed after execution.

**Notes**

- Ensure the `xk6-sftp` extension is installed and properly configured in your k6 setup.
- Verify that the local and remote directories exist and have the necessary permissions.
- This guide should help you quickly get started with uploading and downloading files using the `xk6-sftp` extension.