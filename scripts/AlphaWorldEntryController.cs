using Godot;
using LightAndShadow.Client;

public partial class AlphaWorldEntryController : Control
{
    private Button? _backButton;

    public override void _Ready()
    {
        _backButton = GetNodeOrNull<Button>("Root/TopBar/TopBarHBox/BackButton");
        if (_backButton != null)
        {
            _backButton.Pressed += OnBackButtonPressed;
        }

        GD.Print("AlphaWorldEntryController loaded: UI shell only. Backend wiring pending.");
    }

    private void OnBackButtonPressed()
    {
        SceneFlow.ToDebugAuth(this);
    }
}
