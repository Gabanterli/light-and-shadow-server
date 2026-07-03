package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// Opcodes correspondentes ao cliente C#
const (
	CS_HEARTBEAT           uint16 = 1000
	SC_HEARTBEAT_ACK       uint16 = 1001
	CS_LOGIN_REQUEST       uint16 = 1002
	SC_LOGIN_RESPONSE      uint16 = 1003
	CS_CHAR_LIST_REQUEST   uint16 = 1004
	SC_CHAR_LIST_RESPONSE  uint16 = 1005
	CS_CHAR_SELECT_REQUEST uint16 = 1006
	SC_CHAR_SELECT_RESPONSE uint16 = 1007

	CS_PLAYER_MOVE    uint16 = 2000
	SC_PLAYER_UPDATE  uint16 = 2001
	SC_SPAWN_ENTITY   uint16 = 2002
	SC_DESPAWN_ENTITY uint16 = 2003

	CS_MOVE_REQUEST   uint16 = 2004
	SC_MOVE_CONFIRM   uint16 = 2005
	SC_CHUNK_DATA     uint16 = 2006

	// Combat System Opcodes (Sprint 2 Task 5)
	CS_ATTACK_REQUEST uint16 = 3000
	CS_CAST_SKILL     uint16 = 3001
	SC_DAMAGE_EVENT   uint16 = 3002
	SC_TARGET_DEAD    uint16 = 3003

	// Inventory System Opcodes (Sprint 3 Task 1)
	CS_INVENTORY_REQUEST uint16 = 4000
	SC_INVENTORY_SYNC    uint16 = 4001
	CS_EQUIP_ITEM        uint16 = 4002
	SC_EQUIP_RESPONSE    uint16 = 4003
	CS_UNEQUIP_ITEM      uint16 = 4004
	SC_UNEQUIP_RESPONSE  uint16 = 4005
	CS_SWAP_SLOTS        uint16 = 4006
	SC_SWAP_RESPONSE     uint16 = 4007

	// NPC & Quest System Opcodes (Sprint 3 Task 3)
	CS_NPC_INTERACT       uint16 = 5000
	SC_DIALOGUE_OPEN      uint16 = 5001
	CS_DIALOGUE_RESPONSE  uint16 = 5002
	CS_ACCEPT_QUEST       uint16 = 5003
	CS_COMPLETE_QUEST     uint16 = 5004
	SC_QUEST_UPDATE       uint16 = 5005
	SC_QUEST_COMPLETE     uint16 = 5006

	// Social, Party & Guild System Opcodes (Sprint 3 Task 3 - Party, Guild & Social)
	CS_PARTY_CREATE        uint16 = 6000
	SC_PARTY_INFO          uint16 = 6001
	CS_PARTY_INVITE        uint16 = 6002
	SC_PARTY_INVITE_REQ    uint16 = 6003
	CS_PARTY_INVITE_RESP   uint16 = 6004
	CS_PARTY_LEAVE         uint16 = 6005
	CS_PARTY_KICK          uint16 = 6006
	CS_PARTY_TRANSFER      uint16 = 6007
	CS_PARTY_LOOT_MODE     uint16 = 6008

	CS_GUILD_CREATE        uint16 = 6100
	SC_GUILD_INFO          uint16 = 6101
	CS_GUILD_INVITE        uint16 = 6102
	SC_GUILD_INVITE_REQ    uint16 = 6103
	CS_GUILD_INVITE_RESP   uint16 = 6104
	CS_GUILD_LEAVE         uint16 = 6105
	CS_GUILD_KICK          uint16 = 6106
	CS_GUILD_PROMOTE       uint16 = 6107
	CS_GUILD_DEMOTE        uint16 = 6108
	CS_GUILD_MOTD          uint16 = 6109
	SC_GUILD_AUDIT_LOG     uint16 = 6110

	CS_SOCIAL_ADD_FRIEND   uint16 = 6200
	CS_SOCIAL_REMOVE_FRIEND uint16 = 6201
	CS_SOCIAL_ADD_IGNORE   uint16 = 6202
	CS_SOCIAL_REMOVE_IGNORE uint16 = 6203
	SC_SOCIAL_LISTS        uint16 = 6204
	SC_ONLINE_STATUS       uint16 = 6205

	CS_CHAT_SEND           uint16 = 6300
	SC_CHAT_MESSAGE        uint16 = 6301

	// Economy, Trading & Marketplace System Opcodes (Sprint 3 Task 4)
	CS_TRADE_REQUEST       uint16 = 7000
	SC_TRADE_PROPOSAL      uint16 = 7001
	CS_TRADE_RESPOND       uint16 = 7002
	CS_TRADE_OFFER_ITEM    uint16 = 7003
	CS_TRADE_OFFER_GOLD    uint16 = 7004
	CS_TRADE_LOCK          uint16 = 7005
	CS_TRADE_CONFIRM       uint16 = 7006
	SC_TRADE_UPDATE        uint16 = 7007
	CS_TRADE_CANCEL        uint16 = 7008
	SC_TRADE_CLOSED        uint16 = 7009

	CS_NPC_SHOP_BUY        uint16 = 7100
	CS_NPC_SHOP_SELL       uint16 = 7101
	CS_NPC_SHOP_REPAIR     uint16 = 7102
	SC_NPC_SHOP_RESPONSE   uint16 = 7103

	CS_MARKET_CREATE_ORDER uint16 = 7200
	CS_MARKET_BUY_ITEM     uint16 = 7201
	CS_MARKET_CANCEL_ORDER uint16 = 7202
	CS_MARKET_SEARCH       uint16 = 7203
	SC_MARKET_SEARCH_RESULT uint16 = 7204

	// Crafting & Gathering System Opcodes (Sprint 4 Task 1)
	CS_GATHER_START         uint16 = 8000
	SC_GATHER_PROGRESS      uint16 = 8001
	SC_GATHER_COMPLETE      uint16 = 8002
	CS_CRAFT_START          uint16 = 8003
	SC_CRAFT_RESPONSE       uint16 = 8004
	SC_PROFESSION_XP_UPDATE uint16 = 8005
	CS_GATHER_CANCEL        uint16 = 8006

	// Dungeon, Raid & World Boss Opcodes (Sprint 4 Task 2)
	CS_DUNGEON_ENTER         uint16 = 9000
	SC_DUNGEON_STATE         uint16 = 9001
	SC_BOSS_AI_TELEGRAPH     uint16 = 9002
	SC_BOSS_AI_PHASE         uint16 = 9003
	CS_DUNGEON_CLAIM_LOOT    uint16 = 9004
	SC_LOOT_NOTIFICATION     uint16 = 9005
	CS_DUNGEON_LEAVE         uint16 = 9006
	CS_WORLD_BOSS_SPAWN_REQ  uint16 = 9007

	// Progression & Class/Subclass Opcodes (Sprint 3 Task 5)
	CS_CHOOSE_VOCATION       uint16 = 9100
	SC_CHOOSE_VOCATION_RESP  uint16 = 9101
	CS_UNLOCK_SUBCLASS       uint16 = 9102
	SC_UNLOCK_SUBCLASS_RESP  uint16 = 9103
)

