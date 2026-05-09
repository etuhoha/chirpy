package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT_Basic(t *testing.T) {
	id := uuid.New()
	secret := "passw0rd"

	jwt, err := MakeJWT(id, secret, 10*time.Second)
	if err != nil {
		t.Fatalf("unexpected error on create: %v", err)
	}

	parsedId, err := ValidateJWT(jwt, secret)
	if err != nil {
		t.Fatalf("unexpected error on validate: %v", err)
	}

	if parsedId != id {
		t.Fatalf("wrong ID: expected %v, got %v", id, parsedId)
	}

	_, err = ValidateJWT(jwt, secret+"foo")
	if err == nil {
		t.Fatalf("expected success on wrong password")
	}
}

func TestJWT_Expiration(t *testing.T) {
	id := uuid.New()
	expiresIn := 50 * time.Millisecond
	secret := "passw0rd"

	jwt, err := MakeJWT(id, secret, expiresIn)
	if err != nil {
		t.Fatalf("unexpected error on create: %v", err)
	}

	time.Sleep(expiresIn)

	_, err = ValidateJWT(jwt, secret)
	if err == nil {
		t.Fatalf("expected success after expiration")
	}
}
func TestGetBearerToken(t *testing.T) {
	h := http.Header{}
	_, err := GetBearerToken(h)
	if err == nil {
		t.Fatal("expected error on empty headers")
	}

	h.Set("Authorization", "foo")
	_, err = GetBearerToken(h)
	if err == nil {
		t.Fatal("expected error on bad prefix")
	}

	expected := "foo"
	h.Set("Authorization", "Bearer   "+expected+" ")
	token, err := GetBearerToken(h)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != expected {
		t.Fatalf("expected %v, got %v", expected, token)
	}
}
