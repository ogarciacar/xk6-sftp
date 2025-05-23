# xk6-sftp
A k6 extension for load testing SFTP servers. This extension enables you to simulate multiple concurrent SFTP clients, performing file uploads and downloads under load. It's particularly useful for testing SFTP server performance, capacity, and reliability under various load conditions.

Key features:
- Concurrent SFTP client simulation
- File upload and download operations
- SSH key-based authentication
- Configurable connection parameters
- Detailed logging of operations

## Building k6 with xk6-sftp

To use this extension, you need to build a custom k6 binary that includes it. Use the following command:

```bash
xk6 build --with github.com/ogarciacar/xk6-sftp@latest
```

This will create a k6 binary in your current directory with the SFTP extension built-in.

## Security Setup

Before running any tests, you need to add your SFTP server's host key to your known_hosts file. This is required for secure connections.

For standard SSH port (22):
```bash
# Replace with your actual SFTP server hostname or IP
ssh-keyscan -H your-sftp-server >> ~/.ssh/known_hosts
```

For custom ports, include the port in the command:
```bash
# Replace with your actual SFTP server hostname/IP and port
ssh-keyscan -p PORT -H your-sftp-server >> ~/.ssh/known_hosts
```

For example:
```bash
ssh-keyscan -P 2222 -H sftp.example.com >> ~/.ssh/known_hosts
```

Note: When using IP:PORT combinations, the port becomes part of the host key entry.
For example:

```bash
ssh-keyscan -P 2222 -H 1.2.3.4 >> ~/.ssh/known_hosts
```
becomes the key:

`[1.2.3.4]:2222` and it is treated as a different host than `sftp.example.com:22`.

This step is necessary because the extension uses strict host key verification for security.

## Quick Start Guide

This guide explains how to use the provided example scripts to test SFTP operations. Each script demonstrates different testing scenarios.

### Environment Setup

Before running any script, set the following environment variables:

```bash
export SFTP_HOST="sftp.example.com"
export SFTP_PORT="22"
export SFTP_USER="username"
export SFTP_PEMFILE="$PWD/path/to/private/key.pem"
export SFTP_PASSPHRASE="your-passphrase"
export LOCAL_DIR="$PWD/path/to/local/dir"
export FILENAME="example.txt"
export REMOTE_DIR="/path/to/remote/dir"
```

When setting up environment variables refering to the local machine, it's important to use `$PWD` to reference the current working directory. This is required because:

1. The module performs path validation to ensure files are accessed within the allowed base path
2. The module uses `os.Getwd()` to determine the working directory and restricts file operations to that directory and its subdirectories
3. Without `$PWD`, the paths would be resolved relative to the user's home directory, which could fail the security checks in the code

Example:
```bash
export SFTP_PEMFILE="$PWD/path/to/private/key.pem"
export LOCAL_DIR="$PWD/path/to/local/dir"
```

### Example Scripts

#### 1. Single User Upload
`examples/single_sftp_user_upload.js` demonstrates basic file upload with a single sftp user shared across all virtual user:

```bash
./k6 run examples/single_sftp_user_upload.js
```

This script:
- Connects to the SFTP server
- Uploads a file
- Disconnects after completion

#### 2. Single User Download
`examples/single_sftp_user_download.js` shows how to download files with a single sftp user shared across all virtual user:

```bash
./k6 run examples/single_sftp_user_download.js
```

This script:
- Connects to the SFTP server
- Downloads a file
- Disconnects after completion

#### 3. Multiple Users Upload
`examples/multi_sftp_users_upload.js` demonstrates concurrent uploads with multiple sftp users one per each virtual user:

```bash
./k6 run examples/multi_sftp_users_upload.js
```

This script:
- Creates multiple SFTP connections
- Performs concurrent file uploads
- Manages connection lifecycle for each VU

#### 4. Connect During Iteration
`examples/connect_during_iteration.js` shows how to establish connections during test iterations:

```bash
./k6 run examples/connect_during_iteration.js
```

This script demonstrates:
- Dynamic connection management
- Connection reuse across iterations
- Proper connection cleanup

### Notes

- Ensure the `xk6-sftp` extension is installed and properly configured in your k6 setup
- Verify that the local and remote directories exist and have the necessary permissions
- Each script includes proper setup and teardown of SFTP connections
- The number of virtual users can be configured in the script options