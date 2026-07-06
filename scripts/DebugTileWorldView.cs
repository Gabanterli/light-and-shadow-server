using Godot;
using LightAndShadow.Client;

public partial class DebugTileWorldView : Control
{
    public DebugChunkStore? ChunkStore { get; set; }
    public Vector2I? PlayerTilePosition { get; set; }
    public Vector2I? TargetPosition { get; set; }

    private const int TileSize = 8;
    private const int ChunkWidthInTiles = 32;
    private const int ChunkHeightInTiles = 32;

    private readonly Color _walkableColor = new(0.3f, 0.3f, 0.3f);
    private readonly Color _blockedColor = new(0.8f, 0.2f, 0.2f);
    private readonly Color _playerColor = new(0.9f, 0.9f, 0.2f);
    private readonly Color _targetColor = new(1.0f, 0.5f, 0.2f, 0.5f);

    private Texture2D? _grassTileTexture;
    private Texture2D? _dirtTileTexture;
    private Texture2D? _stoneBlockedTileTexture;
    private Texture2D? _waterTileTexture;

    public override void _Ready()
    {
        // Load placeholder textures. GD.Load returns null if the path is invalid,
        // which allows the _Draw method to gracefully fall back to solid colors.
        _grassTileTexture = GD.Load<Texture2D>("res://assets/placeholders/tiles/tile_grass_placeholder_01.png");
        _dirtTileTexture = GD.Load<Texture2D>("res://assets/placeholders/tiles/tile_dirt_placeholder_01.png");
        _stoneBlockedTileTexture = GD.Load<Texture2D>("res://assets/placeholders/tiles/tile_stone_blocked_placeholder_01.png");
        _waterTileTexture = GD.Load<Texture2D>("res://assets/placeholders/tiles/tile_water_placeholder_01.png");

        if (_grassTileTexture == null || _dirtTileTexture == null || _stoneBlockedTileTexture == null || _waterTileTexture == null)
        {
            GD.PrintErr("DebugTileWorldView: One or more placeholder textures failed to load. The view will fall back to solid colors.");
        }
    }

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
                    var globalTileX = checked((int)chunkX * ChunkWidthInTiles + x);
                    var globalTileY = checked((int)chunkY * ChunkHeightInTiles + y);

                    var drawX = (globalTileX * TileSize) - minPixelX;
                    var drawY = (globalTileY * TileSize) - minPixelY;
                    var tileRect = new Rect2(drawX, drawY, TileSize, TileSize);
                    DrawTile(tileRect, tileType, globalTileX, globalTileY, visibleRect);
                }
            }
        }

        // Draw player marker on top
        if (PlayerTilePosition.HasValue)
        {
            var playerGlobalTileX = PlayerTilePosition.Value.X;
            var playerGlobalTileY = PlayerTilePosition.Value.Y;

            var drawX = (playerGlobalTileX * TileSize) - minPixelX;
            var drawY = (playerGlobalTileY * TileSize) - minPixelY;

            var playerRect = new Rect2(drawX, drawY, TileSize, TileSize);
            if (visibleRect.Intersects(playerRect))
            {
                DrawRect(playerRect, _playerColor);
            }
        }

        // Draw target marker on top
        if (TargetPosition.HasValue)
        {
            var targetGlobalTileX = TargetPosition.Value.X;
            var targetGlobalTileY = TargetPosition.Value.Y;

            var drawX = (targetGlobalTileX * TileSize) - minPixelX;
            var drawY = (targetGlobalTileY * TileSize) - minPixelY;

            var targetRect = new Rect2(drawX, drawY, TileSize, TileSize);
            if (visibleRect.Intersects(targetRect))
            {
                DrawRect(targetRect, _targetColor);
            }
        }
    }

    private void DrawTile(Rect2 tileRect, byte tileType, int globalTileX, int globalTileY, Rect2 visibleRect)
    {
        if (!visibleRect.Intersects(tileRect))
        {
            return;
        }

        Texture2D? textureToDraw = null;
        Color fallbackColor = _blockedColor; // Default to blocked color

        switch (tileType)
        {
            case 0: // Walkable
                // Alternate between grass and dirt for a simple checkerboard pattern
                textureToDraw = (globalTileX + globalTileY) % 2 == 0 ? _grassTileTexture : _dirtTileTexture;
                fallbackColor = _walkableColor;
                break;

            case 1: // Blocked
                textureToDraw = _stoneBlockedTileTexture;
                fallbackColor = _blockedColor;
                break;

            case 2: // Water
                textureToDraw = _waterTileTexture;
                fallbackColor = new Color(0.2f, 0.3f, 0.8f); // Blue for water fallback
                break;

            default: // Any other type is considered blocked
                textureToDraw = _stoneBlockedTileTexture;
                fallbackColor = _blockedColor;
                break;
        }

        if (textureToDraw != null)
        {
            DrawTextureRect(textureToDraw, tileRect, false);
        }
        else
        {
            // Fallback to drawing a solid color if the texture is not loaded
            DrawRect(tileRect, fallbackColor);
        }
    }
}
