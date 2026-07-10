using System;
using System.IO;
using System.Net.Sockets;
using System.Threading;
using System.Threading.Tasks;

namespace LightAndShadow.Client;

public sealed class GatewayTcpClient : IDisposable
{
    private readonly string _host;
    private readonly int _port;
    private TcpClient? _tcpClient;
    private NetworkStream? _stream;
    private uint _nextSequence = 1;
    private readonly SemaphoreSlim _sendLock = new(1, 1);

    public bool IsConnected => _tcpClient?.Connected == true && _stream != null;

    public GatewayTcpClient(string host = "127.0.0.1", int port = 8080)
    {
        _host = host;
        _port = port;
    }

    public async Task ConnectAsync(CancellationToken cancellationToken = default)
    {
        if (IsConnected)
        {
            return;
        }

        _tcpClient?.Dispose();
        _stream?.Dispose();

        _tcpClient = new TcpClient();
        try
        {
            using var timeoutCts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
            timeoutCts.CancelAfter(TimeSpan.FromSeconds(5));

            var connectTask = _tcpClient.ConnectAsync(_host, _port, timeoutCts.Token);
            await connectTask;
            _stream = _tcpClient.GetStream();
            _stream.ReadTimeout = 5000;
            _stream.WriteTimeout = 5000;
        }
        catch (SocketException ex) when (ex.SocketErrorCode == SocketError.ConnectionRefused)
        {
            throw new IOException("Connection refused by the gateway server.", ex);
        }
        catch (OperationCanceledException)
        {
            _tcpClient.Dispose();
            _tcpClient = null;
            throw new TimeoutException("Timed out while connecting to the gateway.");
        }
    }

    public async Task<LoginResponseData> LoginAsync(string username, string password, CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var payload = BinaryProtocol.EncodeLoginRequest(username, password);
        var packet = new Packet(1002, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
        var responsePacket = await ReceivePacketAsync(cancellationToken);
        if (responsePacket.Opcode != 1003)
        {
            throw new InvalidDataException($"Unexpected opcode {responsePacket.Opcode} while waiting for login response.");
        }

        return BinaryProtocol.DecodeLoginResponse(responsePacket.Payload);
    }

    public async Task<CharacterListResponseData> RequestCharacterListAsync(CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var packet = new Packet(1004, NextSequence(), BinaryProtocol.EncodeCharacterListRequest());
        await SendPacketAsync(packet, cancellationToken);
        var responsePacket = await ReceivePacketAsync(cancellationToken);
        if (responsePacket.Opcode != 1005)
        {
            throw new InvalidDataException($"Unexpected opcode {responsePacket.Opcode} while waiting for character list response.");
        }

        return BinaryProtocol.DecodeCharacterListResponse(responsePacket.Payload);
    }

    public async Task<CharacterSelectResponseData> SelectCharacterAsync(string characterName, CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var payload = BinaryProtocol.EncodeCharacterSelectRequest(characterName);
        var packet = new Packet(1006, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
        var responsePacket = await ReceivePacketAsync(cancellationToken);
        if (responsePacket.Opcode != 1007)
        {
            throw new InvalidDataException($"Unexpected opcode {responsePacket.Opcode} while waiting for character selection response.");
        }

        return BinaryProtocol.DecodeCharacterSelectResponse(responsePacket.Payload);
    }

    public async Task<CharacterCreateResponseData> CreateCharacterAsync(string desiredName, string raceId, CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var payload = BinaryProtocol.EncodeCharacterCreateRequest(desiredName, raceId);
        var packet = new Packet(1008, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
        var responsePacket = await ReceivePacketAsync(cancellationToken);
        if (responsePacket.Opcode != 1009)
        {
            throw new InvalidDataException($"Unexpected opcode {responsePacket.Opcode} while waiting for character creation response.");
        }
        return BinaryProtocol.DecodeCharacterCreateResponse(responsePacket.Payload);
    }

    public async Task SendMoveRequestAsync(int targetX, int targetY, sbyte targetZ, byte direction, ulong clientTimestamp, CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var payload = BinaryProtocol.EncodeMoveRequest(targetX, targetY, targetZ, direction, clientTimestamp);
        var packet = new Packet(2004, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
    }


    public async Task SendNpcInteractRequestAsync(string npcId, CancellationToken cancellationToken = default)
    {
        var payload = BinaryProtocol.EncodeNpcInteractRequest(npcId);
        var packet = new Packet(BinaryProtocol.CS_NPC_INTERACT, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
    }
    public async Task SendAttackRequestAsync(string targetId, string weaponType, CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var payload = BinaryProtocol.EncodeAttackRequest(targetId, weaponType);
        var packet = new Packet(3000, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
    }


    public async Task SendCastSkillRequestAsync(uint skillId, string targetId, double targetX, double targetY, CancellationToken cancellationToken = default)
    {
        EnsureConnected();
        var payload = BinaryProtocol.EncodeCastSkillRequest(skillId, targetId, targetX, targetY);
        var packet = new Packet(3001, NextSequence(), payload);
        await SendPacketAsync(packet, cancellationToken);
    }

    public void Disconnect()
    {
        try
        {
            _stream?.Dispose();
            _tcpClient?.Dispose();
        }
        catch
        {
            // Suppress exceptions during disconnection as it's a cleanup operation.
        }
        _stream = null;
        _tcpClient = null;
    }

    private void EnsureConnected()
    {
        if (!IsConnected)
        {
            throw new InvalidOperationException("Gateway client is not connected.");
        }
    }

    private uint NextSequence()
    {
        return _nextSequence++;
    }

    private async Task SendPacketAsync(Packet packet, CancellationToken cancellationToken)
    {
        EnsureConnected();
        var bytes = packet.Serialize();
        await _sendLock.WaitAsync(cancellationToken);
        try
        {
            await _stream!.WriteAsync(bytes, cancellationToken);
            await _stream.FlushAsync(cancellationToken);
        }
        finally
        {
            _sendLock.Release();
        }
    }

    public async Task<Packet> ReceivePacketAsync(CancellationToken cancellationToken)
    {
        EnsureConnected();
        var header = await ReadExactlyAsync(Packet.HeaderSize, cancellationToken);
        var size = BinaryProtocol.ReadUInt16LE(header, 0);
        if (size < Packet.HeaderSize || size > Packet.MaxPacketSize)
        {
            throw new InvalidDataException($"Malformed packet size {size}.");
        }

        var payload = await ReadExactlyAsync(size - Packet.HeaderSize, cancellationToken);
        return new Packet(BinaryProtocol.ReadUInt16LE(header, 2), BinaryProtocol.ReadUInt32LE(header, 4), payload)
        {
            Size = size
        };
    }

    private async Task<byte[]> ReadExactlyAsync(int size, CancellationToken cancellationToken)
    {
        if (size <= 0)
        {
            return Array.Empty<byte>();
        }

        EnsureConnected();
        var buffer = new byte[size];
        var offset = 0;
        while (offset < size)
        {
            var read = await _stream!.ReadAsync(buffer.AsMemory(offset, size - offset), cancellationToken);
            if (read == 0)
            {
                throw new IOException("Connection closed while reading from the gateway.");
            }

            offset += read;
        }

        return buffer;
    }

    public void Dispose()
    {
        Disconnect();
        _sendLock.Dispose();
    }
}
