using Godot;
using LightAndShadow.Client;

public partial class AlphaWorldEntryController : Control
{
    public AuthSession? Session { get; set; }
    public GatewayTcpClient? GatewayClient { get; set; }

    private Button? _backButton;
    private Label? _topBarLabel;
    private Label? _worldStatusLabel;
    private DebugTileWorldView? _worldView;

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
        _worldView = GetNodeOrNull<DebugTileWorldView>("Root/MainArea/WorldPanel/WorldVBox/AlphaWorldView");
        _backButton = GetNodeOrNull<Button>("Root/TopBar/TopBarHBox/BackButton");

        if (_backButton != null)
        {
            _backButton.Pressed += OnBackButtonPressed;
        }

        RefreshTopBarShellState();
        RefreshWorldShellState();

        GD.Print("AlphaWorldEntryController loaded: UI shell only. Backend packet loop pending.");
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

    private void RefreshWorldShellState()
    {
        if (_worldStatusLabel != null)
        {
            var viewState = _worldView != null ? "world view mounted" : "world view missing";
            _worldStatusLabel.Text = $"World sync pending. {viewState}. Alpha packet loop pending.";
        }
    }

    private void ApplyInventorySyncData(InventorySyncData data)
    {
        _hasInventorySync = true;
        _syncedLevel = data.Level;
        _syncedHealth = data.Health;
        _syncedMaxHealth = data.MaxHealth;
        _syncedMana = data.Mana;
        _syncedMaxMana = data.MaxMana;
        _syncedItemCount = data.Items.Count;

        RefreshTopBarShellState();

        GD.Print($"Alpha inventory sync state prepared: level={_syncedLevel}, hp={_syncedHealth:F2}/{_syncedMaxHealth:F2}, mana={_syncedMana:F2}/{_syncedMaxMana:F2}, items={_syncedItemCount}");
    }
    private void OnBackButtonPressed()
    {
        SceneFlow.ToDebugAuth(this);
    }
}