const HeaderSize = 8
const MaxPacketSize = 16384 // 16 KB

// Packet representa o cabeçalho oficial de 8 bytes e payload
type Packet struct {
	Size     uint16 // Total do pacote incluindo cabeçalho
	Opcode   uint16
	Sequence uint32
	Payload  []byte
}

// ReadPacket lê um único pacote de uma conexão TCP respeitando o protocolo (Little Endian)
func ReadPacket(reader io.Reader) (*Packet, error) {
	headerBuf := make([]byte, HeaderSize)
	_, err := io.ReadFull(reader, headerBuf)
	if err != nil {
		return nil, err
	}

	size := binary.LittleEndian.Uint16(headerBuf[0:2])
	opcode := binary.LittleEndian.Uint16(headerBuf[2:4])
	sequence := binary.LittleEndian.Uint32(headerBuf[4:8])

	if size < HeaderSize {
		return nil, fmt.Errorf("packet size %d too small (minimum %d)", size, HeaderSize)
	}
	if size > MaxPacketSize {
		return nil, fmt.Errorf("packet size %d exceeds max %d", size, MaxPacketSize)
	}

	payloadSize := size - HeaderSize
	payload := make([]byte, payloadSize)
	if payloadSize > 0 {
		_, err = io.ReadFull(reader, payload)
		if err != nil {
			return nil, err
		}
	}

	return &Packet{
		Size:     size,
		Opcode:   opcode,
		Sequence: sequence,
		Payload:  payload,
	}, nil
}

// Serialize transforma o pacote em bytes Little Endian prontos para envio
func (p *Packet) Serialize() []byte {
	p.Size = uint16(HeaderSize + len(p.Payload))
	buf := make([]byte, p.Size)
	binary.LittleEndian.PutUint16(buf[0:2], p.Size)
	binary.LittleEndian.PutUint16(buf[2:4], p.Opcode)
	binary.LittleEndian.PutUint32(buf[4:8], p.Sequence)
	if len(p.Payload) > 0 {
		copy(buf[HeaderSize:], p.Payload)
	}
	return buf
}

// Helper to read a length-prefixed string (uint16) from a buffer
func ReadString(buf []byte, offset *int) (string, error) {
	if *offset + 2 > len(buf) {
		return "", fmt.Errorf("buffer overflow reading string length")
	}
	strLen := int(binary.LittleEndian.Uint16(buf[*offset : *offset+2]))
	*offset += 2
	if *offset + strLen > len(buf) {
		return "", fmt.Errorf("buffer overflow reading string payload")
	}
	val := string(buf[*offset : *offset+strLen])
	*offset += strLen
	return val, nil
}

// Helper to write a length-prefixed string to a buffer
func WriteString(buf []byte, val string, offset *int) {
	strLen := len(val)
	binary.LittleEndian.PutUint16(buf[*offset : *offset+2], uint16(strLen))
	*offset += 2
	copy(buf[*offset : *offset+strLen], []byte(val))
	*offset += strLen
}

// Helper to read uint32
func ReadUint32(buf []byte, offset *int) (uint32, error) {
	if *offset + 4 > len(buf) {
		return 0, fmt.Errorf("buffer overflow reading uint32")
	}
	val := binary.LittleEndian.Uint32(buf[*offset : *offset+4])
	*offset += 4
	return val, nil
}

// Helper to write uint32
func WriteUint32(buf []byte, val uint32, offset *int) {
	binary.LittleEndian.PutUint32(buf[*offset : *offset+4], val)
	*offset += 4
}

// Helper to read float64
func ReadFloat64(buf []byte, offset *int) (float64, error) {
	if *offset + 8 > len(buf) {
		return 0, fmt.Errorf("buffer overflow reading float64")
	}
	bits := binary.LittleEndian.Uint64(buf[*offset : *offset+8])
	*offset += 8
	return math.Float64frombits(bits), nil
}

// Helper to write float64
func WriteFloat64(buf []byte, val float64, offset *int) {
	bits := math.Float64bits(val)
	binary.LittleEndian.PutUint64(buf[*offset : *offset+8], bits)
	*offset += 8
}

// Helper to read fixed-point 32-bit float (scaled by 1000)
func ReadFixed32(buf []byte, offset *int) (float64, error) {
	if *offset + 4 > len(buf) {
		return 0, fmt.Errorf("buffer overflow reading fixed32")
	}
	val := int32(binary.LittleEndian.Uint32(buf[*offset : *offset+4]))
	*offset += 4
	return float64(val) / 1000.0, nil
}

// Helper to write fixed-point 32-bit float (scaled by 1000)
func WriteFixed32(buf []byte, val float64, offset *int) {
	fixed := int32(math.Round(val * 1000.0))
	binary.LittleEndian.PutUint32(buf[*offset : *offset+4], uint32(fixed))
	*offset += 4
}

type LoginRequest struct {
    Username string
    Password string
}

