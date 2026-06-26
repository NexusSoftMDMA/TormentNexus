package communication

import (
	"context"
	"testing"
)

// --- LinkedInSender construction tests ---

func TestNewLinkedInSender_DefaultsToEmpty(t *testing.T) {
	t.Setenv("LINKEDIN_USERNAME", "")
	t.Setenv("LINKEDIN_PASSWORD", "")

	sender := NewLinkedInSender()
	if sender == nil {
		t.Fatal("expected non-nil sender")
	}
	if sender.Username != "" {
		t.Errorf("expected empty username, got %q", sender.Username)
	}
	if sender.Password != "" {
		t.Errorf("expected empty password, got %q", sender.Password)
	}
}

func TestNewLinkedInSender_ReadsEnv(t *testing.T) {
	t.Setenv("LINKEDIN_USERNAME", "testuser@example.com")
	t.Setenv("LINKEDIN_PASSWORD", "secret123")

	sender := NewLinkedInSender()
	if sender.Username != "testuser@example.com" {
		t.Errorf("expected username from env, got %q", sender.Username)
	}
	if sender.Password != "secret123" {
		t.Errorf("expected password from env, got %q", sender.Password)
	}
}

// --- HealthCheck tests ---

func TestLinkedInSender_HealthCheck_FailsWithoutCredentials(t *testing.T) {
	sender := &LinkedInSender{Username: "", Password: ""}
	err := sender.HealthCheck(context.Background())
	if err == nil {
		t.Error("expected error when credentials are missing")
	}
}

func TestLinkedInSender_HealthCheck_FailsWithPartialCredentials(t *testing.T) {
	sender := &LinkedInSender{Username: "user@test.com", Password: ""}
	err := sender.HealthCheck(context.Background())
	if err == nil {
		t.Error("expected error when password is missing")
	}

	sender2 := &LinkedInSender{Username: "", Password: "pass"}
	err = sender2.HealthCheck(context.Background())
	if err == nil {
		t.Error("expected error when username is missing")
	}
}

func TestLinkedInSender_HealthCheck_PassesWithCredentials(t *testing.T) {
	sender := &LinkedInSender{Username: "user@test.com", Password: "pass123"}
	err := sender.HealthCheck(context.Background())
	if err != nil {
		t.Errorf("expected no error with valid credentials, got: %v", err)
	}
}

// --- Simulation fallback tests ---

func TestLinkedInSender_Send_SimulationMode(t *testing.T) {
	sender := &LinkedInSender{Username: "", Password: ""}

	msg := LinkedInMessage{
		ProfileURL: "https://linkedin.com/in/janedoe",
		Subject:    "TormentNexus — AI Infrastructure",
		Body:       "Hi Jane, I noticed your team is working on AI orchestration...",
	}

	err := sender.Send(context.Background(), msg)
	if err != nil {
		t.Errorf("simulation send should not return error, got: %v", err)
	}
}

func TestLinkedInSender_SendConnectionRequest_SimulationMode(t *testing.T) {
	sender := &LinkedInSender{Username: "", Password: ""}

	err := sender.SendConnectionRequest(
		context.Background(),
		"https://linkedin.com/in/janedoe",
		"Hi Jane, I'd love to connect about AI infrastructure.",
	)
	if err != nil {
		t.Errorf("simulation connection request should not return error, got: %v", err)
	}
}

// --- LinkedInMessage struct tests ---

func TestLinkedInMessage_Fields(t *testing.T) {
	msg := LinkedInMessage{
		ProfileURL: "https://linkedin.com/in/testuser",
		Subject:    "Test Subject",
		Body:       "Test Body",
	}

	if msg.ProfileURL != "https://linkedin.com/in/testuser" {
		t.Errorf("unexpected ProfileURL: %q", msg.ProfileURL)
	}
	if msg.Subject != "Test Subject" {
		t.Errorf("unexpected Subject: %q", msg.Subject)
	}
	if msg.Body != "Test Body" {
		t.Errorf("unexpected Body: %q", msg.Body)
	}
}

// --- simulateSend tests ---

func TestLinkedInSender_SimulateSend_DoesNotError(t *testing.T) {
	sender := &LinkedInSender{}

	msg := LinkedInMessage{
		ProfileURL: "https://linkedin.com/in/someone",
		Subject:    "Hello",
		Body:       "World",
	}

	err := sender.simulateSend(context.Background(), msg)
	if err != nil {
		t.Errorf("simulateSend should not error, got: %v", err)
	}
}
