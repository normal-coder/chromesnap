package main

import (
	"strings"
	"testing"
)

func TestExpandPattern_Basic(t *testing.T) {
	got := expandPattern("{index}_{host}", 1, "https://example.com/page", "png")
	want := "0001_example.com.png"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpandPattern_WithPort(t *testing.T) {
	got := expandPattern("{host}", 1, "https://example.com:8080/page", "png")
	want := "example.com_8080.png"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpandPattern_IndexPadded(t *testing.T) {
	got := expandPattern("{index}", 42, "https://a.com", "jpeg")
	want := "0042.jpeg"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestExpandPattern_Timestamp(t *testing.T) {
	got := expandPattern("{ts}", 1, "https://a.com", "png")
	// Should be a unix timestamp (digits only) followed by .png
	if !strings.HasSuffix(got, ".png") {
		t.Errorf("expected .png suffix, got %q", got)
	}
	tsPart := strings.TrimSuffix(got, ".png")
	if len(tsPart) < 10 {
		t.Errorf("timestamp part %q seems too short for a unix timestamp", tsPart)
	}
}

func TestExpandPattern_AllPlaceholders(t *testing.T) {
	got := expandPattern("{index}_{host}_{ts}", 7, "https://sub.domain.com:3000/path", "webp")
	if !strings.HasPrefix(got, "0007_sub.domain.com_3000_") {
		t.Errorf("got %q, expected prefix 0007_sub.domain.com_3000_", got)
	}
	if !strings.HasSuffix(got, ".webp") {
		t.Errorf("expected .webp suffix, got %q", got)
	}
}

func TestExpandPattern_NoScheme(t *testing.T) {
	got := expandPattern("{host}", 1, "example.com/page", "png")
	// Without "://", the entire string is treated as host, then "/" splits it
	want := "example.com.png"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- parseClip ---

func TestParseClip_Valid(t *testing.T) {
	vals, err := parseClip("10,20,300,400")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals != [4]float64{10, 20, 300, 400} {
		t.Errorf("got %v, want [10 20 300 400]", vals)
	}
}

func TestParseClip_InvalidCount(t *testing.T) {
	_, err := parseClip("10,20,300")
	if err == nil {
		t.Error("expected error for 3 values")
	}
}

func TestParseClip_InvalidValue(t *testing.T) {
	_, err := parseClip("abc,20,300,400")
	if err == nil {
		t.Error("expected error for non-numeric value")
	}
}

func TestParseClip_Floats(t *testing.T) {
	vals, err := parseClip("10.5,20.3,300,400")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals[0] != 10.5 || vals[1] != 20.3 {
		t.Errorf("got %v, want [10.5 20.3 300 400]", vals)
	}
}

func TestParseClip_WithSpaces(t *testing.T) {
	vals, err := parseClip(" 10 , 20 , 300 , 400 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals != [4]float64{10, 20, 300, 400} {
		t.Errorf("got %v, want [10 20 300 400]", vals)
	}
}

func TestParseClip_TooMany(t *testing.T) {
	_, err := parseClip("1,2,3,4,5")
	if err == nil {
		t.Error("expected error for 5 values")
	}
}