func DecodeLoginRequest(payload []byte) (*LoginRequest, error) {
    offset := 0

    username, err := ReadString(payload, &offset)
    if err != nil {
        return nil, err
    }

    password, err := ReadString(payload, &offset)
    if err != nil {
        return nil, err
    }

    return &LoginRequest{
        Username: username,
        Password: password,
    }, nil
}

func EncodeLoginRequest(username, password string) []byte {
    size := 2 + len(username) + 2 + len(password)
    buf := make([]byte, size)
    offset := 0

    WriteString(buf, username, &offset)
    WriteString(buf, password, &offset)

    return buf
}
type AttackRequest struct {
	TargetID   string
	WeaponType string
}

func DecodeAttackRequest(payload []byte) (*AttackRequest, error) {
	offset := 0
	targetID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	weaponType, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &AttackRequest{TargetID: targetID, WeaponType: weaponType}, nil
}

func EncodeAttackRequest(targetID, weaponType string) []byte {
	size := 2 + len(targetID) + 2 + len(weaponType)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, targetID, &offset)
	WriteString(buf, weaponType, &offset)
	return buf
}

type CastSkillRequest struct {
	SkillID  uint32
	TargetID string
	TargetX  float64
	TargetY  float64
}

func DecodeCastSkillRequest(payload []byte) (*CastSkillRequest, error) {
	offset := 0
	skillID, err := ReadUint32(payload, &offset)
	if err != nil {
		return nil, err
	}
	targetID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	targetX, err := ReadFixed32(payload, &offset)
	if err != nil {
		return nil, err
	}
	targetY, err := ReadFixed32(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &CastSkillRequest{
		SkillID:  skillID,
		TargetID: targetID,
		TargetX:  targetX,
		TargetY:  targetY,
	}, nil
}

func EncodeCastSkillRequest(skillID uint32, targetID string, targetX, targetY float64) []byte {
	size := 4 + 2 + len(targetID) + 4 + 4
	buf := make([]byte, size)
	offset := 0
	WriteUint32(buf, skillID, &offset)
	WriteString(buf, targetID, &offset)
	WriteFixed32(buf, targetX, &offset)
	WriteFixed32(buf, targetY, &offset)
	return buf
}

type DamageEvent struct {
	AttackerID string
	TargetID   string
	Damage     float64
	IsCrit     bool
	IsHit      bool
	Success    bool
	SkillName  string
}

func DecodeDamageEvent(payload []byte) (*DamageEvent, error) {
	offset := 0
	attackerID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	targetID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	damage, err := ReadFixed32(payload, &offset)
	if err != nil {
		return nil, err
	}
	if offset+3 > len(payload) {
		return nil, fmt.Errorf("buffer overflow reading booleans in DamageEvent")
	}
	isCrit := payload[offset] != 0
	isHit := payload[offset+1] != 0
	success := payload[offset+2] != 0
	offset += 3

	skillName, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}

	return &DamageEvent{
		AttackerID: attackerID,
		TargetID:   targetID,
		Damage:     damage,
		IsCrit:     isCrit,
		IsHit:      isHit,
		Success:    success,
		SkillName:  skillName,
	}, nil
}

func EncodeDamageEvent(attackerID, targetID string, damage float64, isCrit, isHit, success bool, skillName string) []byte {
	size := 2 + len(attackerID) + 2 + len(targetID) + 4 + 3 + 2 + len(skillName)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, attackerID, &offset)
	WriteString(buf, targetID, &offset)
	WriteFixed32(buf, damage, &offset)
	if isCrit {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
	if isHit {
		buf[offset+1] = 1
	} else {
		buf[offset+1] = 0
	}
	if success {
		buf[offset+2] = 1
	} else {
		buf[offset+2] = 0
	}
	offset += 3
	WriteString(buf, skillName, &offset)
	return buf
}

type TargetDeadEvent struct {
	TargetID string
}

func DecodeTargetDeadEvent(payload []byte) (*TargetDeadEvent, error) {
	offset := 0
	targetID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &TargetDeadEvent{TargetID: targetID}, nil
}

func EncodeTargetDeadEvent(targetID string) []byte {
	size := 2 + len(targetID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, targetID, &offset)
	return buf
}

// Structs de dados de Inventário para Rede
type SyncItem struct {
	ItemID     string
	Quantity   uint32
	Durability uint32
	SlotIndex  uint16
}

type InventorySyncEvent struct {
	Items        []SyncItem
	Level        uint32
	MaxHealth    float64
	Health       float64
	MaxMana      float64
	Mana         float64
	BaseAttack   float64
	WeaponDamage float64
	Defense      float64
	Resistance   float64
	CritChance   float64
}

// EncodeInventorySync serializa o inventário completo e atributos recalculados em formato binário compacto
func EncodeInventorySync(event *InventorySyncEvent) []byte {
	itemsSize := 2
	for _, it := range event.Items {
		itemsSize += 2 + len(it.ItemID) + 4 + 4 + 2
	}
	statsSize := 4 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8
	buf := make([]byte, itemsSize+statsSize)
	offset := 0

	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Items)))
	offset += 2

	for _, it := range event.Items {
		WriteString(buf, it.ItemID, &offset)
		WriteUint32(buf, it.Quantity, &offset)
		WriteUint32(buf, it.Durability, &offset)
		binary.LittleEndian.PutUint16(buf[offset:offset+2], it.SlotIndex)
		offset += 2
	}

	WriteUint32(buf, event.Level, &offset)
	WriteFloat64(buf, event.MaxHealth, &offset)
	WriteFloat64(buf, event.Health, &offset)
	WriteFloat64(buf, event.MaxMana, &offset)
	WriteFloat64(buf, event.Mana, &offset)
	WriteFloat64(buf, event.BaseAttack, &offset)
	WriteFloat64(buf, event.WeaponDamage, &offset)
	WriteFloat64(buf, event.Defense, &offset)
	WriteFloat64(buf, event.Resistance, &offset)
	WriteFloat64(buf, event.CritChance, &offset)

	return buf
}

type EquipItemRequest struct {
	FromSlot uint16
	ToSlot   uint16
}

