package utils

import "testing"

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" || hash == "correct horse battery staple" {
		t.Fatalf("HashPassword returned an invalid hash: %q", hash)
	}
	if !CheckPassword("correct horse battery staple", hash) {
		t.Fatal("CheckPassword rejected the original password")
	}
	if CheckPassword("wrong password", hash) {
		t.Fatal("CheckPassword accepted a wrong password")
	}
	if CheckPassword("anything", "not-a-bcrypt-hash") {
		t.Fatal("CheckPassword accepted an invalid hash")
	}
}
