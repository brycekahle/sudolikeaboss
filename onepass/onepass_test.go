package onepass_test

import (
	"encoding/json"
	"os"

	. "github.com/brycekahle/sudolikeaboss/onepass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const SAMPLE_RESPONSE_0 = `
{
  "action": "fillItem",
  "payload": {
    "openInTabMode": "NewTab",
    "options": {
      "animate":true,
      "autosubmit":true
    }, 
    "item": {
      "uuid":"someuuid",
      "nakedDomains": ["somedomain.com"],
      "overview": {
        "title": "title",
        "url": "url"
      },
      "secureContents": {
        "htmlForm": {"htmlMethod":"post"},
        "fields": [
          {
            "value":"username",
            "id":"email",
            "name":"email",
            "type":"T",
            "designation":"username"
          },
          {
            "value":"password",
            "id":"password",
            "name":"password",
            "type":"P",
            "designation":"password"
          },
          {
            "value":"Login",
            "id":"",
            "name":"",
            "type":"I"
          }
        ]
      }
    },
    "action":"fillLogin"
  },
  "version":"01"
}
`

type MockWebsocketClient struct {
	responseString string
}

func (mock MockWebsocketClient) SendCommand(cmd *Command) (*Response, error) {
	r := Response{}
	err := json.Unmarshal([]byte(mock.responseString), &r)
	return &r, err
}

func (mock MockWebsocketClient) Close() error {
	return nil
}

var _ = Describe("Sudolikeaboss", func() {
	Describe("Response", func() {
		var (
			response *Response
			err      error
		)

		BeforeEach(func() {
			err = json.Unmarshal([]byte(SAMPLE_RESPONSE_0), response)
			if err != nil {
				panic(err)
			}
		})

		It("should do something", func() {
			password, err := response.GetPassword()
			Expect(password).To(Equal("password"))
			Expect(err).To(BeNil())
		})
	})

	Describe("Client", func() {
		var (
			client              *OnePasswordClient
			mockWebsocketClient *MockWebsocketClient
			err                 error
		)

		BeforeEach(func() {
			mockWebsocketClient = &MockWebsocketClient{}
			client, err = NewCustomClient(mockWebsocketClient, "fakehost", os.TempDir())
			Expect(err).To(BeNil())
		})

		It("should send hello command to 1password", func() {
			mockWebsocketClient.responseString = `{"action":"authBegin"}`

			response, err := client.SendHelloCommand()

			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
		})

		XIt("should send showPopup command to 1password", func() {
			mockWebsocketClient.responseString = SAMPLE_RESPONSE_0

			//response, err := client.SendShowPopupCommand()

			// Expect(err).To(BeNil())
			// Expect(response.GetPassword()).To(Equal("password"))
		})
	})
})