func DecodeEquipItemRequest(payload []byte) (*EquipItemRequest, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload of EquipItemRequest too short: %d", len(payload))
	}
	fromSlot := binary.LittleEndian.Uint16(payload[0:2])
	toSlot := binary.LittleEndian.Uint16(payload[2:4])
	return &EquipItemRequest{FromSlot: fromSlot, ToSlot: toSlot}, nil
}

type EquipItemResponse struct {
	Success      bool
	ErrorMessage string
}

func EncodeEquipItemResponse(success bool, errMsg string) []byte {
	buf := make([]byte, 1+2+len(errMsg))
	if success {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	offset := 1
	WriteString(buf, errMsg, &offset)
	return buf
}

type UnequipItemRequest struct {
	FromSlot uint16
}

func DecodeUnequipItemRequest(payload []byte) (*UnequipItemRequest, error) {
	if len(payload) < 2 {
		return nil, fmt.Errorf("payload of UnequipItemRequest too short: %d", len(payload))
	}
	fromSlot := binary.LittleEndian.Uint16(payload[0:2])
	return &UnequipItemRequest{FromSlot: fromSlot}, nil
}

type UnequipItemResponse struct {
	Success      bool
	ErrorMessage string
}

func EncodeUnequipItemResponse(success bool, errMsg string) []byte {
	buf := make([]byte, 1+2+len(errMsg))
	if success {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	offset := 1
	WriteString(buf, errMsg, &offset)
	return buf
}

type SwapSlotsRequest struct {
	SlotA uint16
	SlotB uint16
}

func DecodeSwapSlotsRequest(payload []byte) (*SwapSlotsRequest, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload of SwapSlotsRequest too short: %d", len(payload))
	}
	slotA := binary.LittleEndian.Uint16(payload[0:2])
	slotB := binary.LittleEndian.Uint16(payload[2:4])
	return &SwapSlotsRequest{SlotA: slotA, SlotB: slotB}, nil
}

type SwapSlotsResponse struct {
	Success      bool
	ErrorMessage string
}

func EncodeSwapSlotsResponse(success bool, errMsg string) []byte {
	buf := make([]byte, 1+2+len(errMsg))
	if success {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	offset := 1
	WriteString(buf, errMsg, &offset)
	return buf
}

// =============================================================================
// NPC & QUEST SYSTEM PROTOCOLS (Sprint 3 Task 3)
// =============================================================================

type NPCInteractRequest struct {
	NPCID string
}

func DecodeNPCInteractRequest(payload []byte) (*NPCInteractRequest, error) {
	offset := 0
	npcID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &NPCInteractRequest{NPCID: npcID}, nil
}

func EncodeNPCInteractRequest(npcID string) []byte {
	size := 2 + len(npcID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, npcID, &offset)
	return buf
}

type DialogueOpenChoice struct {
	NextNodeID string
	Text       string
}

type DialogueOpenEvent struct {
	NPCID    string
	NodeID   string
	NodeText string
	Choices  []DialogueOpenChoice
}

func EncodeDialogueOpen(event *DialogueOpenEvent) []byte {
	size := 2 + len(event.NPCID) + 2 + len(event.NodeID) + 2 + len(event.NodeText) + 2
	for _, ch := range event.Choices {
		size += 2 + len(ch.NextNodeID) + 2 + len(ch.Text)
	}

	buf := make([]byte, size)
	offset := 0
	WriteString(buf, event.NPCID, &offset)
	WriteString(buf, event.NodeID, &offset)
	WriteString(buf, event.NodeText, &offset)

	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Choices)))
	offset += 2

	for _, ch := range event.Choices {
		WriteString(buf, ch.NextNodeID, &offset)
		WriteString(buf, ch.Text, &offset)
	}

	return buf
}

type DialogueResponseRequest struct {
	NPCID      string
	NodeID     string
	NextNodeID string
}

func DecodeDialogueResponseRequest(payload []byte) (*DialogueResponseRequest, error) {
	offset := 0
	npcID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	nodeID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	nextNodeID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &DialogueResponseRequest{
		NPCID:      npcID,
		NodeID:     nodeID,
		NextNodeID: nextNodeID,
	}, nil
}

func EncodeDialogueResponseRequest(npcID, nodeID, nextNodeID string) []byte {
	size := 2 + len(npcID) + 2 + len(nodeID) + 2 + len(nextNodeID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, npcID, &offset)
	WriteString(buf, nodeID, &offset)
	WriteString(buf, nextNodeID, &offset)
	return buf
}

type AcceptQuestRequest struct {
	QuestID string
}

func DecodeAcceptQuestRequest(payload []byte) (*AcceptQuestRequest, error) {
	offset := 0
	questID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &AcceptQuestRequest{QuestID: questID}, nil
}

func EncodeAcceptQuestRequest(questID string) []byte {
	size := 2 + len(questID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, questID, &offset)
	return buf
}

type CompleteQuestRequest struct {
	QuestID string
}

func DecodeCompleteQuestRequest(payload []byte) (*CompleteQuestRequest, error) {
	offset := 0
	questID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &CompleteQuestRequest{QuestID: questID}, nil
}

func EncodeCompleteQuestRequest(questID string) []byte {
	size := 2 + len(questID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, questID, &offset)
	return buf
}

type ProtocolObjectiveState struct {
	Index      uint16
	CurrentQty uint32
}

type QuestUpdateEvent struct {
	QuestID    string
	Status     string
	Objectives []ProtocolObjectiveState
}

func EncodeQuestUpdate(event *QuestUpdateEvent) []byte {
	size := 2 + len(event.QuestID) + 2 + len(event.Status) + 2 + len(event.Objectives)*6
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, event.QuestID, &offset)
	WriteString(buf, event.Status, &offset)

	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Objectives)))
	offset += 2

	for _, obj := range event.Objectives {
		binary.LittleEndian.PutUint16(buf[offset:offset+2], obj.Index)
		offset += 2
		binary.LittleEndian.PutUint32(buf[offset:offset+4], obj.CurrentQty)
		offset += 4
	}

	return buf
}

type QuestCompleteEvent struct {
	QuestID string
}

