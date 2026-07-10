using Godot;
using System;
using System.Collections.Generic;
using System.Text.Json;
using LightAndShadow.Client;

public partial class DebugTileWorldView : Control
{
    public DebugChunkStore? ChunkStore { get; set; }
    public Vector2I? PlayerTilePosition { get; set; }
    // Technical visualization for combat validation. Not final art.
    public bool IsOrcEliteDead { get; set; }
    public bool IsOrcEliteSelected { get; set; }
    public Vector2I? OrcElitePosition { get; set; }
    public Vector2I? TargetPosition { get; set; }

    // Debug remains unchanged by default. Alpha can opt into a focused playable viewport.
    public bool UseFocusedViewport { get; set; }
    public int MinimumFocusedViewportTilesWide { get; set; } = 24;
    public int FocusedViewportTilesHigh { get; set; } = 18;
    public bool ShowFixedCombatDebugOverlay { get; set; } = true;
    public bool UseOneTileEntityMarkers { get; set; }
    public bool ShowAlphaCombatReadabilityHud { get; set; }
    public bool HasPlayerVitals { get; set; }
    public double PlayerHealth { get; set; }
    public double PlayerMaxHealth { get; set; }
    public double PlayerMana { get; set; }
    public double PlayerMaxMana { get; set; }
    private double _alphaOrcEliteMaxHealth = 100.0;
    private double _alphaOrcEliteCurrentHealth = 100.0;
    private string _alphaOrcEliteRuntimeEntityId = string.Empty;

    private const int TileSize = 8;
    private const int ChunkWidthInTiles = 32;
    private const int ChunkHeightInTiles = 32;

    // Debug-only colors
    private readonly Color _walkableColor = new(0.3f, 0.3f, 0.3f);
    private readonly Color _blockedColor = new(0.8f, 0.2f, 0.2f);
    private readonly Color _playerColor = new(0.9f, 0.9f, 0.2f);
    private readonly Color _targetColor = new(1.0f, 0.5f, 0.2f, 0.5f);

    private sealed class AlphaFloatingCombatText
    {
        public Vector2I TilePosition { get; set; }
        public string Text { get; set; } = string.Empty;
        public bool IsCritical { get; set; }
        public bool IsMiss { get; set; }
        public float AgeSeconds { get; set; }
        public float DurationSeconds { get; set; } = 0.95f;
        public float HorizontalOffset { get; set; }
    }

    private sealed class AlphaSpellVfxConfig
    {
        public string SkillName { get; set; } = string.Empty;
        public bool Enabled { get; set; }
        public float ColorR { get; set; }
        public float ColorG { get; set; }
        public float ColorB { get; set; }
        public float DurationSeconds { get; set; }
        public float LineWidth { get; set; }
        public float PulseBaseRadius { get; set; }
        public float PulseGrowthRadius { get; set; }
        public float ImpactRadius { get; set; }
    }

    private sealed class AlphaSpellVfxInstance
    {
        public Vector2I FromTilePosition { get; set; }
        public Vector2I ToTilePosition { get; set; }
        public AlphaSpellVfxConfig Config { get; set; } = new();
        public float AgeSeconds { get; set; }
    }

    private const int MaxAlphaFloatingCombatTexts = 8;
    private const int MaxAlphaConfirmedSpellVisuals = 4;
    private readonly System.Collections.Generic.List<AlphaFloatingCombatText> _alphaFloatingCombatTexts = new();
    private readonly System.Collections.Generic.List<AlphaSpellVfxInstance> _alphaSpellVfxInstances = new();
    private readonly Dictionary<string, AlphaSpellVfxConfig> _alphaSpellVfxConfig = new();

    private Texture2D? _grassTileTexture;
    private Texture2D? _dirtTileTexture;
    private Texture2D? _stoneBlockedTileTexture;
    private Texture2D? _waterTileTexture;

