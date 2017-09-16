package onepass

const clientVersion = "4.6.2.90"

type Command struct {
	Action   SendAction `json:"action"`
	Version  string     `json:"version,omitempty"`
	BundleID string     `json:"bundleId,omitempty"`
	Payload  Payload    `json:"payload"`
}

type Payload struct {
	Version      string            `json:"version,omitempty"`
	ExtID        string            `json:"extId,omitempty"`
	Method       string            `json:"method,omitempty"`
	Secret       string            `json:"secret,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
	URL          string            `json:"url,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
	CC           string            `json:"cc,omitempty"`
	CS           string            `json:"cs,omitempty"`
	M4           string            `json:"M4,omitempty"`
	M3           string            `json:"M3,omitempty"`
	Algorithm    string            `json:"alg,omitempty"`
	Iv           string            `json:"iv,omitempty"`
	Data         string            `json:"data,omitempty"`
	Hmac         string            `json:"hmac,omitempty"`
}

func NewCommand(action SendAction, payload Payload) *Command {
	command := Command{
		Action:  action,
		Version: clientVersion,
		Payload: payload,
	}
	return &command
}
