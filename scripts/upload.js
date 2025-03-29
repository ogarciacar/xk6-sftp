import sftp from "k6/x/sftp";
import exec from 'k6/execution';

const max = 1

export const options = {
  iterations: max * 1,
  vus: max,
  setupTimeout: '360s', // Increase from 60s to 120s
};

export function setup() {
  sftp.connectVus(max, "x.x.x.x", "xx", "xx", "path/to/.pem", "passphrase");
}

export default function () {
  sftp.upload(exec.vu.idInTest, "./testdata/", "fileToUpload", "remoteUploadDir"); 
}

export function teardown() {
  sftp.disconnectVus()
}