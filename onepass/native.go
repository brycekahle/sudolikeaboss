package onepass

import (
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

type nativeMessaging struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

type chromeNativeMessagingManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	AllowedOrigins []string `json:"allowed_origins"`
}

func NewNativeMessagingConnection() (OnePasswordConnection, error) {
	u, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "error getting current user")
	}

	mpath := path.Join(u.HomeDir, "/Library/Application Support/Google/Chrome/NativeMessagingHosts", nmHost+".json")
	log.Debugf("Reading manifest %s", mpath)
	manifestStr, err := ioutil.ReadFile(mpath)
	if err != nil {
		return nil, errors.Wrap(err, "error reading NM host manifest")
	}

	manifest := chromeNativeMessagingManifest{}
	err = json.Unmarshal(manifestStr, &manifest)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshaling NM host manifest")
	}

	nc := nativeMessaging{}

	cmd := exec.Command(manifest.Path, extensionID)
	nc.cmd = cmd
	nc.stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "error getting stdin pipe")
	}

	nc.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "error getting stdout pipe")
	}

	nc.stderr, err = cmd.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "error getting stderr pipe")
	}

	log.Debugf("Launching 1Password NM host at %s", manifest.Path)
	err = cmd.Start()
	if err != nil {
		return nil, errors.Wrap(err, "error starting NM host")
	}
	return &nc, nil
}

func (n *nativeMessaging) SendCommand(c *Command) error {
	return json.NewEncoder(n.stdin).Encode(c)
}

func (n *nativeMessaging) ReadResponse(r interface{}) error {
	return json.NewDecoder(n.stdout).Decode(r)
}

func (n *nativeMessaging) Close() error {
	log.Debugf("Closing NM client")
	d, err := ioutil.ReadAll(n.stdout)
	if err == nil && len(d) > 0 {
		log.Debugf("stdout\n%s\n", d)
	}

	d, err = ioutil.ReadAll(n.stderr)
	if err == nil && len(d) > 0 {
		log.Debugf("stderr\n%s\n", d)
	}
	err = n.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return n.cmd.Process.Kill()
	}
	return n.cmd.Wait()
}
