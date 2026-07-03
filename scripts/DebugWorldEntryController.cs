using Godot;
using LightAndShadow.Client;
using System;

public partial class DebugWorldEntryController : Control
{
    // This property will be set by DebugAuthController before changing scenes.
    public AuthSession? Session { get; set; }

    // UI Node references
    private Label? _statusLabel;
    private Button? _backButton;
    private Label? _isAuthenticatedValueLabel;
    private Label? _isCharacterSelectedValueLabel;
    private Label? _accountIdValueLabel;
    private Label? _selectedCharacterNameValueLabel;

    public override void _Ready()
    {
        // Get node references
        _statusLabel = GetNode<Label>("VBoxContainer/StatusLabel");
        _backButton = GetNode<Button>("VBoxContainer/BackButton");
        _isAuthenticatedValueLabel = GetNode<Label>("VBoxContainer/GridContainer/IsAuthenticatedValueLabel");
        _isCharacterSelectedValueLabel = GetNode<Label>("VBoxContainer/GridContainer/IsCharacterSelectedValueLabel");
        _accountIdValueLabel = GetNode<Label>("VBoxContainer/GridContainer/AccountIdValueLabel");
        _selectedCharacterNameValueLabel = GetNode<Label>("VBoxContainer/GridContainer/SelectedCharacterNameValueLabel");

        _backButton.Pressed += OnBackButtonPressed;

        // Populate UI with session data
        if (Session != null)
        {
            _isAuthenticatedValueLabel.Text = Session.IsAuthenticated.ToString();
            _isCharacterSelectedValueLabel.Text = Session.IsCharacterSelected.ToString();
            _accountIdValueLabel.Text = Session.AccountId.ToString();
            _selectedCharacterNameValueLabel.Text = Session.SelectedCharacterName;
        }
        else
        {
            _statusLabel.Text = "Error: No session data was passed.";
            _isAuthenticatedValueLabel.Text = "N/A";
            _isCharacterSelectedValueLabel.Text = "N/A";
            _accountIdValueLabel.Text = "N/A";
            _selectedCharacterNameValueLabel.Text = "N/A";
        }
    }

    private void OnBackButtonPressed()
    {
        // Use the centralized scene flow manager to go back.
        SceneFlow.ToDebugAuth(this);
    }
}