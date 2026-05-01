package pipeline

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	DefaultInputTopic  = "kinetix-input"
	DefaultOutputTopic = "kinetix-output"
	DefaultSchema      = "user-event-v1"
)

type UserEvent struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	EventType string            `json:"event_type"`
	Timestamp time.Time         `json:"timestamp"`
	Attrs     map[string]string `json:"attrs,omitempty"`
}

type NormalizedEvent struct {
	ID                string            `json:"id"`
	ParentID          string            `json:"parent_id"`
	UserID            string            `json:"user_id"`
	EventType         string            `json:"event_type"`
	Schema            string            `json:"schema"`
	SourceTopic       string            `json:"source_topic"`
	ProcessedAt       time.Time         `json:"processed_at"`
	SchemaFingerprint string            `json:"schema_fingerprint"`
	Attrs             map[string]string `json:"attrs,omitempty"`
}

func NewUserEvent(id, userID, eventType string, ts time.Time, attrs map[string]string) UserEvent {
	return UserEvent{
		ID:        strings.TrimSpace(id),
		UserID:    strings.TrimSpace(userID),
		EventType: strings.ToLower(strings.TrimSpace(eventType)),
		Timestamp: ts.UTC(),
		Attrs:     attrs,
	}
}

func (e UserEvent) Validate() error {
	if e.ID == "" {
		return errors.New("record id is required")
	}
	if e.UserID == "" {
		return errors.New("user id is required")
	}
	if e.EventType == "" {
		return errors.New("event type is required")
	}
	if e.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}

func EncodeUserEvent(e UserEvent) ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}

func DecodeUserEvent(data []byte) (UserEvent, error) {
	var event UserEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return UserEvent{}, fmt.Errorf("decode user event: %w", err)
	}
	event.ID = strings.TrimSpace(event.ID)
	event.UserID = strings.TrimSpace(event.UserID)
	event.EventType = strings.ToLower(strings.TrimSpace(event.EventType))
	event.Timestamp = event.Timestamp.UTC()
	if err := event.Validate(); err != nil {
		return UserEvent{}, err
	}
	return event, nil
}

func Normalize(event UserEvent, sourceTopic string, processedAt time.Time) NormalizedEvent {
	attrs := make(map[string]string, len(event.Attrs))
	for k, v := range event.Attrs {
		attrs[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}

	return NormalizedEvent{
		ID:                stableOutputID(event.ID, sourceTopic),
		ParentID:          event.ID,
		UserID:            strings.ToLower(event.UserID),
		EventType:         strings.ToUpper(event.EventType),
		Schema:            DefaultSchema,
		SourceTopic:       sourceTopic,
		ProcessedAt:       processedAt.UTC(),
		SchemaFingerprint: schemaFingerprint(DefaultSchema),
		Attrs:             attrs,
	}
}

func EncodeNormalizedEvent(e NormalizedEvent) ([]byte, error) {
	if e.ID == "" || e.ParentID == "" {
		return nil, errors.New("normalized event id and parent id are required")
	}
	return json.Marshal(e)
}

func stableOutputID(parentID, sourceTopic string) string {
	sum := sha256.Sum256([]byte(sourceTopic + ":" + parentID))
	return "out_" + hex.EncodeToString(sum[:8])
}

func schemaFingerprint(schema string) string {
	sum := sha256.Sum256([]byte(schema))
	return hex.EncodeToString(sum[:8])
}
