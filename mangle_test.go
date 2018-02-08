package lh

import (
	"net/http"
	"testing"
	"time"
)

func TestMangle(t *testing.T) {
	h := make(http.Header)
	h.Set("Via", "127.0.0.1")
	mangle(h)
	if h.Get("X-Amzn-Remapped-Via") != "127.0.0.1" {
		t.Fatal("Mangle didn't happen")
	}
}

func TestNoMangle(t *testing.T) {
	h := make(http.Header)
	h.Set("Via", "127.0.0.1")
	h.Set("X-Amzn-Remapped-Via", "10.10.10.10")
	mangle(h)
	if h.Get("X-Amzn-Remapped-Via") != "10.10.10.10" {
		t.Fatal("No mangle should have happened")
	}
}

func TestDemangle(t *testing.T) {
	h := make(http.Header)
	now := time.Now().Format(time.RFC850)
	h.Set("X-Amzn-Remapped-Date", now)
	demangle(h)
	if h.Get("Date") != now {
		t.Fatal("Demangle didn't happen")
	}
}

func TestNoDemangle(t *testing.T) {
	h := make(http.Header)
	h.Set("Via", "127.0.0.1")
	h.Set("X-Amzn-Remapped-Via", "10.10.10.10")
	demangle(h)
	if h.Get("Via") != "127.0.0.1" {
		t.Fatal("No demangle should have happened")
	}
}
