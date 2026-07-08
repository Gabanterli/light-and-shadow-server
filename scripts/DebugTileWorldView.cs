using Godot;
using System;
using LightAndShadow.Client;

public partial class DebugTileWorldView : Control
{
    public DebugChunkStore? ChunkStore { get; set; }
    public Vector2I? PlayerTilePosition { get; set; }
    // Technical visualization for combat validation. Not final art.
    public bool IsOrcEliteDead { get; set; }
    public Vector2I? OrcElitePosition { get; set; }
    public Vector2I? TargetPosition { get; set; }

    // Debug remains unchanged by default. Alpha can opt into a focused playable viewport.
    public bool UseFocusedViewport { get; set; }
    public int MinimumFocusedViewportTilesWide { get; set; } = 24;
    public int FocusedViewportTilesHigh { get; set; } = 18;
    public bool ShowFixedCombatDebugOverlay { get; set; } = true;

    private const int TileSize = 8;
    private const int ChunkWidthInTiles = 32;
    private const int ChunkHeightInTiles = 32;

    // Debug-only colors
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

        if (UseFocusedViewport)
        {
            DrawFocusedViewport();
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

        // Draw player marker on top. Technical visualization for combat validation. Not final art.
        if (PlayerTilePosition.HasValue)
        {
            // Use a strong yellow color for the player marker with a black border for contrast.
            DrawDebugMarker(PlayerTilePosition.Value, _playerColor, Colors.Black, minPixelX, minPixelY, visibleRect);
        }

        // Draw Orc Elite marker on top. Technical visualization for combat validation. Not final art.
        if (OrcElitePosition.HasValue)
        {
            // Use a strong red color for the enemy marker with a white border for contrast.
            var orcColor = IsOrcEliteDead ? new Color(0.2f, 0.2f, 0.2f) : new Color(0.9f, 0.1f, 0.1f);
            DrawDebugMarker(OrcElitePosition.Value, orcColor, Colors.White, minPixelX, minPixelY, visibleRect);
        }

        // Draw movement target marker on top
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

        // Always draw the fixed combat overlay on top of everything else when enabled.
        if (ShowFixedCombatDebugOverlay)
        {
            DrawFixedCombatDebugOverlay();
        }
    }

    private void DrawFocusedViewport()
    {
        if (ChunkStore == null || ChunkStore.Chunks.Count == 0)
        {
            return;
        }

        var visibleRect = new Rect2(Vector2.Zero, Size);
        var viewportTilesHigh = Math.Max(1, FocusedViewportTilesHigh);

        if (Size.Y <= 0)
        {
            return;
        }

        var tileSize = Size.Y / viewportTilesHigh;
        if (tileSize <= 0)
        {
            return;
        }

        var viewportTilesWide = Math.Max(
            Math.Max(1, MinimumFocusedViewportTilesWide),
            Mathf.CeilToInt(Size.X / tileSize)
        );

        var centerTile = PlayerTilePosition ?? GetFallbackFocusedViewportCenterTile();

        var startTileX = centerTile.X - viewportTilesWide / 2;
        var startTileY = centerTile.Y - viewportTilesHigh / 2;

        for (var y = 0; y < viewportTilesHigh; y++)
        {
            for (var x = 0; x < viewportTilesWide; x++)
            {
                var globalTileX = startTileX + x;
                var globalTileY = startTileY + y;

                if (!TryGetTileType(globalTileX, globalTileY, out var tileType))
                {
                    continue;
                }

                var drawX = x * tileSize;
                var drawY = y * tileSize;
                var tileRect = new Rect2(drawX, drawY, tileSize, tileSize);
                DrawTile(tileRect, tileType, globalTileX, globalTileY, visibleRect);
            }
        }

        if (PlayerTilePosition.HasValue)
        {
            DrawFocusedDebugMarker(PlayerTilePosition.Value, _playerColor, Colors.Black, startTileX, startTileY, tileSize, visibleRect);
        }

        if (OrcElitePosition.HasValue)
        {
            var orcColor = IsOrcEliteDead ? new Color(0.2f, 0.2f, 0.2f) : new Color(0.9f, 0.1f, 0.1f);
            DrawFocusedDebugMarker(OrcElitePosition.Value, orcColor, Colors.White, startTileX, startTileY, tileSize, visibleRect);
        }

        if (TargetPosition.HasValue)
        {
            var targetRect = new Rect2(
                (TargetPosition.Value.X - startTileX) * tileSize,
                (TargetPosition.Value.Y - startTileY) * tileSize,
                tileSize,
                tileSize
            );

            if (visibleRect.Intersects(targetRect))
            {
                DrawRect(targetRect, _targetColor);
            }
        }
    }

