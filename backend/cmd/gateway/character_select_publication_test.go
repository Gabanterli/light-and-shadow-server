package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/light-and-shadow/backend/pkg/protocol"
)

type characterSelectPublicationConn struct {
	writeFn func([]byte) (int, error)
}

func (c *characterSelectPublicationConn) Read(
	_ []byte,
) (int, error) {
	return 0, io.EOF
}

func (c *characterSelectPublicationConn) Write(
	payload []byte,
) (int, error) {
	if c.writeFn == nil {
		return 0, errors.New("unexpected write")
	}

	return c.writeFn(payload)
}

func (*characterSelectPublicationConn) Close() error {
	return nil
}

func (*characterSelectPublicationConn) LocalAddr() net.Addr {
	return characterSelectPublicationAddr("local")
}

func (*characterSelectPublicationConn) RemoteAddr() net.Addr {
	return characterSelectPublicationAddr("remote")
}

func (*characterSelectPublicationConn) SetDeadline(
	_ time.Time,
) error {
	return nil
}

func (*characterSelectPublicationConn) SetReadDeadline(
	_ time.Time,
) error {
	return nil
}

func (*characterSelectPublicationConn) SetWriteDeadline(
	_ time.Time,
) error {
	return nil
}

type characterSelectPublicationAddr string

func (characterSelectPublicationAddr) Network() string {
	return "test"
}

func (address characterSelectPublicationAddr) String() string {
	return string(address)
}

func newCharacterSelectPublicationResponse() *protocol.Packet {
	return &protocol.Packet{
		Opcode:   protocol.SC_CHAR_SELECT_RESPONSE,
		Sequence: 77,
		Payload: protocol.EncodeCharacterSelectResponse(
			true,
			"Gabriela",
			"",
		),
	}
}

func TestPublishCharacterSelectionSuccessOrdersWriteRegisterSync(
	t *testing.T,
) {
	events := make([]string, 0, 3)

	conn := &characterSelectPublicationConn{
		writeFn: func(payload []byte) (int, error) {
			events = append(events, "write")
			return len(payload), nil
		},
	}

	err := publishCharacterSelectionSuccess(
		conn,
		newCharacterSelectPublicationResponse(),
		"Gabriela",
		func(playerID string, registered net.Conn) {
			if playerID != "Gabriela" {
				t.Fatalf("registered player = %q", playerID)
			}

			if registered != conn {
				t.Fatal("registered connection differs")
			}

			events = append(events, "register")
		},
		func(playerID string) {
			if playerID != "Gabriela" {
				t.Fatalf("synced player = %q", playerID)
			}

			events = append(events, "sync")
		},
	)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	want := []string{"write", "register", "sync"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPublishCharacterSelectionSuccessRegistersOnlyAfterFullPartialWrite(
	t *testing.T,
) {
	response := newCharacterSelectPublicationResponse()
	expected := response.Serialize()

	var written bytes.Buffer
	writeCalls := 0
	completeAtRegister := false

	conn := &characterSelectPublicationConn{
		writeFn: func(payload []byte) (int, error) {
			writeCalls++

			limit := 3
			if len(payload) < limit {
				limit = len(payload)
			}

			_, _ = written.Write(payload[:limit])
			return limit, nil
		},
	}

	err := publishCharacterSelectionSuccess(
		conn,
		response,
		"Gabriela",
		func(_ string, _ net.Conn) {
			completeAtRegister = bytes.Equal(
				written.Bytes(),
				expected,
			)
		},
		func(string) {},
	)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	if writeCalls <= 1 {
		t.Fatalf("write calls = %d, want more than one", writeCalls)
	}

	if !completeAtRegister {
		t.Fatal("registration happened before the response was fully written")
	}

	if !bytes.Equal(written.Bytes(), expected) {
		t.Fatal("written response differs from expected bytes")
	}
}

func TestPublishCharacterSelectionSuccessFirstWriteErrorDoesNotPublish(
	t *testing.T,
) {
	sentinel := errors.New("first write failed")
	published := false

	conn := &characterSelectPublicationConn{
		writeFn: func([]byte) (int, error) {
			return 0, sentinel
		},
	}

	err := publishCharacterSelectionSuccess(
		conn,
		newCharacterSelectPublicationResponse(),
		"Gabriela",
		func(string, net.Conn) {
			published = true
		},
		func(string) {
			published = true
		},
	)

	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}

	if published {
		t.Fatal("callbacks ran after first write error")
	}
}

func TestPublishCharacterSelectionSuccessPartialThenErrorDoesNotPublish(
	t *testing.T,
) {
	sentinel := errors.New("write failed after partial progress")
	calls := 0
	registerCalls := 0
	syncCalls := 0

	conn := &characterSelectPublicationConn{
		writeFn: func(payload []byte) (int, error) {
			calls++

			if calls == 1 {
				return 2, nil
			}

			if len(payload) == 0 {
				t.Fatal("second write received no payload")
			}

			return 1, sentinel
		},
	}

	err := publishCharacterSelectionSuccess(
		conn,
		newCharacterSelectPublicationResponse(),
		"Gabriela",
		func(string, net.Conn) {
			registerCalls++
		},
		func(string) {
			syncCalls++
		},
	)

	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}

	if registerCalls != 0 || syncCalls != 0 {
		t.Fatalf(
			"callbacks after error: register=%d sync=%d",
			registerCalls,
			syncCalls,
		)
	}
}

