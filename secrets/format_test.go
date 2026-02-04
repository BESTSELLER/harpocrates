package secrets

import (
	"testing"
)

// TestGetStringRepresentationString tests with string value
func TestGetStringRepresentationString(t *testing.T) {
	result := getStringRepresentation("hello")
	expected := "'hello'"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationInt tests with int value
func TestGetStringRepresentationInt(t *testing.T) {
	result := getStringRepresentation(42)
	expected := "42"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationFloat tests with float64 value
func TestGetStringRepresentationFloat(t *testing.T) {
	result := getStringRepresentation(3.14)
	expected := "3.140000"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationBool tests with bool value
func TestGetStringRepresentationBool(t *testing.T) {
	result := getStringRepresentation(true)
	expected := "true"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGetStringRepresentationNil tests with nil value
func TestGetStringRepresentationNil(t *testing.T) {
	result := getStringRepresentation(nil)
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
