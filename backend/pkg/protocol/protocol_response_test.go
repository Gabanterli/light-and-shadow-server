package protocol

import "testing"

func TestLoginResponseRoundTripSuccess(t *testing.T) {
	payload := EncodeLoginResponse(true, 42, "session_test_token", "")

	resp, err := DecodeLoginResponse(payload)
	if err != nil {
		t.Fatalf("DecodeLoginResponse returned error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true")
	}

	if resp.AccountID != 42 {
		t.Fatalf("expected account id 42, got %d", resp.AccountID)
	}

	if resp.Token != "session_test_token" {
		t.Fatalf("expected token session_test_token, got %q", resp.Token)
	}

	if resp.ErrorCode != "" {
		t.Fatalf("expected empty error code, got %q", resp.ErrorCode)
	}
}

func TestLoginResponseRoundTripFailure(t *testing.T) {
	payload := EncodeLoginResponse(false, 0, "", "invalid_credentials")

	resp, err := DecodeLoginResponse(payload)
	if err != nil {
		t.Fatalf("DecodeLoginResponse returned error: %v", err)
	}

	if resp.Success {
		t.Fatalf("expected success=false")
	}

	if resp.AccountID != 0 {
		t.Fatalf("expected account id 0, got %d", resp.AccountID)
	}

	if resp.Token != "" {
		t.Fatalf("expected empty token, got %q", resp.Token)
	}

	if resp.ErrorCode != "invalid_credentials" {
		t.Fatalf("expected invalid_credentials, got %q", resp.ErrorCode)
	}
}

func TestCharacterListResponseRoundTripSuccess(t *testing.T) {
	entries := []CharacterListEntry{
		{Name: "Gabriela", Class: "Novice", Level: 1, RaceID: "human"},
		{Name: "TankTest", Class: "Knight", Level: 10, RaceID: "dwarf"},
	}

	payload := EncodeCharacterListResponse(true, "", entries)

	resp, err := DecodeCharacterListResponse(payload)
	if err != nil {
		t.Fatalf("DecodeCharacterListResponse returned error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true")
	}

	if resp.ErrorCode != "" {
		t.Fatalf("expected empty error code, got %q", resp.ErrorCode)
	}

	if len(resp.Characters) != 2 {
		t.Fatalf("expected 2 characters, got %d", len(resp.Characters))
	}

	if resp.Characters[0].Name != "Gabriela" || resp.Characters[0].Class != "Novice" || resp.Characters[0].Level != 1 || resp.Characters[0].RaceID != "human" {
		t.Fatalf("unexpected first character: %+v", resp.Characters[0])
	}

	if resp.Characters[1].Name != "TankTest" || resp.Characters[1].Class != "Knight" || resp.Characters[1].Level != 10 || resp.Characters[1].RaceID != "dwarf" {
		t.Fatalf("unexpected second character: %+v", resp.Characters[1])
	}
}

func TestCharacterListResponseRoundTripFailure(t *testing.T) {
	payload := EncodeCharacterListResponse(false, "not_authenticated", nil)

	resp, err := DecodeCharacterListResponse(payload)
	if err != nil {
		t.Fatalf("DecodeCharacterListResponse returned error: %v", err)
	}

	if resp.Success {
		t.Fatalf("expected success=false")
	}

	if resp.ErrorCode != "not_authenticated" {
		t.Fatalf("expected not_authenticated, got %q", resp.ErrorCode)
	}

	if len(resp.Characters) != 0 {
		t.Fatalf("expected 0 characters, got %d", len(resp.Characters))
	}
}

func TestCharacterSelectResponseRoundTripSuccess(t *testing.T) {
	payload := EncodeCharacterSelectResponse(true, "Gabriela", "")

	resp, err := DecodeCharacterSelectResponse(payload)
	if err != nil {
		t.Fatalf("DecodeCharacterSelectResponse returned error: %v", err)
	}

	if !resp.Success {
		t.Fatalf("expected success=true")
	}

	if resp.CharacterName != "Gabriela" {
		t.Fatalf("expected character Gabriela, got %q", resp.CharacterName)
	}

	if resp.ErrorCode != "" {
		t.Fatalf("expected empty error code, got %q", resp.ErrorCode)
	}
}

func TestCharacterSelectResponseRoundTripFailure(t *testing.T) {
	payload := EncodeCharacterSelectResponse(false, "", "character_not_owned")

	resp, err := DecodeCharacterSelectResponse(payload)
	if err != nil {
		t.Fatalf("DecodeCharacterSelectResponse returned error: %v", err)
	}

	if resp.Success {
		t.Fatalf("expected success=false")
	}

	if resp.CharacterName != "" {
		t.Fatalf("expected empty character name, got %q", resp.CharacterName)
	}

	if resp.ErrorCode != "character_not_owned" {
		t.Fatalf("expected character_not_owned, got %q", resp.ErrorCode)
	}
}

func TestProtocolResponseDecodersRejectMalformedPayloads(t *testing.T) {
	if _, err := DecodeLoginResponse([]byte{1}); err == nil {
		t.Fatalf("expected DecodeLoginResponse to reject malformed payload")
	}

	if _, err := DecodeCharacterListResponse([]byte{1}); err == nil {
		t.Fatalf("expected DecodeCharacterListResponse to reject malformed payload")
	}

	if _, err := DecodeCharacterSelectResponse([]byte{1}); err == nil {
		t.Fatalf("expected DecodeCharacterSelectResponse to reject malformed payload")
	}
}
