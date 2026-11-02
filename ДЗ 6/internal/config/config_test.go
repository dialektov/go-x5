package config

import "testing"

func TestChannelFromEnvironment(t *testing.T) {
	t.Setenv("NOTIFY_CHANNEL", "push")

	if got := Channel(); got != "push" {
		t.Fatalf("канал = %q", got)
	}
}

func TestChannelDefault(t *testing.T) {
	t.Setenv("NOTIFY_CHANNEL", "")

	if got := Channel(); got != "email" {
		t.Fatalf("канал = %q", got)
	}
}
