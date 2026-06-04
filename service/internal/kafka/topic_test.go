package kafka_test

import (
	"testing"

	"service-starter/service/internal/kafka"
)

func TestTopicNameWithPrefix(t *testing.T) {
	got := kafka.TopicName("local", "service", "events")
	want := "local.service.events"
	if got != want {
		t.Fatalf("TopicName() = %q, want %q", got, want)
	}
}

func TestTopicNameWithoutPrefix(t *testing.T) {
	got := kafka.TopicName("", "service", "commands")
	want := "service.commands"
	if got != want {
		t.Fatalf("TopicName() = %q, want %q", got, want)
	}
}

func TestTopicNameTrimsParts(t *testing.T) {
	got := kafka.TopicName(" local ", " service ", " events ")
	want := "local.service.events"
	if got != want {
		t.Fatalf("TopicName() = %q, want %q", got, want)
	}
}
