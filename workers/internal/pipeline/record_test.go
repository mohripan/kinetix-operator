package pipeline

import (
	"strings"
	"testing"
	"time"
)

func TestUserEventRoundTrip(t *testing.T) {
	event := NewUserEvent(" record-1 ", " User-42 ", " Page_View ", time.Date(2026, 5, 1, 12, 0, 0, 0, time.FixedZone("WIB", 7*60*60)), map[string]string{"Path": " /pricing "})

	data, err := EncodeUserEvent(event)
	if err != nil {
		t.Fatalf("EncodeUserEvent() error = %v", err)
	}

	got, err := DecodeUserEvent(data)
	if err != nil {
		t.Fatalf("DecodeUserEvent() error = %v", err)
	}

	if got.ID != "record-1" {
		t.Fatalf("ID = %q, want record-1", got.ID)
	}
	if got.EventType != "page_view" {
		t.Fatalf("EventType = %q, want page_view", got.EventType)
	}
	if got.Timestamp.Location() != time.UTC {
		t.Fatalf("Timestamp location = %v, want UTC", got.Timestamp.Location())
	}
}

func TestDecodeUserEventRequiresID(t *testing.T) {
	_, err := DecodeUserEvent([]byte(`{"user_id":"u1","event_type":"click","timestamp":"2026-05-01T00:00:00Z"}`))
	if err == nil || !strings.Contains(err.Error(), "record id") {
		t.Fatalf("DecodeUserEvent() error = %v, want record id error", err)
	}
}

func TestNormalizeProducesLineageFields(t *testing.T) {
	input := NewUserEvent("evt-1", "USER-1", "signup", time.Now(), map[string]string{" Campaign ": " spring "})
	processedAt := time.Date(2026, 5, 1, 10, 30, 0, 0, time.UTC)

	got := Normalize(input, DefaultInputTopic, processedAt)

	if got.ParentID != "evt-1" {
		t.Fatalf("ParentID = %q, want evt-1", got.ParentID)
	}
	if got.UserID != "user-1" {
		t.Fatalf("UserID = %q, want user-1", got.UserID)
	}
	if got.EventType != "SIGNUP" {
		t.Fatalf("EventType = %q, want SIGNUP", got.EventType)
	}
	if got.SourceTopic != DefaultInputTopic {
		t.Fatalf("SourceTopic = %q, want %q", got.SourceTopic, DefaultInputTopic)
	}
	if got.SchemaFingerprint == "" {
		t.Fatal("SchemaFingerprint is empty")
	}
	if got.Attrs["campaign"] != "spring" {
		t.Fatalf("Attrs[campaign] = %q, want spring", got.Attrs["campaign"])
	}
}
