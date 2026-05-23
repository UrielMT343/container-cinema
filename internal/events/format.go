package events

import (
	"encoding/json"
	"fmt"
)

type SeatEventPayload struct {
	IDSeat int    `json:"seatId"`
	Status string `json:"status"`
}

func FormatSSE(eventName string, payload any) ([]byte, error) {
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	sseString := fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, string(dataBytes))
	return []byte(sseString), nil
}