    private Vector2I GetFallbackFocusedViewportCenterTile()
    {
        long minGlobalTileX = long.MaxValue;
        long minGlobalTileY = long.MaxValue;
        long maxGlobalTileX = long.MinValue;
        long maxGlobalTileY = long.MinValue;

        if (ChunkStore == null)
        {
            return Vector2I.Zero;
        }

        foreach (var coords in ChunkStore.Chunks.Keys)
        {
            var chunkGlobalMinX = (long)coords.Item1 * ChunkWidthInTiles;
            var chunkGlobalMinY = (long)coords.Item2 * ChunkHeightInTiles;
            var chunkGlobalMaxX = chunkGlobalMinX + ChunkWidthInTiles - 1;
            var chunkGlobalMaxY = chunkGlobalMinY + ChunkHeightInTiles - 1;

            minGlobalTileX = Math.Min(minGlobalTileX, chunkGlobalMinX);
            minGlobalTileY = Math.Min(minGlobalTileY, chunkGlobalMinY);
            maxGlobalTileX = Math.Max(maxGlobalTileX, chunkGlobalMaxX);
            maxGlobalTileY = Math.Max(maxGlobalTileY, chunkGlobalMaxY);
        }

        if (minGlobalTileX == long.MaxValue || minGlobalTileY == long.MaxValue)
        {
            return Vector2I.Zero;
        }

        return new Vector2I(
            checked((int)((minGlobalTileX + maxGlobalTileX) / 2)),
            checked((int)((minGlobalTileY + maxGlobalTileY) / 2))
        );
    }

    private bool TryGetTileType(int globalTileX, int globalTileY, out byte tileType)
    {
        tileType = 0;

        if (ChunkStore == null || globalTileX < 0 || globalTileY < 0)
        {
            return false;
        }

        var chunkX = (uint)(globalTileX / ChunkWidthInTiles);
        var chunkY = (uint)(globalTileY / ChunkHeightInTiles);
        var localTileX = globalTileX % ChunkWidthInTiles;
        var localTileY = globalTileY % ChunkHeightInTiles;

        if (!ChunkStore.Chunks.TryGetValue((chunkX, chunkY), out var tiles))
        {
            return false;
        }

        var tileIndex = localTileY * ChunkWidthInTiles + localTileX;
        if (tileIndex < 0 || tileIndex >= tiles.Length)
        {
            return false;
        }

        tileType = tiles[tileIndex];
        return true;
    }

    private void DrawFocusedDebugMarker(
        Vector2I tilePosition,
        Color fillColor,
        Color borderColor,
        int startTileX,
        int startTileY,
        float tileSize,
        Rect2 visibleRect)
    {
        var drawX = (tilePosition.X - startTileX) * tileSize - tileSize;
        var drawY = (tilePosition.Y - startTileY) * tileSize - tileSize;
        var markerRect = new Rect2(drawX, drawY, tileSize * 3, tileSize * 3);

        if (visibleRect.Intersects(markerRect))
        {
            DrawRect(markerRect, fillColor);
            DrawRect(markerRect, borderColor, false, 2.0f);
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

    // Debug-only technical marker for combat validation.
    private void DrawDebugMarker(Vector2I tilePosition, Color fillColor, Color borderColor, long minPixelX, long minPixelY, Rect2 visibleRect)
    {
        var markerGlobalTileX = tilePosition.X;
        var markerGlobalTileY = tilePosition.Y;

        // Center the 3x3 marker on the tile position by offsetting by one tile size.
        var drawX = (markerGlobalTileX * TileSize) - minPixelX - TileSize;
        var drawY = (markerGlobalTileY * TileSize) - minPixelY - TileSize;

        var markerRect = new Rect2(drawX, drawY, TileSize * 3, TileSize * 3);

        if (visibleRect.Intersects(markerRect))
        {
            // Draw the main filled color block.
            DrawRect(markerRect, fillColor);
            // Draw a thick border on top for high contrast.
            DrawRect(markerRect, borderColor, false, 2.0f);
        }
    }

    // Debug-only fixed overlay marker for combat validation.
    private void DrawFixedCombatDebugOverlay()
    {
        if (!PlayerTilePosition.HasValue)
        {
            return;
        }

        var playerTile = PlayerTilePosition.Value;
        // The backend spawns the orc at playerX+2, playerY+2, so this is a safe default before the controller gets the update.
        var orcTile = OrcElitePosition ?? new Vector2I(playerTile.X + 2, playerTile.Y + 2);

        var viewSize = Size;
        var playerScreenPosition = viewSize / 2;
        var markerSize = new Vector2(48, 48);

        // Calculate Orc position relative to the player's fixed center position.
        var delta = orcTile - playerTile;
        // Use a larger multiplier for better visual separation in the overlay.
        var orcScreenPosition = playerScreenPosition + new Vector2(delta.X * 32, delta.Y * 32);

        // Clamp the Orc's position to ensure it's always visible within the view.
        orcScreenPosition.X = Mathf.Clamp(orcScreenPosition.X, markerSize.X / 2, viewSize.X - markerSize.X / 2);
        orcScreenPosition.Y = Mathf.Clamp(orcScreenPosition.Y, markerSize.Y / 2, viewSize.Y - markerSize.Y / 2);

        // --- Draw Player Marker (Fixed Center) ---
        var playerRect = new Rect2(playerScreenPosition - markerSize / 2, markerSize);
        // Yellow fill
        DrawRect(playerRect, new Color(0.9f, 0.9f, 0.2f));
        // Black border
        DrawRect(playerRect, Colors.Black, false, 2.0f);

        // --- Draw Orc Marker (Relative to Player) ---
        var orcRect = new Rect2(orcScreenPosition - markerSize / 2, markerSize);
        // Red fill
        var orcFillColor = IsOrcEliteDead ? new Color(0.2f, 0.2f, 0.2f) : new Color(0.9f, 0.1f, 0.1f);
        DrawRect(orcRect, orcFillColor);
        // White border
        DrawRect(orcRect, Colors.White, false, 2.0f);
    }
}
