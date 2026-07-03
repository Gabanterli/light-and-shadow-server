using Godot;
using System.Collections.Generic;

namespace LightAndShadow.Client;

public class DebugChunkStore
{
    private readonly Dictionary<(uint, uint), byte[]> _chunks = new();
    public IReadOnlyDictionary<(uint, uint), byte[]> Chunks => _chunks;

    public uint MinChunkX { get; private set; } = uint.MaxValue;
    public uint MinChunkY { get; private set; } = uint.MaxValue;

    public void AddChunk(uint chunkX, uint chunkY, byte[] tiles)
    {
        _chunks[(chunkX, chunkY)] = tiles;

        if (chunkX < MinChunkX)
        {
            MinChunkX = chunkX;
        }
        if (chunkY < MinChunkY)
        {
            MinChunkY = chunkY;
        }
    }
}