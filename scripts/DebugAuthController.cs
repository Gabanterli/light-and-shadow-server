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
    private Button? _openAlphaShellButton;
    private Label? _statusLabel;
    private TextEdit? _logTextEdit;

    // R1-N-E: Character Creation UI Nodes
    private LineEdit? _createCharacterNameLineEdit;
    private OptionButton? _createCharacterRaceOptionButton;
    private Button? _createCharacterButton;

    private GatewayTcpClient _gatewayClient = null!;

    private readonly AuthSession _authSession = new();

    private readonly List<string> _characterNames = new();

    private bool _isTransferringGatewayClientToWorldEntry = false;

    public override void _Ready()
    {
        // Get nodes from the scene tree
        _usernameLineEdit = GetNode<LineEdit>("VBoxContainer/HBoxContainer/UsernameLineEdit");
        _passwordLineEdit = GetNode<LineEdit>("VBoxContainer/HBoxContainer/PasswordLineEdit");
        _loginButton = GetNode<Button>("VBoxContainer/LoginButton");
        _requestCharactersButton = GetNode<Button>("VBoxContainer/RequestCharactersButton");
        _characterList = GetNode<ItemList>("VBoxContainer/CharacterList");
        _selectCharacterButton = GetNode<Button>("VBoxContainer/SelectCharacterButton");
        _openAlphaShellButton = GetNode<Button>("VBoxContainer/OpenAlphaShellButton");
        _statusLabel = GetNode<Label>("VBoxContainer/StatusLabel");
        _logTextEdit = GetNode<TextEdit>("VBoxContainer/LogTextEdit");

        // R1-N-E: Get Character Creation Nodes
        _createCharacterNameLineEdit = GetNode<LineEdit>("VBoxContainer/CreateCharacterHBox/CreateCharacterNameLineEdit");
        _createCharacterRaceOptionButton = GetNode<OptionButton>("VBoxContainer/CreateCharacterHBox/CreateCharacterRaceOptionButton");
        _createCharacterButton = GetNode<Button>("VBoxContainer/CreateCharacterHBox/CreateCharacterButton");

        ConfigureCharacterRaceOptions();

        // Connect signals
        _loginButton.Pressed += OnLoginButtonPressed;
        _requestCharactersButton.Pressed += OnRequestCharactersButtonPressed;
        _selectCharacterButton.Pressed += OnSelectCharacterButtonPressed;
        _openAlphaShellButton.Pressed += OnOpenAlphaShellButtonPressed;
        _createCharacterButton.Pressed += OnCreateCharacterButtonPressed;

        // Initial state
        _requestCharactersButton.Disabled = true;
        _selectCharacterButton!.Disabled = true;
        _createCharacterButton.Disabled = true;
        _statusLabel.Text = "Status: Disconnected";

        var gatewayConfig = new GatewayRuntimeConfig();
        _gatewayClient = new GatewayTcpClient(gatewayConfig.Host, gatewayConfig.Port);

        Log($"Gateway config loaded. {gatewayConfig}");
        Log("Debug Auth Controller Initialized. Ready to connect.");
    }

    public override void _ExitTree()
    {
        // Only dispose of the client if we are not handing it off to the next scene.
        if (!_isTransferringGatewayClientToWorldEntry)
        {
            _gatewayClient.Dispose();
        }
        _authSession.Clear();
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
                _authSession.SetLogin(response.AccountId, response.Token);
                _statusLabel.Text = $"Status: Logged in! Account ID: {response.AccountId}";
                Log("Login successful. AuthSession updated.");
                _requestCharactersButton!.Disabled = false;
                _createCharacterButton!.Disabled = false;
            }
            else
            {
                _statusLabel.Text = "Status: Login failed.";
                Log($"Login failed. Error: {response.ErrorCode}");
                _authSession.Clear();
                _createCharacterButton!.Disabled = true;
                _gatewayClient.Disconnect(); // Disconnect on failed login
            }
        }
        catch (Exception ex)
        {
            _statusLabel.Text = "Status: Error.";
            Log($"Exception during login: {ex.Message}");
            _gatewayClient.Disconnect();
            _authSession.Clear();
            _createCharacterButton!.Disabled = true;
        }
        finally
        {
            _loginButton.Disabled = false;
        }
    }

    private async void OnRequestCharactersButtonPressed()
    {
        await RefreshCharacterListAsync();
    }

    private async Task RefreshCharacterListAsync()
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
                    _characterList.AddItem($"{character.Name} (Lvl {character.Level} {character.Class} / {character.RaceId})"); // (R1-I-B)
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
            _authSession.Clear();
            _requestCharactersButton.Disabled = true;
            _createCharacterButton!.Disabled = true;
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

    private async void OnCreateCharacterButtonPressed()
    {
        var desiredName = _createCharacterNameLineEdit?.Text?.Trim() ?? string.Empty;
        if (string.IsNullOrEmpty(desiredName))
        {
            _statusLabel!.Text = "Status: Character name cannot be empty.";
            Log("Error: Character name cannot be empty.");
            return;
        }

        if (_createCharacterRaceOptionButton is null)
        {
            _statusLabel!.Text = "Status: Race selector is missing.";
            Log("Error: Race option button node is missing.");
            return;
        }

        if (_createCharacterRaceOptionButton.Selected < 0)
        {
            _statusLabel!.Text = "Status: No character race selected.";
            Log("Error: No character race selected.");
            return;
        }

        var selectedRace = _createCharacterRaceOptionButton.GetItemText(_createCharacterRaceOptionButton.Selected);
        if (string.IsNullOrEmpty(selectedRace))
        {
            _statusLabel!.Text = "Status: Character race cannot be empty.";
            Log("Error: Character race cannot be empty.");
            return;
        }

        _createCharacterButton!.Disabled = true;
        _statusLabel!.Text = $"Status: Creating character '{desiredName}'...";
        Log($"Creating character '{desiredName}' with race '{selectedRace}'...");

        try
        {
            var response = await _gatewayClient.CreateCharacterAsync(desiredName, selectedRace);

            if (response.Status)
            {
                _statusLabel.Text = $"Status: Character '{response.Character.Name}' created.";
                Log($"Character created: {response.Character.Name} (Lvl {response.Character.Level} {response.Character.Class} / {response.Character.RaceId})");
                _createCharacterNameLineEdit!.Text = string.Empty;

                // Refresh the character list to show the new character
                await RefreshCharacterListAsync();
            }
            else
            {
                _statusLabel.Text = "Status: Character creation failed.";
                Log($"Character creation failed. Error: {response.ErrorCode}");
            }
        }
        catch (Exception ex)
        {
            _statusLabel.Text = "Status: Error.";
            Log($"Exception during character creation: {ex.Message}");
            // A network error during creation is more severe, so we disconnect.
            _gatewayClient.Disconnect();
            _authSession.Clear();
            _requestCharactersButton!.Disabled = true;
            _createCharacterButton!.Disabled = true;
            _selectCharacterButton!.Disabled = true;
        }
        finally
        {
            if (_gatewayClient.IsConnected)
            {
                _createCharacterButton.Disabled = false;
            }
        }
    }

    private void OnOpenAlphaShellButtonPressed()
    {
        Log("Opening Alpha UI shell. Backend integration is pending.");
        SceneFlow.ToAlphaWorldEntryShell(this, _authSession, _gatewayClient);
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
                _authSession.SetSelectedCharacter(response.CharacterName);
                _statusLabel.Text = $"Status: Character '{response.CharacterName}' selected!";
                Log($"Successfully selected character. AuthSession updated with '{response.CharacterName}'.");
                // Mark that we are handing off the client, so _ExitTree doesn't dispose it.
                _isTransferringGatewayClientToWorldEntry = true;
                SceneFlow.ToWorldEntry(this, _authSession, _gatewayClient);
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
            _authSession.Clear();
            _requestCharactersButton!.Disabled = true;
            _createCharacterButton!.Disabled = true;
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

    private void ConfigureCharacterRaceOptions()
    {
        if (_createCharacterRaceOptionButton is null)
        {
            return;
        }

        _createCharacterRaceOptionButton.Clear();
        _createCharacterRaceOptionButton.AddItem("human");
        _createCharacterRaceOptionButton.AddItem("forest_elf");
        _createCharacterRaceOptionButton.AddItem("dwarf");
        _createCharacterRaceOptionButton.AddItem("ice_elf");
        _createCharacterRaceOptionButton.AddItem("green_orc");
        _createCharacterRaceOptionButton.Select(0);
    }
}