func EncodeQuestComplete(questID string) []byte {
	size := 2 + len(questID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, questID, &offset)
	return buf
}

// =============================================================================
// SOCIAL, PARTY, GUILD & CHAT SYSTEMS PROTOCOLS (Sprint 3 Task 3)
// =============================================================================

// Party Member representation in binary
type PartyMemberInfo struct {
	PlayerID string
	Online   uint8
}

type PartyInfoEvent struct {
	LootMode uint8
	LeaderID string
	Members  []PartyMemberInfo
}

func EncodePartyInfo(event *PartyInfoEvent) []byte {
	size := 1 + 2 + len(event.LeaderID) + 2
	for _, m := range event.Members {
		size += 2 + len(m.PlayerID) + 1
	}
	buf := make([]byte, size)
	offset := 0
	buf[offset] = event.LootMode
	offset += 1
	WriteString(buf, event.LeaderID, &offset)
	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Members)))
	offset += 2
	for _, m := range event.Members {
		WriteString(buf, m.PlayerID, &offset)
		buf[offset] = m.Online
		offset += 1
	}
	return buf
}

type PartyInviteResponse struct {
	InviterID string
	Accept    uint8
}

func DecodePartyInviteResponse(payload []byte) (*PartyInviteResponse, error) {
	if len(payload) < 3 {
		return nil, fmt.Errorf("payload too short for PartyInviteResponse")
	}
	offset := 0
	inviterID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	if offset >= len(payload) {
		return nil, fmt.Errorf("payload end reached before accept flag")
	}
	accept := payload[offset]
	return &PartyInviteResponse{InviterID: inviterID, Accept: accept}, nil
}

func EncodePartyInviteReq(inviterID string) []byte {
	size := 2 + len(inviterID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, inviterID, &offset)
	return buf
}

// Guild Member representation in binary
type GuildMemberInfo struct {
	PlayerID string
	Role     uint8
	Online   uint8
}

type GuildInfoEvent struct {
	GuildName string
	MOTD      string
	Version   uint32
	Members   []GuildMemberInfo
}

func EncodeGuildInfo(event *GuildInfoEvent) []byte {
	size := 2 + len(event.GuildName) + 2 + len(event.MOTD) + 4 + 2
	for _, m := range event.Members {
		size += 2 + len(m.PlayerID) + 1 + 1
	}
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, event.GuildName, &offset)
	WriteString(buf, event.MOTD, &offset)
	WriteUint32(buf, event.Version, &offset)
	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Members)))
	offset += 2
	for _, m := range event.Members {
		WriteString(buf, m.PlayerID, &offset)
		buf[offset] = m.Role
		offset += 1
		buf[offset] = m.Online
		offset += 1
	}
	return buf
}

type GuildInviteResponse struct {
	GuildName string
	Accept    uint8
}

func DecodeGuildInviteResponse(payload []byte) (*GuildInviteResponse, error) {
	if len(payload) < 3 {
		return nil, fmt.Errorf("payload too short for GuildInviteResponse")
	}
	offset := 0
	guildName, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	if offset >= len(payload) {
		return nil, fmt.Errorf("payload end reached before accept flag")
	}
	accept := payload[offset]
	return &GuildInviteResponse{GuildName: guildName, Accept: accept}, nil
}

func EncodeGuildInviteReq(guildName, inviterID string) []byte {
	size := 2 + len(guildName) + 2 + len(inviterID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, guildName, &offset)
	WriteString(buf, inviterID, &offset)
	return buf
}

type GuildMOTDRequest struct {
	MOTD            string
	ExpectedVersion uint32
}

func DecodeGuildMOTDRequest(payload []byte) (*GuildMOTDRequest, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short for GuildMOTDRequest")
	}
	offset := 0
	motd, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	expectedVersion, err := ReadUint32(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &GuildMOTDRequest{MOTD: motd, ExpectedVersion: expectedVersion}, nil
}

type AuditLogEntry struct {
	Action    string
	Timestamp string
}

type GuildAuditLogEvent struct {
	Entries []AuditLogEntry
}

func EncodeGuildAuditLog(event *GuildAuditLogEvent) []byte {
	size := 2
	for _, entry := range event.Entries {
		size += 2 + len(entry.Action) + 2 + len(entry.Timestamp)
	}
	buf := make([]byte, size)
	offset := 0
	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Entries)))
	offset += 2
	for _, entry := range event.Entries {
		WriteString(buf, entry.Action, &offset)
		WriteString(buf, entry.Timestamp, &offset)
	}
	return buf
}

type FriendInfo struct {
	FriendID string
	Online   uint8
}

type SocialListsEvent struct {
	Friends []FriendInfo
	Ignores []string
}

func EncodeSocialLists(event *SocialListsEvent) []byte {
	size := 2
	for _, f := range event.Friends {
		size += 2 + len(f.FriendID) + 1
	}
	size += 2
	for _, ig := range event.Ignores {
		size += 2 + len(ig)
	}
	buf := make([]byte, size)
	offset := 0

	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Friends)))
	offset += 2
	for _, f := range event.Friends {
		WriteString(buf, f.FriendID, &offset)
		buf[offset] = f.Online
		offset += 1
	}

	binary.LittleEndian.PutUint16(buf[offset:offset+2], uint16(len(event.Ignores)))
	offset += 2
	for _, ig := range event.Ignores {
		WriteString(buf, ig, &offset)
	}
	return buf
}

type OnlineStatusEvent struct {
	PlayerID string
	Online   uint8
}

func EncodeOnlineStatus(playerID string, online uint8) []byte {
	size := 2 + len(playerID) + 1
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, playerID, &offset)
	buf[offset] = online
	return buf
}

type ChatSendRequest struct {
	Channel uint8
	Target  string
	Message string
}

func DecodeChatSendRequest(payload []byte) (*ChatSendRequest, error) {
	if len(payload) < 5 {
		return nil, fmt.Errorf("payload too short for ChatSendRequest")
	}
	channel := payload[0]
	offset := 1
	target, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	message, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &ChatSendRequest{Channel: channel, Target: target, Message: message}, nil
}

