using Godot;
using LightAndShadow.Client;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

public partial class DebugAuthController : Control
{
    // UI Node references
    private LineEdit? _usernameLineEdit;
    private LineEdit? _passwordLineEdit;
    private Button? _loginButton;
    private Button? _requestCharactersButton;
    private ItemList? _characterList;
    private Button? _selectCharacterButton;
    private Label? _statusLabel;
    private TextEdit? _logTextEdit;

    private GatewayTcpClient _gatewayClient = new("127.0.0.1", 8080);
    
    private readonly List<string> _characterNames = new();

    public override void _Ready()
    {
        // Get nodes from the scene tree
        _usernameLineEdit = GetNode<LineEdit>("VBoxContainer/HBoxContainer/UsernameLineEdit");
        _passwordLineEdit = GetNode<LineEdit>("VBoxContainer/HBoxContainer/PasswordLineEdit");
        _loginButton = GetNode<Button>("VBoxContainer/LoginButton");
        _requestCharactersButton = GetNode<Button>("VBoxContainer/RequestCharactersButton");
        _characterList = GetNode<ItemList>("VBoxContainer/CharacterList");
        _selectCharacterButton = GetNode<Button>("VBoxContainer/SelectCharacterButton");
        _statusLabel = GetNode<Label>("VBoxContainer/StatusLabel");
        _logTextEdit = GetNode<TextEdit>("VBoxContainer/LogTextEdit");

        // Connect signals
        _loginButton.Pressed += OnLoginButtonPressed;
        _requestCharactersButton.Pressed += OnRequestCharactersButtonPressed;
        _selectCharacterButton.Pressed += OnSelectCharacterButtonPressed;

        // Initial state
        _requestCharactersButton.Disabled = true;
        _selectCharacterButton!.Disabled = true;
        _statusLabel.Text = "Status: Disconnected";
        Log("Debug Auth Controller Initialized. Ready to connect.");
    }

    public override void _ExitTree()
    {
        _gatewayClient.Dispose();
    }

    private async void OnLoginButtonPressed()
    {
        var username = _usernameLineEdit?.Text ?? string.Empty;
        var password = _passwordLineEdit?.Text ?? string.Empty;

        if (string.IsNullOrWhiteSpace(username) || string.IsNullOrWhiteSpace(password))
        {
            Log("Error: Username and password cannot be empty.");
            return;
        }

        _loginButton!.Disabled = true;
        _statusLabel!.Text = "Status: Connecting...";
        Log($"Attempting to connect and login as '{username}'...");

        try
        {
            if (!_gatewayClient.IsConnected)
            {
                await _gatewayClient.ConnectAsync();
            }

            var response = await _gatewayClient.LoginAsync(username, password);

            if (response.Status)
            {
                _statusLabel.Text = $"Status: Logged in! Account ID: {response.AccountId}";
                Log("Login successful. Session token received but not printed.");
                _requestCharactersButton!.Disabled = false;
            }
            else
            {
                _statusLabel.Text = "Status: Login failed.";
                Log($"Login failed. Error: {response.ErrorCode}");
                _gatewayClient.Disconnect();
            }
        }
        catch (Exception ex)
        {
            _statusLabel.Text = "Status: Error.";
            Log($"Exception during login: {ex.Message}");
            _gatewayClient.Disconnect();
        }
        finally
        {
            _loginButton.Disabled = false;
        }
    }

    private async void OnRequestCharactersButtonPressed()
    {
        _requestCharactersButton!.Disabled = true;
        _statusLabel!.Text = "Status: Requesting character list...";
        Log("Requesting character list...");

        try
        {
            var response = await _gatewayClient.RequestCharacterListAsync();
            if (response.Status)
            {
                _statusLabel.Text = $"Status: {response.Characters.Count} character(s) found.";
                Log($"Character list received. Count: {response.Characters.Count}");
                _characterNames.Clear();
                _characterList!.Clear();
                foreach (var character in response.Characters)
                {
                    _characterNames.Add(character.Name);
                    _characterList.AddItem($"{character.Name} (Lvl {character.Level} {character.Class})");
                }
                _selectCharacterButton!.Disabled = response.Characters.Count == 0;
            }
            else
            {
                _statusLabel.Text = "Status: Failed to get characters.";
                Log($"Failed to get character list. Error: {response.ErrorCode}");
            }
        }
        catch (Exception ex)
        {
            _statusLabel.Text = "Status: Error.";
            Log($"Exception during character list request: {ex.Message}");
            _gatewayClient.Disconnect();
            _requestCharactersButton.Disabled = true;
            _selectCharacterButton!.Disabled = true;
        }
        finally
        {
            if (_gatewayClient.IsConnected)
            {
                _requestCharactersButton.Disabled = false;
            }
        }
    }

    private async void OnSelectCharacterButtonPressed()
    {
        var selectedIndexes = _characterList!.GetSelectedItems();
        if (selectedIndexes.Length == 0)
        {
            Log("Error: No character selected.");
            return;
        }

        var selectedIndex = selectedIndexes[0];
        if (selectedIndex < 0 || selectedIndex >= _characterNames.Count)
        {
            Log($"Error: Invalid selected index {selectedIndex}.");
            _selectCharacterButton!.Disabled = false;
            return;
        }

        var characterName = _characterNames[selectedIndex];

        _selectCharacterButton!.Disabled = true;
        _statusLabel!.Text = $"Status: Selecting '{characterName}'...";
        Log($"Attempting to select character '{characterName}'...");

        try
        {
            var response = await _gatewayClient.SelectCharacterAsync(characterName);
            if (response.Status)
            {
                _statusLabel.Text = $"Status: Character '{response.CharacterName}' selected!";
                Log($"Successfully selected character: {response.CharacterName}");
            }
            else
            {
                _statusLabel.Text = "Status: Character selection failed.";
                Log($"Character selection failed. Error: {response.ErrorCode}");
            }
        }
        catch (Exception ex)
        {
            _statusLabel.Text = "Status: Error.";
            Log($"Exception during character selection: {ex.Message}");
            _gatewayClient.Disconnect();
            _requestCharactersButton!.Disabled = true;
        }
        finally
        {
            if (_gatewayClient.IsConnected)
            {
                _selectCharacterButton.Disabled = false;
            }
        }
    }

    private void Log(string message)
    {
        var timestamp = DateTime.Now.ToString("HH:mm:ss");
        _logTextEdit!.Text += $"[{timestamp}] {message}\n";
    }
}
