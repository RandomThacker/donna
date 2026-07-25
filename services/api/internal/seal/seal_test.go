package seal

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := KeyFromSecret("test-secret-for-credentials")
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := Encrypt(key, []byte(`{"refresh_token":"rt"}`))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := Decrypt(key, sealed)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != `{"refresh_token":"rt"}` {
		t.Fatalf("plain = %s", plain)
	}
}

func TestDecryptTamperedFails(t *testing.T) {
	key, err := KeyFromSecret("test-secret-for-credentials")
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := Encrypt(key, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	sealed[len(sealed)-1] ^= 0xff
	if _, err := Decrypt(key, sealed); err == nil {
		t.Fatal("expected error")
	}
}
