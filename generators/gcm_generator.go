package generators

import (
	"encoding/json"

	"github.com/topfreegames/pusher-tester/constants"
)

type KafkaGCMMessage struct {
	To           string                 `json:"to"`
	Notification interface{}            `json:"notification"`
	DryRun       bool                   `json:"dry_run"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	PushExpiry   int64                  `json:"push_expiry,omitempty"`
}

type GCMMessageGenerator struct {
}

func (g *GCMMessageGenerator) Generate() []byte {
	to, err := idGenerator(152)
	if err != nil {
		panic("error generating device id")
	}

	msg := KafkaGCMMessage{
		To:           to,
		Notification: `{"title": "Come play!", "body": "Helena miss you! come play!"}`,
		DryRun:       true,
		Metadata: map[string]interface{}{
			"jobId": "77372c1e-c124-4552-b77a-f4775bbad850",
		},
		PushExpiry: 0,
	}

	j, err := json.Marshal(msg)
	if err != nil {
		panic("error marshelling message")
	}

	return j
}

func (g *GCMMessageGenerator) Platform() string {
	return constants.GCMPlatform
}