    public override void _Process(double delta)
    {
        if (_alphaFloatingCombatTexts.Count == 0 && _alphaSpellVfxInstances.Count == 0)
        {
            return;
        }

        for (var i = _alphaFloatingCombatTexts.Count - 1; i >= 0; i--)
        {
            _alphaFloatingCombatTexts[i].AgeSeconds += (float)delta;

            if (_alphaFloatingCombatTexts[i].AgeSeconds >= _alphaFloatingCombatTexts[i].DurationSeconds)
            {
                _alphaFloatingCombatTexts.RemoveAt(i);
            }
        }

        for (var i = _alphaSpellVfxInstances.Count - 1; i >= 0; i--)
        {
            _alphaSpellVfxInstances[i].AgeSeconds += (float)delta;

            if (_alphaSpellVfxInstances[i].AgeSeconds >= _alphaSpellVfxInstances[i].Config.DurationSeconds)
            {
                _alphaSpellVfxInstances.RemoveAt(i);
            }
        }

        QueueRedraw();
    }

    public void AddAlphaOrcEliteFloatingCombatText(string text, bool isCritical, bool isMiss)
    {
        if (!OrcElitePosition.HasValue || string.IsNullOrWhiteSpace(text))
        {
            return;
        }

        var offsetSlot = _alphaFloatingCombatTexts.Count % 3;
        var horizontalOffset = (offsetSlot - 1) * 12.0f;

        _alphaFloatingCombatTexts.Add(new AlphaFloatingCombatText
        {
            TilePosition = OrcElitePosition.Value,
            Text = text.Trim(),
            IsCritical = isCritical,
            IsMiss = isMiss,
            HorizontalOffset = horizontalOffset
        });

        while (_alphaFloatingCombatTexts.Count > MaxAlphaFloatingCombatTexts)
        {
            _alphaFloatingCombatTexts.RemoveAt(0);
        }

        QueueRedraw();
    }

    public void AddAlphaConfirmedSpellVisual(string skillName)
    {
        var normalizedSkillName = skillName.Trim();
        if (!PlayerTilePosition.HasValue || !OrcElitePosition.HasValue || string.IsNullOrWhiteSpace(normalizedSkillName))
        {
            return;
        }

        if (!_alphaSpellVfxConfig.TryGetValue(normalizedSkillName, out var vfxConfig) || !vfxConfig.Enabled)
        {
            return;
        }

        _alphaSpellVfxInstances.Add(new AlphaSpellVfxInstance
        {
            FromTilePosition = PlayerTilePosition.Value,
            ToTilePosition = OrcElitePosition.Value,
            Config = vfxConfig
        });

        while (_alphaSpellVfxInstances.Count > MaxAlphaConfirmedSpellVisuals)
        {
            _alphaSpellVfxInstances.RemoveAt(0);
        }

        QueueRedraw();
    }

    public void ApplyAlphaOrcEliteConfirmedDamage(double damage, string runtimeEntityId)
    {
        if (string.IsNullOrEmpty(_alphaOrcEliteRuntimeEntityId) && !string.IsNullOrEmpty(runtimeEntityId))
        {
            _alphaOrcEliteRuntimeEntityId = runtimeEntityId;
        }
        else if (runtimeEntityId != _alphaOrcEliteRuntimeEntityId)
        {
            GD.Print($"Ignoring damage for stale Orc Elite instance. Current: {_alphaOrcEliteRuntimeEntityId}, Stale: {runtimeEntityId}");
            return;
        }

        _alphaOrcEliteCurrentHealth = Math.Clamp(_alphaOrcEliteCurrentHealth - damage, 0, _alphaOrcEliteMaxHealth);
        GD.Print($"Alpha Orc Elite overhead HP damaged. New HP: {_alphaOrcEliteCurrentHealth}/{_alphaOrcEliteMaxHealth}");
        QueueRedraw();
    }

