using System;
using System.Collections.Generic;
using System.IO;
using System.Text.Json;
using System.Text;

namespace LightAndShadow.Client;

public static class BinaryProtocol
{
    public static byte[] EncodeLoginRequest(string username, string password)
    {
        var payload = new byte[4 + Encoding.UTF8.GetByteCount(username) + Encoding.UTF8.GetByteCount(password)];
        var offset = 0;
        offset = WriteStringUInt16(payload, offset, username);
        offset = WriteStringUInt16(payload, offset, password);
        return payload;
    }

    public static LoginResponseData DecodeLoginResponse(byte[] payload)
    {
        var offset = 0;
        if (payload.Length < 5)
        {
            throw new InvalidDataException("Login response payload is too small.");
        }

        var status = payload[offset++];
        var accountId = ReadUInt32LE(payload, offset);
        offset += 4;
        var token = ReadStringUInt16(payload, offset, out offset);
        var errorCode = ReadStringUInt16(payload, offset, out _);

        return new LoginResponseData
        {
            Status = status == 1,
            AccountId = accountId,
            Token = token,
            ErrorCode = errorCode
        };
    }

    public static byte[] EncodeCharacterListRequest()
    {
        return Array.Empty<byte>();
    }

    public static CharacterListResponseData DecodeCharacterListResponse(byte[] payload)
    {
        if (payload.Length < 3)
        {
            throw new InvalidDataException("Character list response payload is too small.");
        }

        var offset = 0;
        var status = payload[offset++];
        var errorCode = ReadStringUInt16(payload, offset, out offset);
        var count = ReadUInt16LE(payload, offset);
        offset += 2;

        var characters = new List<CharacterListEntryData>();
        for (var index = 0; index < count; index++)
        {
            var name = ReadStringUInt16(payload, offset, out offset);
            var className = ReadStringUInt16(payload, offset, out offset);
            var level = ReadUInt32LE(payload, offset);
            offset += 4;
            var raceId = ReadStringUInt16(payload, offset, out offset); // (R1-I-B)

            characters.Add(new CharacterListEntryData
            {
                Name = name,
                Class = className,
                Level = level,
                RaceId = raceId, // (R1-I-B)
            });
        }

        return new CharacterListResponseData
        {
            Status = status == 1,
            ErrorCode = errorCode,
            Characters = characters
        };
    }

    public static byte[] EncodeCharacterSelectRequest(string characterName)
    {
        var payload = new byte[2 + Encoding.UTF8.GetByteCount(characterName)];
        var offset = 0;
        offset = WriteStringUInt16(payload, offset, characterName);
        return payload;
    }

    public static CharacterSelectResponseData DecodeCharacterSelectResponse(byte[] payload)
    {
        if (payload.Length < 5)
        {
            throw new InvalidDataException("Character select response payload is too small.");
        }

        var offset = 0;
        var status = payload[offset++];
        var characterName = ReadStringUInt16(payload, offset, out offset);
        var errorCode = ReadStringUInt16(payload, offset, out _);

        return new CharacterSelectResponseData
        {
            Status = status == 1,
            CharacterName = characterName,
            ErrorCode = errorCode
        };
    }

    public static byte[] EncodeCharacterCreateRequest(string desiredName, string raceId)
    {
        var safeDesiredName = desiredName ?? string.Empty;
        var safeRaceId = raceId ?? string.Empty;

        var payload = new byte[2 + Encoding.UTF8.GetByteCount(safeDesiredName) + 2 + Encoding.UTF8.GetByteCount(safeRaceId)];
        var offset = 0;
        offset = WriteStringUInt16(payload, offset, safeDesiredName);
        offset = WriteStringUInt16(payload, offset, safeRaceId);
        return payload;
    }

    public static CharacterCreateResponseData DecodeCharacterCreateResponse(byte[] payload)
    {
        if (payload.Length < 1)
        {
            throw new InvalidDataException("Character create response payload is too small.");
        }

        var offset = 0;
        var status = payload[offset++];
        var errorCode = ReadStringUInt16(payload, offset, out offset);
        var name = ReadStringUInt16(payload, offset, out offset);
        var className = ReadStringUInt16(payload, offset, out offset);

        if (offset + 4 > payload.Length)
        {
            throw new InvalidDataException("Payload truncated before reading character level.");
        }
        var level = ReadUInt32LE(payload, offset);
        offset += 4;

        var raceId = ReadStringUInt16(payload, offset, out _);

        return new CharacterCreateResponseData
        {
            Status = status == 1,
            ErrorCode = errorCode,
            Character = new CharacterListEntryData
            {
                Name = name,
                Class = className,
                Level = level,
                RaceId = raceId
            }
        };
    }