type ChatMessageEvent struct {
	Channel uint8
	Sender  string
	Message string
}

func EncodeChatMessage(channel uint8, sender, message string) []byte {
	size := 1 + 2 + len(sender) + 2 + len(message)
	buf := make([]byte, size)
	offset := 0
	buf[offset] = channel
	offset += 1
	WriteString(buf, sender, &offset)
	WriteString(buf, message, &offset)
	return buf
}

// =========================================================================
// ECONOMY, PLAYER TRADING, NPC SHOP & MARKETPLACE BINARY CODECS (Sprint 3 Task 4)
// =========================================================================

// TradeRequest
type TradeRequest struct {
	TargetName string
}

func DecodeTradeRequest(payload []byte) (*TradeRequest, error) {
	offset := 0
	target, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &TradeRequest{TargetName: target}, nil
}

func EncodeTradeProposal(inviterID string) []byte {
	size := 2 + len(inviterID)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, inviterID, &offset)
	return buf
}

// TradeRespond
type TradeRespond struct {
	Accepted uint8 // 1 = accept, 0 = reject
}

func DecodeTradeRespond(payload []byte) (*TradeRespond, error) {
	if len(payload) < 1 {
		return nil, fmt.Errorf("payload too short for TradeRespond")
	}
	return &TradeRespond{Accepted: payload[0]}, nil
}

// TradeOfferItem
type TradeOfferItem struct {
	SlotIndex uint32
	Quantity  uint32
}

func DecodeTradeOfferItem(payload []byte) (*TradeOfferItem, error) {
	if len(payload) < 8 {
		return nil, fmt.Errorf("payload too short for TradeOfferItem")
	}
	offset := 0
	slot, _ := ReadUint32(payload, &offset)
	qty, _ := ReadUint32(payload, &offset)
	return &TradeOfferItem{SlotIndex: slot, Quantity: qty}, nil
}

// TradeOfferGold
type TradeOfferGold struct {
	Gold uint32
}

func DecodeTradeOfferGold(payload []byte) (*TradeOfferGold, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short for TradeOfferGold")
	}
	offset := 0
	gold, _ := ReadUint32(payload, &offset)
	return &TradeOfferGold{Gold: gold}, nil
}

// TradeItemCodec representa um item serializado dentro do update de troca
type TradeItemCodec struct {
	SlotIndex uint32
	ItemID    string
	Quantity  uint32
	ItemUUID  string
}

// TradeUpdateEvent carrega o estado de sincronização da tela de troca (Dual Confirm)
type TradeUpdateEvent struct {
	GoldA      uint32
	GoldB      uint32
	LockedA    uint8
	LockedB    uint8
	AcceptedA  uint8
	AcceptedB  uint8
	ItemsA     []TradeItemCodec
	ItemsB     []TradeItemCodec
}

func EncodeTradeUpdate(event *TradeUpdateEvent) []byte {
	// Calcula tamanho total dinamicamente
	size := 4 + 4 + 1 + 1 + 1 + 1 + 2 + 2 // goldA, goldB, lockedA, lockedB, acceptedA, acceptedB, itemsCountA, itemsCountB
	for _, item := range event.ItemsA {
		size += 4 + 2 + len(item.ItemID) + 4 + 2 + len(item.ItemUUID)
	}
	for _, item := range event.ItemsB {
		size += 4 + 2 + len(item.ItemID) + 4 + 2 + len(item.ItemUUID)
	}

	buf := make([]byte, size)
	offset := 0

	WriteUint32(buf, event.GoldA, &offset)
	WriteUint32(buf, event.GoldB, &offset)
	buf[offset] = event.LockedA
	offset += 1
	buf[offset] = event.LockedB
	offset += 1
	buf[offset] = event.AcceptedA
	offset += 1
	buf[offset] = event.AcceptedB
	offset += 1

	// Escreve quantidades de itens como uint16
	binary.LittleEndian.PutUint16(buf[offset:], uint16(len(event.ItemsA)))
	offset += 2
	binary.LittleEndian.PutUint16(buf[offset:], uint16(len(event.ItemsB)))
	offset += 2

	// Escreve itens de A
	for _, item := range event.ItemsA {
		WriteUint32(buf, item.SlotIndex, &offset)
		WriteString(buf, item.ItemID, &offset)
		WriteUint32(buf, item.Quantity, &offset)
		WriteString(buf, item.ItemUUID, &offset)
	}

	// Escreve itens de B
	for _, item := range event.ItemsB {
		WriteUint32(buf, item.SlotIndex, &offset)
		WriteString(buf, item.ItemID, &offset)
		WriteUint32(buf, item.Quantity, &offset)
		WriteString(buf, item.ItemUUID, &offset)
	}

	return buf
}

func EncodeTradeClosed(reason string) []byte {
	size := 2 + len(reason)
	buf := make([]byte, size)
	offset := 0
	WriteString(buf, reason, &offset)
	return buf
}

// NPCShopBuy
type NPCShopBuy struct {
	ItemID   string
	Quantity uint32
}

func DecodeNPCShopBuy(payload []byte) (*NPCShopBuy, error) {
	if len(payload) < 6 {
		return nil, fmt.Errorf("payload too short for NPCShopBuy")
	}
	offset := 0
	itemID, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	qty, _ := ReadUint32(payload, &offset)
	return &NPCShopBuy{ItemID: itemID, Quantity: qty}, nil
}

// NPCShopSell
type NPCShopSell struct {
	SlotIndex uint32
	Quantity  uint32
}

func DecodeNPCShopSell(payload []byte) (*NPCShopSell, error) {
	if len(payload) < 8 {
		return nil, fmt.Errorf("payload too short for NPCShopSell")
	}
	offset := 0
	slot, _ := ReadUint32(payload, &offset)
	qty, _ := ReadUint32(payload, &offset)
	return &NPCShopSell{SlotIndex: slot, Quantity: qty}, nil
}

// NPCShopRepair
type NPCShopRepair struct {
	SlotIndex uint32
}

