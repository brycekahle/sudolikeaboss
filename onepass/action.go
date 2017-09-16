package onepass

import (
	"encoding/json"
)

type SendAction Action

const SendHello SendAction = "hello"
const SendAuthNew SendAction = "authNew"
const SendAuthBegin SendAction = "authBegin"
const SendAuthRegister SendAction = "authRegister"
const SendAuthVerify SendAction = "authVerify"
const SendAuthFail SendAction = "authFail"
const ShowPopup SendAction = "showPopup"

type ResponseAction Action

const ResponseWelcome ResponseAction = "welcome"
const ResponseAuthNew ResponseAction = "authNew"
const ResponseAuthBegin ResponseAction = "authBegin"
const ResponseAuthRegistered ResponseAction = "authRegistered"
const ResponseAuthContinue ResponseAction = "authContinue"
const ResponseAuthFail ResponseAction = "authFail"
const FillItem ResponseAction = "fillItem"

type PayloadAction Action

const FillLogin PayloadAction = "fillLogin"
const FillPassword PayloadAction = "fillPassword"

// unused actions, but found in 1Password extension JS

const ResetConnection ResponseAction = "resetConnection"
const HangUp ResponseAction = "hangUp"

type Action string

func (a Action) String() string {
	return string(a)
}

func (a *Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Action) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*a = Action(s)
	return nil
}
