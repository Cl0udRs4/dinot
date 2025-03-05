package encryption

import (
    "encoding/json"
    "time"
)

// MessageHeader represents the header of an encrypted message
type MessageHeader struct {
    Version    int           `json:"version"`
    Encryption EncryptionType `json:"encryption"`
    KeyID      uint32        `json:"key_id"`
    Timestamp  int64         `json:"timestamp"`
}

// Message represents an encrypted message
type Message struct {
    Header  MessageHeader `json:"header"`
    Payload []byte        `json:"payload"`
}

// NewMessage creates a new message with the specified encryption type and payload
func NewMessage(encryptionType EncryptionType, keyID uint32, payload []byte) *Message {
    return &Message{
        Header: MessageHeader{
            Version:    1,
            Encryption: encryptionType,
            KeyID:      keyID,
            Timestamp:  time.Now().Unix(),
        },
        Payload: payload,
    }
}

// ToJSON converts the message to JSON
func (m *Message) ToJSON() ([]byte, error) {
    return json.Marshal(m)
}

// FromJSON updates the message from JSON
func (m *Message) FromJSON(data []byte) error {
    return json.Unmarshal(data, m)
}