    public static byte[] EncodeMoveRequest(int targetX, int targetY, sbyte targetZ, byte direction, ulong clientTimestamp)
    {
        var payload = new byte[18];
        var offset = 0;
        
        WriteInt32LE(payload, offset, targetX);
        offset += 4;
        
        WriteInt32LE(payload, offset, targetY);
        offset += 4;

        payload[offset] = (byte)targetZ;
        offset += 1;

        payload[offset] = direction;
        offset += 1;

        WriteUInt64LE(payload, offset, clientTimestamp);
        
        return payload;
    }

    public static MoveConfirmData? DecodeMoveConfirm(byte[] payload)
    {
        return JsonSerializer.Deserialize<MoveConfirmData>(payload);
    }

    public static PlayerUpdateData? DecodePlayerUpdate(byte[] payload)
    {
        return JsonSerializer.Deserialize<PlayerUpdateData>(payload);
    }

    public static InventorySyncData DecodeInventorySync(byte[] payload)
    {
        var offset = 0;
        var itemCount = ReadUInt16LE(payload, offset);
        offset += 2;

        var items = new List<InventoryItemData>();
        for (var i = 0; i < itemCount; i++)
        {
            var itemId = ReadStringUInt16(payload, offset, out offset);
            var quantity = ReadUInt32LE(payload, offset);
            offset += 4;
            var durability = ReadUInt32LE(payload, offset);
            offset += 4;
            var slotIndex = ReadUInt16LE(payload, offset);
            offset += 2;
            items.Add(new InventoryItemData { ItemId = itemId, Quantity = quantity, Durability = durability, SlotIndex = slotIndex });
        }

        return new InventorySyncData
        {
            Items = items,
            Level = ReadUInt32LE(payload, offset, out offset),
            MaxHealth = ReadFloat64LE(payload, offset, out offset),
            Health = ReadFloat64LE(payload, offset, out offset),
            MaxMana = ReadFloat64LE(payload, offset, out offset),
            Mana = ReadFloat64LE(payload, offset, out offset),
            BaseAttack = ReadFloat64LE(payload, offset, out offset),
            WeaponDamage = ReadFloat64LE(payload, offset, out offset),
            Defense = ReadFloat64LE(payload, offset, out offset),
            Resistance = ReadFloat64LE(payload, offset, out offset),
            CritChance = ReadFloat64LE(payload, offset, out offset)
        };
    }

    public static ChunkData DecodeChunkData(byte[] payload)
    {
        var offset = 0;
        var chunkX = ReadUInt32LE(payload, offset);
        offset += 4;
        var chunkY = ReadUInt32LE(payload, offset);
        offset += 4;

        var tiles = new byte[1024];
        Buffer.BlockCopy(payload, offset, tiles, 0, 1024);

        return new ChunkData
        {
            ChunkX = chunkX,
            ChunkY = chunkY,
            Tiles = tiles
        };
    }

    public static byte[] EncodeAttackRequest(string targetId, string weaponType)
    {
        var safeTargetId = targetId ?? string.Empty;
        var safeWeaponType = weaponType ?? string.Empty;

        var payload = new byte[2 + Encoding.UTF8.GetByteCount(safeTargetId) + 2 + Encoding.UTF8.GetByteCount(safeWeaponType)];
        var offset = 0;
        offset = WriteStringUInt16(payload, offset, safeTargetId);
        offset = WriteStringUInt16(payload, offset, safeWeaponType);
        return payload;
    }

    public static byte[] EncodeCastSkillRequest(uint skillId, string targetId, double targetX, double targetY)
    {
        var safeTargetId = targetId ?? string.Empty;
        var targetIdByteCount = Encoding.UTF8.GetByteCount(safeTargetId);

        // 4 (skillId) + 2 (str len) + N (str bytes) + 4 (targetX) + 4 (targetY)
        var payload = new byte[4 + 2 + targetIdByteCount + 4 + 4];
        var offset = 0;

        WriteUInt32LE(payload, offset, skillId);
        offset += 4;

        offset = WriteStringUInt16(payload, offset, safeTargetId);

        WriteFixed32LE(payload, offset, targetX);
        offset += 4;

        WriteFixed32LE(payload, offset, targetY);

        return payload;
    }

