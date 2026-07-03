using Godot;
using LightAndShadow.Client;

public partial class DebugTileWorldView : Control
{
    public DebugChunkStore? ChunkStore { get; set; }

    private const int TileSize = 4;
    private const int ChunkWidthInTiles = 32;
    private const int ChunkHeightInTiles = 32;

    private readonly Color _walkableColor = new(0.3f, 0.3f, 0.3f);
    private readonly Color _blockedColor = new(0.8f, 0.2f, 0.2f);

    public override void _Draw()
    {
        if (ChunkStore == null || ChunkStore.Chunks.Count == 0)
        {
            return;
        }

        var minPixelX = ChunkStore.MinChunkX * ChunkWidthInTiles * TileSize;
        var minPixelY = ChunkStore.MinChunkY * ChunkHeightInTiles * TileSize;

        var visibleRect = new Rect2(Vector2.Zero, Size);

        foreach (var (coords, tiles) in ChunkStore.Chunks)
        {
            var chunkX = coords.Item1;
            var chunkY = coords.Item2;

            for (var y = 0; y < ChunkHeightInTiles; y++)
            {
                for (var x = 0; x < ChunkWidthInTiles; x++)
                {
                    var tileIndex = y * ChunkWidthInTiles + x;
                    if (tileIndex >= tiles.Length) continue;

                    var tileType = tiles[tileIndex];
                    var color = tileType == 0 ? _walkableColor : _blockedColor;

                    var globalTileX = chunkX * ChunkWidthInTiles + x;
                    var globalTileY = chunkY * ChunkHeightInTiles + y;

                    var drawX = (globalTileX * TileSize) - minPixelX;
                    var drawY = (globalTileY * TileSize) - minPixelY;

                    var tileRect = new Rect2(drawX, drawY, TileSize, TileSize);
                    if (visibleRect.Intersects(tileRect))
                    {
                        DrawRect(tileRect, color);
                    }
                }
            }
        }
    }
}