    public void MarkAlphaOrcEliteDead(string runtimeEntityId)
    {
        if (runtimeEntityId != _alphaOrcEliteRuntimeEntityId && !string.IsNullOrEmpty(_alphaOrcEliteRuntimeEntityId))
        {
            GD.Print($"Ignoring dead marker for stale Orc Elite instance. Current: {_alphaOrcEliteRuntimeEntityId}, Stale: {runtimeEntityId}");
            return;
        }
        _alphaOrcEliteCurrentHealth = 0;
        _alphaOrcEliteRuntimeEntityId = runtimeEntityId;
        GD.Print("Alpha Orc Elite overhead HP dead.");
        QueueRedraw();
    }

    public void ResetAlphaOrcEliteHealthForRespawn(string runtimeEntityId)
    {
        _alphaOrcEliteRuntimeEntityId = runtimeEntityId;
        _alphaOrcEliteCurrentHealth = _alphaOrcEliteMaxHealth;
        IsOrcEliteDead = false;
        GD.Print($"Alpha Orc Elite overhead HP respawn reset. New ID: {runtimeEntityId}");
        QueueRedraw();
    }

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

        LoadAlphaSpellVfxConfig();
        GD.Print("Alpha Orc Elite overhead HP initialized.");
    }

    private void LoadAlphaSpellVfxConfig()
    {
        const string configPath = "res://config/alpha_spell_vfx.json";
        if (!FileAccess.FileExists(configPath))
        {
            GD.PrintErr($"Alpha spell VFX config not found at: {configPath}");
            return;
        }

        try
        {
            using var file = FileAccess.Open(configPath, FileAccess.ModeFlags.Read);
            var content = file.GetAsText();
            var configRoot = JsonSerializer.Deserialize<JsonElement>(content);
            var spells = configRoot.GetProperty("Spells").EnumerateArray();

            foreach (var spellElement in spells)
            {
                var spellConfig = JsonSerializer.Deserialize<AlphaSpellVfxConfig>(spellElement.GetRawText());
                if (spellConfig != null && !string.IsNullOrWhiteSpace(spellConfig.SkillName) && !_alphaSpellVfxConfig.ContainsKey(spellConfig.SkillName))
                {
                    _alphaSpellVfxConfig[spellConfig.SkillName] = spellConfig;
                }
            }
            GD.Print($"DebugTileWorldView: Alpha spell VFX config loaded. Count: {_alphaSpellVfxConfig.Count}");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Failed to load or parse Alpha spell VFX config: {ex.Message}");
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

            if (IsOrcEliteSelected && !IsOrcEliteDead)
            {
                DrawDebugSelectionOutline(OrcElitePosition.Value, minPixelX, minPixelY, visibleRect);
            }
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
            if (ShowAlphaCombatReadabilityHud)
            {
                DrawFocusedPlayerVitalsHud(PlayerTilePosition.Value, startTileX, startTileY, tileSize, visibleRect);
            }
        }

        if (OrcElitePosition.HasValue)
        {
            var orcColor = IsOrcEliteDead ? new Color(0.2f, 0.2f, 0.2f) : new Color(0.9f, 0.1f, 0.1f);
            DrawFocusedDebugMarker(OrcElitePosition.Value, orcColor, Colors.White, startTileX, startTileY, tileSize, visibleRect);
            if (ShowAlphaCombatReadabilityHud)
            {
                DrawFocusedOrcHealthHud(OrcElitePosition.Value, startTileX, startTileY, tileSize, visibleRect);
                DrawFocusedAlphaFloatingCombatTexts(startTileX, startTileY, tileSize, visibleRect);
            }

            if (IsOrcEliteSelected && !IsOrcEliteDead)
            {
                DrawFocusedDebugSelectionOutline(OrcElitePosition.Value, startTileX, startTileY, tileSize, visibleRect);
            }
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

        // A30-R1: draw confirmed spell visual on top so it is not hidden by debug markers.
        DrawFocusedAlphaConfirmedSpellVisuals(startTileX, startTileY, tileSize, visibleRect);
    }


    private void DrawFocusedPlayerVitalsHud(Vector2I tilePosition, int startTileX, int startTileY, float tileSize, Rect2 visibleRect)
    {
        if (!HasPlayerVitals)
        {
            return;
        }

        var markerTileSize = UseOneTileEntityMarkers ? tileSize : tileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0.0f : tileSize;
        var drawX = (tilePosition.X - startTileX) * tileSize - markerOffset;
        var drawY = (tilePosition.Y - startTileY) * tileSize - markerOffset;
        var markerRect = new Rect2(drawX, drawY, markerTileSize, markerTileSize);

        if (!visibleRect.Intersects(markerRect))
        {
            return;
        }

        var hpBarRect = new Rect2(markerRect.Position.X - 10.0f, markerRect.Position.Y, 6.0f, markerRect.Size.Y);
        var manaBarRect = new Rect2(markerRect.Position.X + markerRect.Size.X + 4.0f, markerRect.Position.Y, 6.0f, markerRect.Size.Y);

        DrawVerticalVitalBar(hpBarRect, PlayerHealth, PlayerMaxHealth, new Color(0.1f, 0.9f, 0.2f, 0.75f), visibleRect);
        DrawVerticalVitalBar(manaBarRect, PlayerMana, PlayerMaxMana, new Color(0.2f, 0.45f, 1.0f, 0.75f), visibleRect);

        DrawSmallDebugLabel(new Vector2(hpBarRect.Position.X - 18.0f, hpBarRect.Position.Y - 4.0f), $"{PlayerHealth:F0}", Colors.White, visibleRect);
        DrawSmallDebugLabel(new Vector2(manaBarRect.Position.X + 8.0f, manaBarRect.Position.Y - 4.0f), $"{PlayerMana:F0}", Colors.White, visibleRect);
    }

    private void DrawFocusedOrcHealthHud(Vector2I tilePosition, int startTileX, int startTileY, float tileSize, Rect2 visibleRect)
    {
        var markerTileSize = UseOneTileEntityMarkers ? tileSize : tileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0.0f : tileSize;
        var drawX = (tilePosition.X - startTileX) * tileSize - markerOffset;
        var drawY = (tilePosition.Y - startTileY) * tileSize - markerOffset;
        var markerRect = new Rect2(drawX, drawY, markerTileSize, markerTileSize);

        if (!visibleRect.Intersects(markerRect))
        {
            return;
        }

        var hudWidth = Math.Max(markerRect.Size.X, 96.0f);
        var hudPosition = new Vector2(markerRect.GetCenter().X - hudWidth / 2.0f, markerRect.Position.Y - 26.0f);
        var backgroundRect = new Rect2(hudPosition, new Vector2(hudWidth, 18.0f));
        var barRect = new Rect2(hudPosition.X + 4.0f, hudPosition.Y + 11.0f, hudWidth - 8.0f, 4.0f);

        DrawRect(backgroundRect, new Color(0.0f, 0.0f, 0.0f, 0.45f));

        var healthRatio = CalculateVitalRatio(_alphaOrcEliteCurrentHealth, _alphaOrcEliteMaxHealth);
        DrawRect(barRect, new Color(0.15f, 0.0f, 0.0f, 0.75f));
        DrawRect(new Rect2(barRect.Position, new Vector2(barRect.Size.X * healthRatio, barRect.Size.Y)), new Color(0.9f, 0.1f, 0.1f, 0.85f));

        var stateText = $"Orc_Elite HP {_alphaOrcEliteCurrentHealth:F0}/{_alphaOrcEliteMaxHealth:F0}";
        DrawSmallDebugLabel(new Vector2(hudPosition.X + 4.0f, hudPosition.Y + 9.0f), stateText, Colors.White, visibleRect);
    }

    private void DrawVerticalVitalBar(Rect2 barRect, double current, double max, Color fillColor, Rect2 visibleRect)
    {
        if (!visibleRect.Intersects(barRect))
        {
            return;
        }

        var ratio = CalculateVitalRatio(current, max);
        DrawRect(barRect, new Color(0.0f, 0.0f, 0.0f, 0.45f));

        var fillHeight = barRect.Size.Y * ratio;
        var fillRect = new Rect2(
            barRect.Position.X,
            barRect.Position.Y + barRect.Size.Y - fillHeight,
            barRect.Size.X,
            fillHeight
        );

        DrawRect(fillRect, fillColor);
        DrawRect(barRect, Colors.Black, false, 1.0f);
    }

    private static float CalculateVitalRatio(double current, double max)
    {
        if (max <= 0)
        {
            return 0.0f;
        }

        return Mathf.Clamp((float)(current / max), 0.0f, 1.0f);
    }

    private void DrawSmallDebugLabel(Vector2 position, string text, Color color, Rect2 visibleRect)
    {
        if (string.IsNullOrWhiteSpace(text))
        {
            return;
        }

        var labelRect = new Rect2(position.X, position.Y - 12.0f, Math.Max(64.0f, text.Length * 7.0f), 14.0f);
        if (!visibleRect.Intersects(labelRect))
        {
            return;
        }

        var font = GetThemeDefaultFont();
        if (font == null)
        {
            return;
        }

        DrawString(font, position, text, HorizontalAlignment.Left, -1.0f, 11, color);
    }
    private void DrawFocusedAlphaConfirmedSpellVisuals(int startTileX, int startTileY, float tileSize, Rect2 visibleRect)
    {
        if (_alphaSpellVfxInstances.Count == 0)
        {
            return;
        }

        foreach (var spellVisual in _alphaSpellVfxInstances)
        {
            var progress = spellVisual.Config.DurationSeconds <= 0.0f
                ? 1.0f
                : Mathf.Clamp(spellVisual.AgeSeconds / spellVisual.Config.DurationSeconds, 0.0f, 1.0f);

            var opacity = 1.0f - progress;
            var from = GetFocusedTileCenter(spellVisual.FromTilePosition, startTileX, startTileY, tileSize);
            var to = GetFocusedTileCenter(spellVisual.ToTilePosition, startTileX, startTileY, tileSize);

            var bounds = new Rect2(
                Mathf.Min(from.X, to.X) - 16.0f,
                Mathf.Min(from.Y, to.Y) - 16.0f,
                Mathf.Abs(to.X - from.X) + 32.0f,
                Mathf.Abs(to.Y - from.Y) + 32.0f
            );

            if (!visibleRect.Intersects(bounds))
            {
                continue;
            }

            var color = new Color(spellVisual.Config.ColorR, spellVisual.Config.ColorG, spellVisual.Config.ColorB, opacity);
            DrawLine(from, to, color, spellVisual.Config.LineWidth);

            var pulsePosition = from.Lerp(to, progress);
            DrawCircle(pulsePosition, spellVisual.Config.PulseBaseRadius + spellVisual.Config.PulseGrowthRadius * progress, color);
            DrawCircle(to, spellVisual.Config.ImpactRadius * opacity, color);
        }
    }

    private Vector2 GetFocusedTileCenter(Vector2I tilePosition, int startTileX, int startTileY, float tileSize)
    {
        var markerTileSize = UseOneTileEntityMarkers ? tileSize : tileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0.0f : tileSize;

        return new Vector2(
            (tilePosition.X - startTileX) * tileSize - markerOffset + markerTileSize / 2.0f,
            (tilePosition.Y - startTileY) * tileSize - markerOffset + markerTileSize / 2.0f
        );
    }

    private static Color GetAlphaSpellVisualColor(string skillName, float opacity)
    {
        // Fallback for any legacy calls, though the new flow uses the config directly.
        return skillName.Trim() switch
        {
            "Fire Bolt" => new Color(1.0f, 0.35f, 0.05f, opacity),
            "Holy Spark" => new Color(1.0f, 0.95f, 0.55f, opacity),
            "Shadow Dart" => new Color(0.55f, 0.25f, 1.0f, opacity),
            _ => new Color(0.7f, 0.9f, 1.0f, opacity)
        };
    }

    private void DrawFocusedAlphaFloatingCombatTexts(int startTileX, int startTileY, float tileSize, Rect2 visibleRect)
    {
        if (_alphaFloatingCombatTexts.Count == 0)
        {
            return;
        }

        var markerTileSize = UseOneTileEntityMarkers ? tileSize : tileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0.0f : tileSize;

        foreach (var floatingText in _alphaFloatingCombatTexts)
        {
            var drawX = (floatingText.TilePosition.X - startTileX) * tileSize - markerOffset;
            var drawY = (floatingText.TilePosition.Y - startTileY) * tileSize - markerOffset;
            var markerRect = new Rect2(drawX, drawY, markerTileSize, markerTileSize);

            if (!visibleRect.Intersects(markerRect))
            {
                continue;
            }

            var progress = floatingText.DurationSeconds <= 0.0f
                ? 1.0f
                : Mathf.Clamp(floatingText.AgeSeconds / floatingText.DurationSeconds, 0.0f, 1.0f);

            var opacity = 1.0f - progress;
            var verticalRise = 34.0f * progress;

            var color = floatingText.IsMiss
                ? new Color(0.9f, 0.9f, 0.9f, opacity)
                : floatingText.IsCritical
                    ? new Color(1.0f, 0.75f, 0.1f, opacity)
                    : new Color(1.0f, 1.0f, 1.0f, opacity);

            var textPosition = new Vector2(
                markerRect.GetCenter().X - 22.0f + floatingText.HorizontalOffset,
                markerRect.Position.Y - 18.0f - verticalRise
            );

            DrawSmallDebugLabel(textPosition, floatingText.Text, color, visibleRect);
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
        var markerTileSize = UseOneTileEntityMarkers ? tileSize : tileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0 : tileSize;
        var drawX = (tilePosition.X - startTileX) * tileSize - markerOffset;
        var drawY = (tilePosition.Y - startTileY) * tileSize - markerOffset;
        var markerRect = new Rect2(drawX, drawY, markerTileSize, markerTileSize);

        if (visibleRect.Intersects(markerRect))
        {
            DrawRect(markerRect, fillColor);
            DrawRect(markerRect, borderColor, false, 2.0f);
        }
    }

    private void DrawFocusedDebugSelectionOutline(
        Vector2I tilePosition,
        int startTileX,
        int startTileY,
        float tileSize,
        Rect2 visibleRect)
    {
        var markerTileSize = UseOneTileEntityMarkers ? tileSize : tileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0 : tileSize;
        var drawX = (tilePosition.X - startTileX) * tileSize - markerOffset - 3.0f;
        var drawY = (tilePosition.Y - startTileY) * tileSize - markerOffset - 3.0f;
        var markerRect = new Rect2(drawX, drawY, markerTileSize + 6.0f, markerTileSize + 6.0f);

        if (visibleRect.Intersects(markerRect))
        {
            DrawRect(markerRect, new Color(1.0f, 0.0f, 0.0f, 0.95f), false, 4.0f);
        }
    }

    private void DrawDebugSelectionOutline(Vector2I tilePosition, long minPixelX, long minPixelY, Rect2 visibleRect)
    {
        var markerTileSize = UseOneTileEntityMarkers ? TileSize : TileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0 : TileSize;
        var drawX = (tilePosition.X * TileSize) - minPixelX - markerOffset - 2;
        var drawY = (tilePosition.Y * TileSize) - minPixelY - markerOffset - 2;
        var markerRect = new Rect2(drawX, drawY, markerTileSize + 4, markerTileSize + 4);

        if (visibleRect.Intersects(markerRect))
        {
            DrawRect(markerRect, new Color(1.0f, 0.0f, 0.0f, 0.95f), false, 4.0f);
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

        var markerTileSize = UseOneTileEntityMarkers ? TileSize : TileSize * 3;
        var markerOffset = UseOneTileEntityMarkers ? 0 : TileSize;

        var drawX = (markerGlobalTileX * TileSize) - minPixelX - markerOffset;
        var drawY = (markerGlobalTileY * TileSize) - minPixelY - markerOffset;

        var markerRect = new Rect2(drawX, drawY, markerTileSize, markerTileSize);

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
