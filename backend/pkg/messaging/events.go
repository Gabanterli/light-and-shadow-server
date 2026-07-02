package messaging

// MonsterKilledPayload define os dados enviados quando um monstro é derrotado.
type MonsterKilledPayload struct {
	PlayerID  string
	MonsterID string
}

// ItemLootedPayload define os dados enviados quando um item é saqueado.
type ItemLootedPayload struct {
	PlayerID string
	ItemID   string
	Qty      int
}

// NPCInteractedPayload define os dados enviados quando um jogador interage com um NPC.
type NPCInteractedPayload struct {
	PlayerID string
	NPCID    string
}

// LocationReachedPayload define os dados enviados quando um jogador atinge uma localização.
type LocationReachedPayload struct {
	PlayerID string
	X        float64
	Y        float64
	Z        float64
}
