using Godot;
using LightAndShadow.Client;
using System;

public partial class DebugWorldEntryController : Control
{
    // This property will be set by DebugAuthController before changing scenes.
    public AuthSession? Session { get; set; }

    private Label? _statusLabel;
    private Button? _backButton;

    public override void _Ready()
    {
        _statusLabel = GetNode<Label>("VBoxContainer/StatusLabel");
        _backButton = GetNode<Button>("VBoxContainer/BackButton");

        _backButton.Pressed += OnBackButtonPressed;

        if (Session != null && Session.IsCharacterSelected)
        {
            _statusLabel.Text = $"Character '{Session.SelectedCharacterName}' selected.\nReady to enter world...";
        }
        else
        {
            _statusLabel.Text = "Error: No session or character data was passed.\nReturning to auth screen.";
        }
    }

    private void OnBackButtonPressed()
    {
        // Use the centralized scene flow manager to go back.
        SceneFlow.ToDebugAuth(this);
    }
}