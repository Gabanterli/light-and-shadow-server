using LightAndShadow.Client;
using System;

namespace LightAndShadow.Client;

/// <summary>
/// A simple in-memory model to hold the initial world state snapshot received from the server.
/// This is for debug purposes only.
/// </summary>
public class DebugWorldBootstrapSnapshot
{
    public bool HasReceivedInventorySync { get; private set; }
    public int ItemCount { get; private set; }
    public uint Level { get; private set; }
    public double CurrentHealth { get; private set; }
    public double MaxHealth { get; private set; }
    public double CurrentMana { get; private set; }
    public double MaxMana { get; private set; }
    public int TotalChunksReceived { get; private set; }
    public uint LastChunkX { get; private set; }
    public uint LastChunkY { get; private set; }
    public int LastChunkTileCount { get; private set; }
    public DateTime LastPacketTimestamp { get; private set; }

    public void UpdateFromInventorySync(InventorySyncData data)
    {
        HasReceivedInventorySync = true;
        ItemCount = data.Items.Count;
        Level = data.Level;
        CurrentHealth = data.Health;
        MaxHealth = data.MaxHealth;
        CurrentMana = data.Mana;
        MaxMana = data.MaxMana;
        LastPacketTimestamp = DateTime.Now;
    }

    public void UpdateFromChunkData(ChunkData data)
    {
        TotalChunksReceived++;
        LastChunkX = data.ChunkX;
        LastChunkY = data.ChunkY;
        LastChunkTileCount = data.Tiles.Length;
        LastPacketTimestamp = DateTime.Now;
    }
}