using LightAndShadow.Client;
using System;
using System.Collections.Generic;

namespace LightAndShadow.Client;

/// <summary>
/// A simple, provisional packet router to dispatch incoming packets based on their opcode.
/// </summary>
public class DebugIncomingPacketRouter
{
    private readonly Dictionary<ushort, Action<Packet>> _handlers = new();
    private Action<Packet>? _fallbackHandler;

    /// <summary>
    /// Registers a handler for a specific opcode.
    /// </summary>
    public void RegisterHandler(ushort opcode, Action<Packet> handler)
    {
        _handlers[opcode] = handler;
    }

    /// <summary>
    /// Registers a fallback handler for any opcode that doesn't have a specific handler.
    /// </summary>
    public void RegisterFallback(Action<Packet> handler)
    {
        _fallbackHandler = handler;
    }

    /// <summary>
    /// Dispatches a packet to the appropriate registered handler.
    /// </summary>
    public void Dispatch(Packet packet)
    {
        if (!_handlers.TryGetValue(packet.Opcode, out var handler))
        {
            handler = _fallbackHandler;
        }
        handler?.Invoke(packet);
    }
}