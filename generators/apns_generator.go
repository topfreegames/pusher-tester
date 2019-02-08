package generators

import (
	"encoding/json"

	"github.com/topfreegames/pusher-tester/constants"
)

type APNSNotification struct {
	DeviceToken string
	Payload     interface{}
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	PushExpiry  int64                  `json:"push_expiry"`
}

type APSNMessageGenerator struct {
}

func (a *APSNMessageGenerator) Generate() []byte {
	deviceToken, err := idGenerator(64)
	if err != nil {
		panic("error generating device id")
	}

	msg := APNSNotification{
		DeviceToken: deviceToken,
		Payload:     `{"aps":{"alert":"Helena miss you! come play!"}}`,
		Metadata: map[string]interface{}{
			"jobId": "86edb3c3-6b5e-40dc-9f14-4ba831daf87c",
		},
		PushExpiry: 0,
	}

	j, err := json.Marshal(msg)
	if err != nil {
		panic("error marshelling message")
	}

	return j
}

func (a *APSNMessageGenerator) Platform() string {
	return constants.APNSPlatform
}
