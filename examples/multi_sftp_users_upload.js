import sftp from "k6/x/sftp";
import exec from 'k6/execution';
import { SharedArray } from 'k6/data';

const maxVUs = 2;
const host = __ENV.SFTP_HOST;
const port = __ENV.SFTP_PORT;
const user = __ENV.SFTP_USER;
const pemFile = __ENV.SFTP_PEMFILE;
const passphrase = __ENV.SFTP_PASSPHRASE;
const localDir = __ENV.LOCAL_DIR;
const filename = __ENV.FILENAME;
const remoteDir = __ENV.REMOTE_DIR

const data = new SharedArray('sftpVUs', function () {
  const dataArray = [];
  
  for (let i = 0; i < maxVUs; i++) {
    dataArray.push({
      host: host,
      port: port,
      user: user,
      pemFile: pemFile,
      passphrase: passphrase  
    });
  }
  return dataArray; // must be an array
});

export const options = {
  vus: maxVUs,
  iterations: maxVUs,
};

export function setup() {
  for (let i = 0; i < maxVUs; i++) {
    sftp.connect(data[i].host, data[i].port, data[i].user, data[i].pemFile, data[i].passphrase);
  }
}

export default function() {
  sftp.upload(exec.vu.idInTest, localDir, filename, remoteDir);
}

export function teardown() {
  sftp.disconnectVus();
}