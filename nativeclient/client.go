package nativeclient

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const nmHost = "2bua8c4s2c.com.agilebits.1password"
const extensionID = "chrome-extension://aomjjhallfgjeglblehebfpbcfeobpgk/"

type chromeNativeMessagingManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	AllowedOrigins []string `json:"allowed_origins"`
}

type NativeClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func NewClient() *NativeClient {
	return &NativeClient{}
}

func (nc *NativeClient) Connect() error {
	u, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "error getting current user")
	}

	mpath := path.Join(u.HomeDir, "/Library/Application Support/Google/Chrome/NativeMessagingHosts", nmHost+".json")
	log.Debugf("Reading manifest %s", mpath)
	manifestStr, err := ioutil.ReadFile(mpath)
	if err != nil {
		return errors.Wrap(err, "error reading NM host manifest")
	}

	manifest := chromeNativeMessagingManifest{}
	err = json.Unmarshal(manifestStr, &manifest)
	if err != nil {
		return errors.Wrap(err, "error unmarshaling NM host manifest")
	}

	cmd := exec.Command(manifest.Path, extensionID)
	nc.cmd = cmd
	nc.stdin, err = cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "error getting stdin pipe")
	}

	nc.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "error getting stdout pipe")
	}

	nc.stderr, err = cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "error getting stderr pipe")
	}

	log.Debugf("Launching 1Password NM host at %s", manifest.Path)
	err = cmd.Start()
	return errors.Wrap(err, "error starting NM host")
}

func (nc *NativeClient) Close() {
	log.Debugf("Closing NM client")
	d, err := ioutil.ReadAll(nc.stdout)
	if err == nil && len(d) > 0 {
		log.Debugf("stdout\n%s\n", d)
	}

	d, err = ioutil.ReadAll(nc.stderr)
	if err == nil && len(d) > 0 {
		log.Debugf("stderr\n%s\n", d)
	}
	nc.cmd.Process.Signal(os.Interrupt)
	nc.cmd.Wait()
}

func (nc *NativeClient) Send(v string) error {
	dlen := len(v)
	log.Debugf("Sending %d: %s", dlen, string(v))
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(dlen))
	_, err := nc.stdin.Write(b)
	if err != nil {
		return errors.Wrap(err, "error writing sent message length")
	}
	_, err = nc.stdin.Write([]byte(v))
	return errors.Wrap(err, "error writing message")
}

func (nc *NativeClient) Receive() (string, error) {
	b := make([]byte, 4)
	_, err := nc.stdout.Read(b)
	if err != nil {
		return "", errors.Wrap(err, "error reading received message length")
	}

	msglen := binary.LittleEndian.Uint32(b)
	log.Debugf("Reading %d bytes", msglen)
	data := make([]byte, msglen)
	_, err = nc.stdout.Read(data)
	if err != nil {
		return "", errors.Wrap(err, "error reading message")
	}

	log.Debugf("Read %s", string(data))
	return string(data), nil
}