func DecodeNPCShopRepair(payload []byte) (*NPCShopRepair, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short for NPCShopRepair")
	}
	offset := 0
	slot, _ := ReadUint32(payload, &offset)
	return &NPCShopRepair{SlotIndex: slot}, nil
}

func EncodeNPCShopResponse(success uint8, message string) []byte {
	size := 1 + 2 + len(message)
	buf := make([]byte, size)
	offset := 0
	buf[offset] = success
	offset += 1
	WriteString(buf, message, &offset)
	return buf
}

// MarketCreateOrder
type MarketCreateOrder struct {
	SlotIndex uint32
	Quantity  uint32
	Price     uint32
}

func DecodeMarketCreateOrder(payload []byte) (*MarketCreateOrder, error) {
	if len(payload) < 12 {
		return nil, fmt.Errorf("payload too short for MarketCreateOrder")
	}
	offset := 0
	slot, _ := ReadUint32(payload, &offset)
	qty, _ := ReadUint32(payload, &offset)
	price, _ := ReadUint32(payload, &offset)
	return &MarketCreateOrder{SlotIndex: slot, Quantity: qty, Price: price}, nil
}

// MarketBuyItem
type MarketBuyItem struct {
	OrderID uint32
}

func DecodeMarketBuyItem(payload []byte) (*MarketBuyItem, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short for MarketBuyItem")
	}
	offset := 0
	orderID, _ := ReadUint32(payload, &offset)
	return &MarketBuyItem{OrderID: orderID}, nil
}

// MarketCancelOrder
type MarketCancelOrder struct {
	OrderID uint32
}

func DecodeMarketCancelOrder(payload []byte) (*MarketCancelOrder, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("payload too short for MarketCancelOrder")
	}
	offset := 0
	orderID, _ := ReadUint32(payload, &offset)
	return &MarketCancelOrder{OrderID: orderID}, nil
}

// MarketSearch
type MarketSearch struct {
	FilterItemID string
}

func DecodeMarketSearch(payload []byte) (*MarketSearch, error) {
	offset := 0
	filter, err := ReadString(payload, &offset)
	if err != nil {
		return nil, err
	}
	return &MarketSearch{FilterItemID: filter}, nil
}

// MarketOrderCodec para serialização de ordens
type MarketOrderCodec struct {
	OrderID       uint32
	SellerName    string
	ItemID        string
	Quantity      uint32
	PriceGold     uint32
	ExpiresEpoch  uint32
}

func EncodeMarketSearchResult(orders []MarketOrderCodec) []byte {
	size := 2 // Quantidade de ordens
	for _, o := range orders {
		size += 4 + 2 + len(o.SellerName) + 2 + len(o.ItemID) + 4 + 4 + 4
	}

	buf := make([]byte, size)
	offset := 0

	binary.LittleEndian.PutUint16(buf[offset:], uint16(len(orders)))
	offset += 2

	for _, o := range orders {
		WriteUint32(buf, o.OrderID, &offset)
		WriteString(buf, o.SellerName, &offset)
		WriteString(buf, o.ItemID, &offset)
		WriteUint32(buf, o.Quantity, &offset)
		WriteUint32(buf, o.PriceGold, &offset)
		WriteUint32(buf, o.ExpiresEpoch, &offset)
	}

	return buf
}

// Decoders and Encoders for Gathering & Crafting System (Sprint 4 Task 1)

func DecodeGatherStart(payload []byte) (string, error) {
	if len(payload) < 2 {
		return "", fmt.Errorf("payload too short for DecodeGatherStart")
	}
	offset := 0
	return ReadString(payload, &offset)
}

func EncodeGatherProgress(nodeID string, duration float64) []byte {
	buf := make([]byte, 2+len(nodeID)+8)
	offset := 0
	WriteString(buf, nodeID, &offset)
	bits := math.Float64bits(duration)
	binary.LittleEndian.PutUint64(buf[offset:], bits)
	return buf
}

func EncodeGatherComplete(nodeID string, itemID string, xp uint32, success uint8) []byte {
	buf := make([]byte, 2+len(nodeID)+2+len(itemID)+4+1)
	offset := 0
	WriteString(buf, nodeID, &offset)
	WriteString(buf, itemID, &offset)
	WriteUint32(buf, xp, &offset)
	buf[offset] = success
	return buf
}

func DecodeCraftStart(payload []byte) (string, error) {
	if len(payload) < 2 {
		return "", fmt.Errorf("payload too short for DecodeCraftStart")
	}
	offset := 0
	return ReadString(payload, &offset)
}

func EncodeCraftResponse(recipeID string, outputItemID string, xp uint32, success uint8) []byte {
	buf := make([]byte, 2+len(recipeID)+2+len(outputItemID)+4+1)
	offset := 0
	WriteString(buf, recipeID, &offset)
	WriteString(buf, outputItemID, &offset)
	WriteUint32(buf, xp, &offset)
	buf[offset] = success
	return buf
}

func EncodeProfessionXPUpdate(profession string, level uint32, experience uint32) []byte {
	buf := make([]byte, 2+len(profession)+4+4)
	offset := 0
	WriteString(buf, profession, &offset)
	WriteUint32(buf, level, &offset)
	WriteUint32(buf, experience, &offset)
	return buf
}

// Dungeon, Raid & World Boss Encoders and Decoders (Sprint 4 Task 2)

func DecodeDungeonEnter(payload []byte) (string, string, error) {
	if len(payload) < 4 {
		return "", "", fmt.Errorf("payload too short for DecodeDungeonEnter")
	}
	offset := 0
	dungeonID, err := ReadString(payload, &offset)
	if err != nil {
		return "", "", err
	}
	mode, err := ReadString(payload, &offset)
	if err != nil {
		return "", "", err
	}
	return dungeonID, mode, nil
}

func EncodeDungeonState(instanceID string, dungeonID string, checkpoint string, timeLeft float64, bossAlive uint8) []byte {
	buf := make([]byte, 2+len(instanceID)+2+len(dungeonID)+2+len(checkpoint)+8+1)
	offset := 0
	WriteString(buf, instanceID, &offset)
	WriteString(buf, dungeonID, &offset)
	WriteString(buf, checkpoint, &offset)
	bits := math.Float64bits(timeLeft)
	binary.LittleEndian.PutUint64(buf[offset:], bits)
	offset += 8
	buf[offset] = bossAlive
	return buf
}