    public static DamageEventData DecodeDamageEvent(byte[] payload)
    {
        var offset = 0;
        var attackerId = ReadStringUInt16(payload, offset, out offset);
        var targetId = ReadStringUInt16(payload, offset, out offset);
        var damage = ReadFixed32LE(payload, offset, out offset);
        if (offset + 3 > payload.Length)
        {
            throw new InvalidDataException("Buffer overflow while reading damage event booleans.");
        }
        var isCrit = payload[offset++] != 0;
        var isHit = payload[offset++] != 0;
        var success = payload[offset++] != 0;
        var skillName = ReadStringUInt16(payload, offset, out _);

        return new DamageEventData
        {
            AttackerID = attackerId,
            TargetID = targetId,
            Damage = damage,
            IsCrit = isCrit,
            IsHit = isHit,
            Success = success,
            SkillName = skillName
        };
    }

    public static TargetDeadEventData DecodeTargetDeadEvent(byte[] payload)
    {
        return new TargetDeadEventData { TargetID = ReadStringUInt16(payload, 0, out _) };
    }

    public static void WriteUInt16LE(byte[] buffer, int offset, ushort value)
    {
        if (offset + 2 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while writing uint16.");
        }

        buffer[offset] = (byte)(value & 0xFF);
        buffer[offset + 1] = (byte)((value >> 8) & 0xFF);
    }

    public static ushort ReadUInt16LE(byte[] buffer, int offset)
    {
        if (offset + 2 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while reading uint16.");
        }

        return (ushort)(buffer[offset] | (buffer[offset + 1] << 8));
    }

    public static void WriteInt32LE(byte[] buffer, int offset, int value)
    {
        if (offset + 4 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while writing int32.");
        }

        buffer[offset] = (byte)(value & 0xFF);
        buffer[offset + 1] = (byte)((value >> 8) & 0xFF);
        buffer[offset + 2] = (byte)((value >> 16) & 0xFF);
        buffer[offset + 3] = (byte)((value >> 24) & 0xFF);
    }

    public static void WriteUInt32LE(byte[] buffer, int offset, uint value)
    {
        if (offset + 4 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while writing uint32.");
        }

        buffer[offset] = (byte)(value & 0xFF);
        buffer[offset + 1] = (byte)((value >> 8) & 0xFF);
        buffer[offset + 2] = (byte)((value >> 16) & 0xFF);
        buffer[offset + 3] = (byte)((value >> 24) & 0xFF);
    }

    public static uint ReadUInt32LE(byte[] buffer, int offset)
    {
        if (offset + 4 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while reading uint32.");
        }

        return (uint)(buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24));
    }

    public static uint ReadUInt32LE(byte[] buffer, int offset, out int nextOffset)
    {
        nextOffset = offset + 4;
        return ReadUInt32LE(buffer, offset);
    }

    public static void WriteUInt64LE(byte[] buffer, int offset, ulong value)
    {
        if (offset + 8 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while writing uint64.");
        }
        buffer[offset] = (byte)(value & 0xFF);
        buffer[offset + 1] = (byte)((value >> 8) & 0xFF);
        buffer[offset + 2] = (byte)((value >> 16) & 0xFF);
        buffer[offset + 3] = (byte)((value >> 24) & 0xFF);
        buffer[offset + 4] = (byte)((value >> 32) & 0xFF);
        buffer[offset + 5] = (byte)((value >> 40) & 0xFF);
        buffer[offset + 6] = (byte)((value >> 48) & 0xFF);
        buffer[offset + 7] = (byte)((value >> 56) & 0xFF);
    }

    public static int WriteStringUInt16(byte[] buffer, int offset, string value)
    {
        var bytes = Encoding.UTF8.GetBytes(value ?? string.Empty);
        if (bytes.Length > 65535)
        {
            throw new ArgumentOutOfRangeException(nameof(value), "String exceeds the maximum supported size.");
        }

        WriteUInt16LE(buffer, offset, (ushort)bytes.Length);
        Buffer.BlockCopy(bytes, 0, buffer, offset + 2, bytes.Length);
        return offset + 2 + bytes.Length;
    }

    public static string ReadStringUInt16(byte[] buffer, int offset, out int nextOffset)
    {
        if (offset + 2 > buffer.Length)
        {
            throw new InvalidDataException("String length prefix is truncated.");
        }

        var length = ReadUInt16LE(buffer, offset);
        offset += 2;
        if (offset + length > buffer.Length)
        {
            throw new InvalidDataException("String payload is truncated.");
        }

        nextOffset = offset + length;
        return Encoding.UTF8.GetString(buffer, offset, length);
    }

    public static double ReadFloat64LE(byte[] buffer, int offset, out int nextOffset)
    {
        if (offset + 8 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while reading float64.");
        }

        nextOffset = offset + 8;
        // BitConverter.ToDouble can be sensitive to endianness.
        // We assume the system running the client is Little Endian, like the server.
        // For cross-platform safety, manual conversion would be better, but this is fine for now.
        return BitConverter.ToDouble(buffer, offset);
    }

    public static void WriteFixed32LE(byte[] buffer, int offset, double value)
    {
        if (offset + 4 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while writing fixed32.");
        }
        var fixedValue = (int)Math.Round(value * 1000.0);
        WriteInt32LE(buffer, offset, fixedValue);
    }

    public static double ReadFixed32LE(byte[] buffer, int offset, out int nextOffset)
    {
        if (offset + 4 > buffer.Length)
        {
            throw new InvalidDataException("Buffer overflow while reading fixed32.");
        }
        var intValue = buffer[offset] | (buffer[offset + 1] << 8) | (buffer[offset + 2] << 16) | (buffer[offset + 3] << 24);
        nextOffset = offset + 4;
        return intValue / 1000.0;
    }
}

