package main

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/light-and-shadow/backend/pkg/protocol"
)

// publishCharacterSelectionSuccess establishes the publication barrier for an
// activated character session. The acknowledgement is written completely
// before AOI registration makes the connection reachable by asynchronous
// subsystems.
func publishCharacterSelectionSuccess(
	conn net.Conn,
	response *protocol.Packet,
	playerID string,
	registerPlayer func(string, net.Conn),
	syncActiveQuests func(string),
) error {
	if conn == nil {
		return fmt.Errorf("publish character selection: connection is nil")
	}

	if response == nil {
		return fmt.Errorf("publish character selection: response is nil")
	}

	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return fmt.Errorf("publish character selection: player ID is empty")
	}

	if registerPlayer == nil {
		return fmt.Errorf("publish character selection: register callback is nil")
	}

	if syncActiveQuests == nil {
		return fmt.Errorf("publish character selection: quest sync callback is nil")
	}

	serialized := response.Serialize()
	if len(serialized) == 0 {
		return fmt.Errorf("publish character selection: serialized response is empty")
	}

	if err := writeAllCharacterSelectionResponse(conn, serialized); err != nil {
		return fmt.Errorf("publish character selection acknowledgement: %w", err)
	}

	registerPlayer(playerID, conn)
	syncActiveQuests(playerID)

	return nil
}

func writeAllCharacterSelectionResponse(
	conn net.Conn,
	remaining []byte,
) error {
	for len(remaining) > 0 {
		written, err := conn.Write(remaining)

		if written < 0 || written > len(remaining) {
			return fmt.Errorf(
				"invalid character selection write count: %d for %d bytes",
				written,
				len(remaining),
			)
		}

		if written > 0 {
			remaining = remaining[written:]
		}

		if err != nil {
			return err
		}

		if written == 0 {
			return io.ErrNoProgress
		}
	}

	return nil
}