func TestPublishCharacterSelectionSuccessZeroWriteDoesNotPublish(
	t *testing.T,
) {
	published := false

	conn := &characterSelectPublicationConn{
		writeFn: func([]byte) (int, error) {
			return 0, nil
		},
	}

	err := publishCharacterSelectionSuccess(
		conn,
		newCharacterSelectPublicationResponse(),
		"Gabriela",
		func(string, net.Conn) {
			published = true
		},
		func(string) {
			published = true
		},
	)

	if !errors.Is(err, io.ErrNoProgress) {
		t.Fatalf("error = %v, want io.ErrNoProgress", err)
	}

	if published {
		t.Fatal("callbacks ran after zero-progress write")
	}
}

func TestPublishCharacterSelectionSuccessRejectsInvalidInputs(
	t *testing.T,
) {
	validConn := &characterSelectPublicationConn{
		writeFn: func(payload []byte) (int, error) {
			return len(payload), nil
		},
	}
	validResponse := newCharacterSelectPublicationResponse()
	validRegister := func(string, net.Conn) {}
	validSync := func(string) {}

	tests := []struct {
		name     string
		conn     net.Conn
		response *protocol.Packet
		playerID string
		register func(string, net.Conn)
		sync     func(string)
	}{
		{
			name:     "nil connection",
			response: validResponse,
			playerID: "Gabriela",
			register: validRegister,
			sync:     validSync,
		},
		{
			name:     "nil response",
			conn:     validConn,
			playerID: "Gabriela",
			register: validRegister,
			sync:     validSync,
		},
		{
			name:     "empty player",
			conn:     validConn,
			response: validResponse,
			register: validRegister,
			sync:     validSync,
		},
		{
			name:     "whitespace player",
			conn:     validConn,
			response: validResponse,
			playerID: "   ",
			register: validRegister,
			sync:     validSync,
		},
		{
			name:     "nil register",
			conn:     validConn,
			response: validResponse,
			playerID: "Gabriela",
			sync:     validSync,
		},
		{
			name:     "nil sync",
			conn:     validConn,
			response: validResponse,
			playerID: "Gabriela",
			register: validRegister,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			err := publishCharacterSelectionSuccess(
				test.conn,
				test.response,
				test.playerID,
				test.register,
				test.sync,
			)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestPublishCharacterSelectionSuccessTrimsPlayerAndCallsOnce(
	t *testing.T,
) {
	registerCalls := 0
	syncCalls := 0
	registeredPlayer := ""
	syncedPlayer := ""

	conn := &characterSelectPublicationConn{
		writeFn: func(payload []byte) (int, error) {
			return len(payload), nil
		},
	}

	err := publishCharacterSelectionSuccess(
		conn,
		newCharacterSelectPublicationResponse(),
		"  Gabriela  ",
		func(playerID string, _ net.Conn) {
			registerCalls++
			registeredPlayer = playerID
		},
		func(playerID string) {
			syncCalls++
			syncedPlayer = playerID
		},
	)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	if registerCalls != 1 || syncCalls != 1 {
		t.Fatalf(
			"callback counts: register=%d sync=%d",
			registerCalls,
			syncCalls,
		)
	}

	if registeredPlayer != "Gabriela" {
		t.Fatalf("registered player = %q", registeredPlayer)
	}

	if syncedPlayer != "Gabriela" {
		t.Fatalf("synced player = %q", syncedPlayer)
	}

	if strings.TrimSpace(registeredPlayer) != registeredPlayer {
		t.Fatal("registered player was not canonicalized")
	}
}