public sealed class LoginResponseData
{
    public bool Status { get; set; }
    public uint AccountId { get; set; }
    public string Token { get; set; } = string.Empty;
    public string ErrorCode { get; set; } = string.Empty;
}

public sealed class CharacterListResponseData
{
    public bool Status { get; set; }
    public string ErrorCode { get; set; } = string.Empty;
    public List<CharacterListEntryData> Characters { get; set; } = new();
}

public sealed class CharacterListEntryData
{
    public string Name { get; set; } = string.Empty;
    public string Class { get; set; } = string.Empty;
    public uint Level { get; set; }
    public string RaceId { get; set; } = string.Empty; // (R1-I-B)
}

public sealed class CharacterSelectResponseData
{
    public bool Status { get; set; }
    public string CharacterName { get; set; } = string.Empty;
    public string ErrorCode { get; set; } = string.Empty;
}

public sealed class CharacterCreateResponseData
{
    public bool Status { get; set; }
    public string ErrorCode { get; set; } = string.Empty;
    public CharacterListEntryData Character { get; set; } = new();
}

public sealed class InventoryItemData
{
    public string ItemId { get; set; } = string.Empty;
    public uint Quantity { get; set; }
    public uint Durability { get; set; }
    public ushort SlotIndex { get; set; }
}

public sealed class InventorySyncData
{
    public List<InventoryItemData> Items { get; set; } = new();
    public uint Level { get; set; }
    public double MaxHealth { get; set; }
    public double Health { get; set; }
    public double MaxMana { get; set; }
    public double Mana { get; set; }
    public double BaseAttack { get; set; }
    public double WeaponDamage { get; set; }
    public double Defense { get; set; }
    public double Resistance { get; set; }
    public double CritChance { get; set; }
}

public sealed class ChunkData
{
    public uint ChunkX { get; set; }
    public uint ChunkY { get; set; }
    public byte[] Tiles { get; set; } = Array.Empty<byte>();
}

public sealed class MoveConfirmData
{
    [System.Text.Json.Serialization.JsonPropertyName("x")]
    public double X { get; set; }
    [System.Text.Json.Serialization.JsonPropertyName("y")]
    public double Y { get; set; }
    [System.Text.Json.Serialization.JsonPropertyName("z")]
    public int Z { get; set; }
    [System.Text.Json.Serialization.JsonPropertyName("seq")]
    public uint Seq { get; set; }
    [System.Text.Json.Serialization.JsonPropertyName("success")]
    public bool Success { get; set; }
}

public sealed class DamageEventData
{
    public string AttackerID { get; set; } = string.Empty;
    public string TargetID { get; set; } = string.Empty;
    public double Damage { get; set; }
    public bool IsCrit { get; set; }
    public bool IsHit { get; set; }
    public bool Success { get; set; }
    public string SkillName { get; set; } = string.Empty;
}

public sealed class TargetDeadEventData
{
    public string TargetID { get; set; } = string.Empty;
}


public sealed class PlayerUpdateData
{
    [System.Text.Json.Serialization.JsonPropertyName("id")]
    public string PlayerID { get; set; } = string.Empty;
    [System.Text.Json.Serialization.JsonPropertyName("x")]
    public double X { get; set; }
    [System.Text.Json.Serialization.JsonPropertyName("y")]
    public double Y { get; set; }
    [System.Text.Json.Serialization.JsonPropertyName("z")]
    public int Z { get; set; }
}