func EncodeBossAITelegraph(bossID string, x, y, radius, delay float64) []byte {
	buf := make([]byte, 2+len(bossID)+8+8+8+8)
	offset := 0
	WriteString(buf, bossID, &offset)
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(x))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(y))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(radius))
	offset += 8
	binary.LittleEndian.PutUint64(buf[offset:], math.Float64bits(delay))
	offset += 8
	return buf
}

func EncodeBossAIPhase(bossID string, phase uint32) []byte {
	buf := make([]byte, 2+len(bossID)+4)
	offset := 0
	WriteString(buf, bossID, &offset)
	WriteUint32(buf, phase, &offset)
	return buf
}

func DecodeDungeonClaimLoot(payload []byte) (string, string, error) {
	if len(payload) < 4 {
		return "", "", fmt.Errorf("payload too short for DecodeDungeonClaimLoot")
	}
	offset := 0
	instanceID, err := ReadString(payload, &offset)
	if err != nil {
		return "", "", err
	}
	itemID, err := ReadString(payload, &offset)
	if err != nil {
		return "", "", err
	}
	return instanceID, itemID, nil
}

func EncodeLootNotification(itemID string, qty uint32, rank uint32, canClaim uint8) []byte {
	buf := make([]byte, 2+len(itemID)+4+4+1)
	offset := 0
	WriteString(buf, itemID, &offset)
	WriteUint32(buf, qty, &offset)
	WriteUint32(buf, rank, &offset)
	buf[offset] = canClaim
	return buf
}

// Decodifica a escolha de vocação enviada pelo cliente (Sprint 3 Task 5)
func DecodeChooseVocationRequest(payload []byte) (string, error) {
	if len(payload) < 2 {
		return "", fmt.Errorf("choose vocation payload too short")
	}
	length := binary.BigEndian.Uint16(payload[0:2])
	if len(payload) < 2+int(length) {
		return "", fmt.Errorf("choose vocation payload truncated")
	}
	return string(payload[2 : 2+length]), nil
}

// Codifica a resposta de escolha de vocação (Sprint 3 Task 5)
func EncodeChooseVocationResponse(success bool, errorMessage string, class string) []byte {
	payload := make([]byte, 3)
	if success {
		payload[0] = 1
	} else {
		payload[0] = 0
	}
	
	errBytes := []byte(errorMessage)
	binary.BigEndian.PutUint16(payload[1:3], uint16(len(errBytes)))
	payload = append(payload, errBytes...)

	classBytes := []byte(class)
	classLenBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(classLenBytes, uint16(len(classBytes)))
	payload = append(payload, classLenBytes...)
	payload = append(payload, classBytes...)

	return payload
}

// Codifica a resposta de desbloqueio de subclass (Sprint 3 Task 5)
func EncodeUnlockSubclassResponse(success bool, errorMessage string, subclass string, element string) []byte {
	payload := make([]byte, 3)
	if success {
		payload[0] = 1
	} else {
		payload[0] = 0
	}

	errBytes := []byte(errorMessage)
	binary.BigEndian.PutUint16(payload[1:3], uint16(len(errBytes)))
	payload = append(payload, errBytes...)

	subBytes := []byte(subclass)
	subLen := make([]byte, 2)
	binary.BigEndian.PutUint16(subLen, uint16(len(subBytes)))
	payload = append(payload, subLen...)
	payload = append(payload, subBytes...)

	elBytes := []byte(element)
	elLen := make([]byte, 2)
	binary.BigEndian.PutUint16(elLen, uint16(len(elBytes)))
	payload = append(payload, elLen...)
	payload = append(payload, elBytes...)

	return payload
}

type LoginResponse struct {
    Success   bool
    AccountID uint32
    Token     string
    ErrorCode string
}

func EncodeLoginResponse(success bool, accountID uint32, token string, errorCode string) []byte {
    status := byte(0)
    if success {
        status = 1
    }

    payload := []byte{status}

    var accountBuf [4]byte
    binary.LittleEndian.PutUint32(accountBuf[:], accountID)
    payload = append(payload, accountBuf[:]...)

    payload = appendLengthPrefixedProtocolString(payload, token)
    payload = appendLengthPrefixedProtocolString(payload, errorCode)

    return payload
}

func DecodeLoginResponse(payload []byte) (*LoginResponse, error) {
    if len(payload) < 5 {
        return nil, fmt.Errorf("login response payload too small: %d", len(payload))
    }

    status := payload[0]
    accountID := binary.LittleEndian.Uint32(payload[1:5])
    offset := 5

    token, nextOffset, err := readLengthPrefixedProtocolString(payload, offset)
    if err != nil {
        return nil, err
    }
    offset = nextOffset

    errorCode, _, err := readLengthPrefixedProtocolString(payload, offset)
    if err != nil {
        return nil, err
    }

    return &LoginResponse{
        Success:   status == 1,
        AccountID: accountID,
        Token:     token,
        ErrorCode: errorCode,
    }, nil
}

func appendLengthPrefixedProtocolString(payload []byte, value string) []byte {
    raw := []byte(value)
    if len(raw) > 65535 {
        raw = raw[:65535]
    }

    var lenBuf [2]byte
    binary.LittleEndian.PutUint16(lenBuf[:], uint16(len(raw)))

    payload = append(payload, lenBuf[:]...)
    payload = append(payload, raw...)

    return payload
}

func readLengthPrefixedProtocolString(payload []byte, offset int) (string, int, error) {
    if offset+2 > len(payload) {
        return "", offset, fmt.Errorf("not enough bytes to read string length")
    }

    length := int(binary.LittleEndian.Uint16(payload[offset : offset+2]))
    offset += 2

    if offset+length > len(payload) {
        return "", offset, fmt.Errorf("not enough bytes to read string payload")
    }

    value := string(payload[offset : offset+length])
    offset += length

    return value, offset, nil
}
