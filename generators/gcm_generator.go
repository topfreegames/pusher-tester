package generators

import (
	"encoding/json"

	"github.com/topfreegames/pusher-tester/constants"
)

type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type KafkaGCMMessage struct {
	To           string                 `json:"to"`
	Notification Notification           `json:"notification"`
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
		To: to,
		Notification: Notification{
			Title: "Come play!",
			Body:  "Helena miss you! come play!",
		},
		DryRun: true,
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
