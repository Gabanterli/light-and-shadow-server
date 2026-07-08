using Godot;
using LightAndShadow.Client;
using System;
using System.Threading;

public partial class AlphaWorldEntryController : Control
{
    public AuthSession? Session { get; set; }
    public GatewayTcpClient? GatewayClient { get; set; }

    private Button? _backButton;
    private Label? _topBarLabel;
    private Label? _worldStatusLabel;
    private Label? _systemFeedbackLabel;
    private Label? _backpackLabel;
    private DebugTileWorldView? _worldView;

    private CancellationTokenSource? _packetLoopCts;
    private int _ignoredPacketCount;

    private bool _hasInventorySync;
    private uint _syncedLevel;
    private double _syncedHealth;
    private double _syncedMaxHealth;
    private double _syncedMana;
    private double _syncedMaxMana;
    private int _syncedItemCount;

    public override void _Ready()
    {
        _topBarLabel = GetNodeOrNull<Label>("Root/TopBar/TopBarHBox/TopBarLabel");
        _worldStatusLabel = GetNodeOrNull<Label>("Root/MainArea/WorldPanel/WorldVBox/WorldStatusLabel");
        _systemFeedbackLabel = GetNodeOrNull<Label>("Root/BottomTabs/System");
        _backpackLabel = GetNodeOrNull<Label>("Root/MainArea/SideTabs/Backpack");
        _worldView = GetNodeOrNull<DebugTileWorldView>("Root/MainArea/WorldPanel/WorldVBox/AlphaWorldView");
        _backButton = GetNodeOrNull<Button>("Root/TopBar/TopBarHBox/BackButton");

        if (_backButton != null)
        {
            _backButton.Pressed += OnBackButtonPressed;
        }

        RefreshTopBarShellState();
        RefreshBackpackShellState();
        RefreshWorldShellState();
        StartAlphaInventorySyncPacketLoop();

        GD.Print("AlphaWorldEntryController loaded: InventorySync-only packet loop boundary active.");
    }

    public override void _ExitTree()
    {
        StopAlphaPacketLoop();
        _packetLoopCts?.Dispose();
        _packetLoopCts = null;
        GatewayClient?.Dispose();
    }

    private void RefreshTopBarShellState()
    {
        if (_topBarLabel == null)
        {
            return;
        }

        var sessionState = Session != null ? "session received" : "session missing";
        var characterState = Session?.IsCharacterSelected == true
            ? Session.SelectedCharacterName
            : "pending character selection";
        var clientState = GatewayClient?.IsConnected == true ? "client connected" : "client disconnected";
        var levelState = _hasInventorySync ? _syncedLevel.ToString() : "pending sync";
        var hpState = _hasInventorySync ? $"{_syncedHealth:F0}/{_syncedMaxHealth:F0}" : "pending sync";
        var manaState = _hasInventorySync ? $"{_syncedMana:F0}/{_syncedMaxMana:F0}" : "pending sync";

        _topBarLabel.Text = $"Player: {characterState} | Level: {levelState} | HP: {hpState} | Mana: {manaState} | {sessionState} | {clientState}";
    }

    private void RefreshBackpackShellState()
    {
        if (_backpackLabel == null)
        {
            return;
        }

        var itemCountState = _hasInventorySync ? $"{_syncedItemCount} synced" : "pending sync";
        _backpackLabel.Text = $"Backpack\n\nItems: {itemCountState}\nReal inventory sync only.";
    }

    private void RefreshWorldShellState()
    {
        if (_worldStatusLabel != null)
        {
            var viewState = _worldView != null ? "world view mounted" : "world view missing";
            _worldStatusLabel.Text = $"World sync pending. {viewState}. Alpha packet loop: InventorySync 4001 only.";
        }
    }

    private async void StartAlphaInventorySyncPacketLoop()
    {
        if (GatewayClient == null)
        {
            SetAlphaSystemMessage("Alpha packet loop not started: client missing.");
            return;
        }

        if (!GatewayClient.IsConnected)
        {
            SetAlphaSystemMessage("Alpha packet loop not started: client disconnected.");
            return;
        }

        if (_packetLoopCts != null && !_packetLoopCts.IsCancellationRequested)
        {
            SetAlphaSystemMessage("Alpha packet loop already running.");
            return;
        }

        _packetLoopCts = new CancellationTokenSource();
        var token = _packetLoopCts.Token;

        SetAlphaSystemMessage("Alpha InventorySync listener started. Waiting for opcode 4001.");

        try
        {
            while (!token.IsCancellationRequested)
            {
                var packet = await GatewayClient.ReceivePacketAsync(token);

                if (packet.Opcode == 4001)
                {
                    HandleAlphaInventorySyncPacket(packet);
                }
                else
                {
                    _ignoredPacketCount++;
                    if (_ignoredPacketCount == 1 || _ignoredPacketCount % 10 == 0)
                    {
                        CallDeferred(nameof(SetAlphaSystemMessage), $"Alpha listener ignoring non-4001 packets. Ignored: {_ignoredPacketCount}");
                    }
                }
            }
        }
        catch (OperationCanceledException)
        {
            GD.Print("Alpha InventorySync listener stopped.");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha InventorySync listener stopped with error: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Alpha packet loop error: {ex.GetType().Name}");
        }
    }

    private void StopAlphaPacketLoop()
    {
        if (_packetLoopCts == null)
        {
            return;
        }

        if (!_packetLoopCts.IsCancellationRequested)
        {
            _packetLoopCts.Cancel();
        }
    }

    private void HandleAlphaInventorySyncPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeInventorySync(packet.Payload);
            CallDeferred(
                nameof(ApplyInventorySyncValues),
                data.Level,
                data.Health,
                data.MaxHealth,
                data.Mana,
                data.MaxMana,
                data.Items.Count
            );
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha InventorySync decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"InventorySync decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyInventorySyncData(InventorySyncData data)
    {
        ApplyInventorySyncValues(
            data.Level,
            data.Health,
            data.MaxHealth,
            data.Mana,
            data.MaxMana,
            data.Items.Count
        );
    }

    private void ApplyInventorySyncValues(uint level, double health, double maxHealth, double mana, double maxMana, int itemCount)
    {
        _hasInventorySync = true;
        _syncedLevel = level;
        _syncedHealth = health;
        _syncedMaxHealth = maxHealth;
        _syncedMana = mana;
        _syncedMaxMana = maxMana;
        _syncedItemCount = itemCount;

        RefreshTopBarShellState();
        RefreshBackpackShellState();
        SetAlphaSystemMessage($"InventorySync 4001 received. Items: {_syncedItemCount}");

        GD.Print($"Alpha inventory sync applied: level={_syncedLevel}, hp={_syncedHealth:F2}/{_syncedMaxHealth:F2}, mana={_syncedMana:F2}/{_syncedMaxMana:F2}, items={_syncedItemCount}");
    }

    private void SetAlphaSystemMessage(string message)
    {
        if (_systemFeedbackLabel != null)
        {
            _systemFeedbackLabel.Text = $"System\n- {message}";
        }

        GD.Print($"Alpha System: {message}");
    }

    private void OnBackButtonPressed()
    {
        StopAlphaPacketLoop();
        SceneFlow.ToDebugAuth(this);
    }
}
