package pi

import "testing"

func TestSessionStoreCreateAppendListAndFork(t *testing.T) {
	dir := t.TempDir()
	store := NewSessionStore(dir)
	session, err := store.Create("alpha", "/workspace/project")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AppendEntry(session.Metadata.SessionID, SessionEntry{Kind: "message", Role: "user", Text: "hello"}); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load(session.Metadata.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	listed, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].SessionID != session.Metadata.SessionID {
		t.Fatalf("unexpected session list: %#v", listed)
	}
	forked, err := store.Fork(session.Metadata.SessionID, loaded.Entries[0].ID, "beta")
	if err != nil {
		t.Fatal(err)
	}
	if forked.Metadata.SessionID == session.Metadata.SessionID {
		t.Fatal("expected forked session to have a new id")
	}
	if len(forked.Entries) != 1 {
		t.Fatalf("expected forked session to copy entries, got %d", len(forked.Entries))
	}
}
