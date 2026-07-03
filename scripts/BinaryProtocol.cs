using System;
using System.Collections.Generic;
using System.IO;
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

            characters.Add(new CharacterListEntryData
            {
                Name = name,
                Class = className,
                Level = level
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
}

public sealed class CharacterSelectResponseData
{
    public bool Status { get; set; }
    public string CharacterName { get; set; } = string.Empty;
    public string ErrorCode { get; set; } = string.Empty;
}
