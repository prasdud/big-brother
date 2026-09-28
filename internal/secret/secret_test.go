package secret

import "testing"

func TestBoxRoundTrip(t *testing.T) {
	box, err := NewBox("test-key")
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("xoxb-secret-token")
	ciphertext, err := box.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if string(ciphertext) == string(plaintext) {
		t.Fatal("ciphertext equals plaintext")
	}
	got, err := box.Decrypt(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plaintext) {
		t.Fatalf("got %q, want %q", got, plaintext)
	}
}

func TestBoxRejectsOtherKey(t *testing.T) {
	a, _ := NewBox("key-a")
	b, _ := NewBox("key-b")
	ciphertext, _ := a.Encrypt([]byte("secret"))
	if _, err := b.Decrypt(ciphertext); err == nil {
		t.Fatal("expected decryption failure with a different key")
	}
}

func TestNewBoxRequiresKey(t *testing.T) {
	if _, err := NewBox(""); err == nil {
		t.Fatal("expected error for empty key")
	}
}
