import sftp from "k6/x/sftp";
import exec from 'k6/execution';

const maxVUs = 2;
const host = __ENV.SFTP_HOST;
const port = __ENV.SFTP_PORT;
const user = __ENV.SFTP_USER;
const pemFile = __ENV.SFTP_PEMFILE;
const passphrase = __ENV.SFTP_PASSPHRASE;
const localDir = __ENV.LOCAL_DIR;
const filename = __ENV.FILENAME;
const remoteDir = __ENV.REMOTE_DIR

export const options = {
  vus: maxVUs,
  iterations: maxVUs,
};

export default function() {
  
  sftp.connect(host, port, user, pemFile, passphrase);
  
  sftp.upload(exec.vu.idInTest, localDir, filename, remoteDir);

  sftp.download(exec.vu.idInTest, remoteDir, filename, localDir);

  sftp.disconnect(exec.vu.idInTest);
}
