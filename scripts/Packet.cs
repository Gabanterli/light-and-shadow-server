using System;
using System.Buffers.Binary;
using System.IO;
using System.Linq;

namespace LightAndShadow.Client;

public sealed class Packet
{
    public const ushort HeaderSize = 8;
    public const ushort MaxPacketSize = 16384;

    public ushort Size { get; set; }
    public ushort Opcode { get; set; }
    public uint Sequence { get; set; }
    public byte[] Payload { get; set; }

    public Packet(ushort opcode, uint sequence, byte[]? payload = null)
    {
        Opcode = opcode;
        Sequence = sequence;
        Payload = payload ?? Array.Empty<byte>();
        Size = (ushort)(HeaderSize + Payload.Length);
    }

    public byte[] Serialize()
    {
        if (Payload.Length > MaxPacketSize - HeaderSize)
        {
            throw new InvalidOperationException("Payload exceeds the maximum packet size.");
        }

        Size = (ushort)(HeaderSize + Payload.Length);
        var buffer = new byte[Size];
        BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(0, 2), Size);
        BinaryPrimitives.WriteUInt16LittleEndian(buffer.AsSpan(2, 2), Opcode);
        BinaryPrimitives.WriteUInt32LittleEndian(buffer.AsSpan(4, 4), Sequence);
        if (Payload.Length > 0)
        {
            Buffer.BlockCopy(Payload, 0, buffer, HeaderSize, Payload.Length);
        }

        return buffer;
    }

    public static Packet FromBytes(byte[] data)
    {
        if (data.Length < HeaderSize)
        {
            throw new InvalidDataException("Packet is smaller than the fixed header size.");
        }

        var size = BinaryPrimitives.ReadUInt16LittleEndian(data.AsSpan(0, 2));
        if (size < HeaderSize || size > MaxPacketSize)
        {
            throw new InvalidDataException($"Packet size {size} is invalid.");
        }

        if (data.Length < size)
        {
            throw new InvalidDataException("Packet payload is incomplete.");
        }

        var opcode = BinaryPrimitives.ReadUInt16LittleEndian(data.AsSpan(2, 2));
        var sequence = BinaryPrimitives.ReadUInt32LittleEndian(data.AsSpan(4, 4));
        var payload = data.Skip(HeaderSize).Take(size - HeaderSize).ToArray();

        return new Packet(opcode, sequence, payload)
        {
            Size = size
        };
    }
}
