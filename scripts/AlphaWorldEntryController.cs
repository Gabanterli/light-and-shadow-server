using Godot;
using LightAndShadow.Client;

public partial class AlphaWorldEntryController : Control
{
    public AuthSession? Session { get; set; }
    public GatewayTcpClient? GatewayClient { get; set; }

    private Button? _backButton;
    private Label? _topBarLabel;

    public override void _Ready()
    {
        _topBarLabel = GetNodeOrNull<Label>("Root/TopBar/TopBarHBox/TopBarLabel");
        _backButton = GetNodeOrNull<Button>("Root/TopBar/TopBarHBox/BackButton");

        if (_backButton != null)
        {
            _backButton.Pressed += OnBackButtonPressed;
        }

        RefreshTopBarShellState();

        GD.Print("AlphaWorldEntryController loaded: UI shell only. Backend packet loop pending.");
    }

    private void RefreshTopBarShellState()
    {
        if (_topBarLabel == null)
        {
            return;
        }

        var sessionState = Session != null ? "session received" : "session missing";
        var clientState = GatewayClient?.IsConnected == true ? "client connected" : "client disconnected";

        _topBarLabel.Text = $"Player: pending backend sync | Level: - | HP: - | Mana: - | {sessionState} | {clientState}";
    }

    private void OnBackButtonPressed()
    {
        SceneFlow.ToDebugAuth(this);
    }
}
