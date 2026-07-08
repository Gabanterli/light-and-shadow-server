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

        _topBarLabel.Text = $"Player: {characterState} | Level: pending sync | HP: pending sync | Mana: pending sync | {sessionState} | {clientState}";
    }

    private void RefreshWorldShellState()
    {
        if (_worldStatusLabel != null)
        {
            var viewState = _worldView != null ? "world view mounted" : "world view missing";
            _worldStatusLabel.Text = $"World sync pending. {viewState}. Alpha packet loop pending.";
        }
    }

    private void OnBackButtonPressed()
    {
        SceneFlow.ToDebugAuth(this);
    }
}
