package secrets

import (
	"testing"
)

// TestGetStringRepresentationString tests the generic version with string
func TestGetStringRepresentationString(t *testing.T) {
	result := getStringRepresentation("hello")
	expected := "'hello'"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationInt tests the generic version with int
func TestGetStringRepresentationInt(t *testing.T) {
	result := getStringRepresentation(42)
	expected := "42"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationFloat tests the generic version with float64
func TestGetStringRepresentationFloat(t *testing.T) {
	result := getStringRepresentation(3.14)
	expected := "3.140000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationBool tests the generic version with bool
func TestGetStringRepresentationBool(t *testing.T) {
	result := getStringRepresentation(true)
	expected := "true"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationAnyString tests the any version with string
func TestGetStringRepresentationAnyString(t *testing.T) {
	result := getStringRepresentationAny("world")
	expected := "'world'"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationAnyInt tests the any version with int
func TestGetStringRepresentationAnyInt(t *testing.T) {
	result := getStringRepresentationAny(123)
	expected := "123"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationAnyFloat tests the any version with float64
func TestGetStringRepresentationAnyFloat(t *testing.T) {
	result := getStringRepresentationAny(2.718)
	expected := "2.718000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationAnyBool tests the any version with bool
func TestGetStringRepresentationAnyBool(t *testing.T) {
	result := getStringRepresentationAny(false)
	expected := "false"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationAnyNil tests the any version with nil
func TestGetStringRepresentationAnyNil(t *testing.T) {
	result := getStringRepresentationAny(nil)
	expected := "null"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestResultAdd tests that the Result.Add method works with any type
func TestResultAdd(t *testing.T) {
	result := make(Result)
	upperCase := false

	// Add various types
	result.Add("key1", "value1", "", upperCase)
	result.Add("key2", 42, "", upperCase)
	result.Add("key3", 3.14, "", upperCase)
	result.Add("key4", true, "", upperCase)

	// Check if values are added correctly
	if result["key1"] != "value1" {
		t.Errorf("expected key1 to be 'value1', got %v", result["key1"])
	}
	if result["key2"] != 42 {
		t.Errorf("expected key2 to be 42, got %v", result["key2"])
	}
	if result["key3"] != 3.14 {
		t.Errorf("expected key3 to be 3.14, got %v", result["key3"])
	}
	if result["key4"] != true {
		t.Errorf("expected key4 to be true, got %v", result["key4"])
	}
}

// TestResultAddWithPrefix tests that the Result.Add method works with prefix
func TestResultAddWithPrefix(t *testing.T) {
	result := make(Result)
	upperCase := false

	result.Add("key1", "value1", "PREFIX_", upperCase)

	if result["PREFIX_key1"] != "value1" {
		t.Errorf("expected PREFIX_key1 to be 'value1', got %v", result["PREFIX_key1"])
	}
}

// TestResultAddWithUpperCase tests that the Result.Add method works with uppercase
func TestResultAddWithUpperCase(t *testing.T) {
	result := make(Result)
	upperCase := true

	result.Add("key1", "value1", "PREFIX_", upperCase)

	if result["PREFIX_KEY1"] != "value1" {
		t.Errorf("expected PREFIX_KEY1 to be 'value1', got %v", result["PREFIX_KEY1"])
	}
}
