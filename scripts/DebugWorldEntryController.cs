using Godot;
using LightAndShadow.Client;
using System;
using System.IO;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

public partial class DebugWorldEntryController : Control
{
    // This property will be set by DebugAuthController before changing scenes.
    public AuthSession? Session { get; set; }
    public GatewayTcpClient? GatewayClient { get; set; }

    private CancellationTokenSource? _cts;
    private DebugIncomingPacketRouter? _router;
    private readonly DebugWorldBootstrapSnapshot _snapshot = new();

    // UI Node references
    private Label? _statusLabel;
    private Button? _backButton;
    private Label? _isAuthenticatedValueLabel;
    private Label? _isCharacterSelectedValueLabel;
    private Label? _accountIdValueLabel;
    private Label? _selectedCharacterNameValueLabel;
    private TextEdit? _packetLogTextEdit;
    private Button? _sendMoveButton;
    private Label? _lastMoveResultLabel;
    
    // Snapshot UI Node references
    private Label? _invSyncValueLabel;
    private Label? _levelValueLabel;
    private Label? _hpValueLabel;
    private Label? _manaValueLabel;
    private Label? _chunksValueLabel;
    private Label? _lastChunkValueLabel;
    private Label? _lastTimestampValueLabel;

    // Local state for debug movement
    private (int x, int y, int z) _currentConfirmedPos = (103, 102, 0);

    public override void _Ready()
    {
        // Get node references
        _statusLabel = GetNode<Label>("VBoxContainer/StatusLabel");
        _backButton = GetNode<Button>("VBoxContainer/BackButton");
        _isAuthenticatedValueLabel = GetNode<Label>("VBoxContainer/GridContainer/IsAuthenticatedValueLabel");
        _isCharacterSelectedValueLabel = GetNode<Label>("VBoxContainer/GridContainer/IsCharacterSelectedValueLabel");
        _accountIdValueLabel = GetNode<Label>("VBoxContainer/GridContainer/AccountIdValueLabel");
        _selectedCharacterNameValueLabel = GetNode<Label>("VBoxContainer/GridContainer/SelectedCharacterNameValueLabel");
        _packetLogTextEdit = GetNode<TextEdit>("VBoxContainer/PacketLogTextEdit");
        _sendMoveButton = GetNode<Button>("VBoxContainer/SendMoveButton");
        _lastMoveResultLabel = GetNode<Label>("VBoxContainer/LastMoveResultLabel");
        
        // Get snapshot node references
        _invSyncValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/InvSyncValueLabel");
        _levelValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/LevelValueLabel");
        _hpValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/HPValueLabel");
        _manaValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/ManaValueLabel");
        _chunksValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/ChunksValueLabel");
        _lastChunkValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/LastChunkValueLabel");
        _lastTimestampValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/LastTimestampValueLabel");

        _backButton.Pressed += OnBackButtonPressed;
        _sendMoveButton.Pressed += OnSendMoveButtonPressed;

        // Populate UI with session data
        if (Session != null)
        {
            _isAuthenticatedValueLabel.Text = Session.IsAuthenticated.ToString();
            _isCharacterSelectedValueLabel.Text = Session.IsCharacterSelected.ToString();
            _accountIdValueLabel.Text = Session.AccountId.ToString();
            _selectedCharacterNameValueLabel.Text = Session.SelectedCharacterName;

            // Setup the packet router
            _router = new DebugIncomingPacketRouter();
            _router.RegisterHandler(4001, OnInventorySyncReceived);
            _router.RegisterHandler(2006, OnChunkDataReceived);
            _router.RegisterHandler(2005, OnMoveConfirmReceived);
            _router.RegisterHandler(2001, OnPlayerUpdateReceived);
            _router.RegisterFallback(OnUnknownPacketReceived);

            StartPacketListenerLoop();
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

    public override void _ExitTree()
    {
        // Ensure the listening loop is stopped and resources are cleaned up.
        _cts?.Cancel();
        _cts?.Dispose();
        GatewayClient?.Dispose();
    }

    private void OnBackButtonPressed()
    {
        // Cancel the listening loop before changing scenes.
        _cts?.Cancel();
        // Use the centralized scene flow manager to go back.
        SceneFlow.ToDebugAuth(this);
    }
    
    private async void OnSendMoveButtonPressed()
    {
        SetMoveResultText("Last Move Result: button clicked");

        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            LogPacketInfo("Cannot send move: Not connected.");
            SetMoveResultText("Last Move Result: cannot send - not connected");
            return;
        }
        if (_cts == null || _cts.IsCancellationRequested)
        {
            LogPacketInfo("Cannot send move: Packet listener is not active.");
            SetMoveResultText("Last Move Result: cannot send - listener inactive");
            return;
        }

        var targetX = _currentConfirmedPos.x + 1;
        var targetY = _currentConfirmedPos.y;
        var targetZ = (sbyte)_currentConfirmedPos.z;
        var clientTimestamp = (ulong)DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();

        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[SEND] Opcode: 2004 (CS_MOVE_REQUEST)");
        logMessage.AppendLine($"  Target: ({targetX}, {targetY}, {targetZ})");
        LogPacketInfo(logMessage.ToString());
        SetMoveResultText($"Last Move Result: sending move request to ({targetX}, {targetY}, {targetZ})");

        try
        {
            await GatewayClient.SendMoveRequestAsync(targetX, targetY, targetZ, 0, clientTimestamp, _cts.Token);
            SetMoveResultText($"Last Move Result: move request sent to ({targetX}, {targetY}, {targetZ}), waiting confirm");
        }
        catch (OperationCanceledException)
        {
            // This is expected if the scene is closing while a send is in progress.
            LogPacketInfo("Move request was canceled.");
        }
        catch (Exception ex)
        {
            LogPacketInfo($"Error sending move request: {ex.Message}");
        }
    }

    private async void StartPacketListenerLoop()
    {
        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            LogPacketInfo("Error: GatewayClient is not available or not connected.");
            return;
        }

        _cts = new CancellationTokenSource();
        LogPacketInfo("Packet listener started. Waiting for server data...");

        try
        {
            while (!_cts.Token.IsCancellationRequested)
            {
                var packet = await GatewayClient.ReceivePacketAsync(_cts.Token);
                _router?.Dispatch(packet);
            }
        }
        catch (OperationCanceledException)
        {
            LogPacketInfo("Packet listener stopped by user.");
        }
        catch (IOException ex)
        {
            LogPacketInfo($"Connection error: {ex.Message}. Listener stopped.");
        }
        catch (Exception ex)
        {
            LogPacketInfo($"An unexpected error occurred: {ex.Message}. Listener stopped.");
        }
        finally
        {
            GatewayClient.Disconnect();
            LogPacketInfo("Disconnected from gateway.");
        }
    }

    private void OnInventorySyncReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode}, Size: {packet.Size}");

        var inventoryData = BinaryProtocol.DecodeInventorySync(packet.Payload);
        _snapshot.UpdateFromInventorySync(inventoryData);

        logMessage.AppendLine("  Type: Inventory Sync");
        logMessage.AppendLine($"  Item Count: {inventoryData.Items.Count}");
        logMessage.AppendLine($"  Level: {inventoryData.Level}");
        logMessage.AppendLine($"  HP: {inventoryData.Health:F2} / {inventoryData.MaxHealth:F2}");
        logMessage.AppendLine($"  Mana: {inventoryData.Mana:F2} / {inventoryData.MaxMana:F2}");
        logMessage.AppendLine("  > Snapshot updated.");
        CallDeferred(nameof(UpdateSnapshotDisplay));
        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnChunkDataReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode}, Size: {packet.Size}");

        var chunkData = BinaryProtocol.DecodeChunkData(packet.Payload);
        _snapshot.UpdateFromChunkData(chunkData);

        logMessage.AppendLine("  Type: Chunk Data");
        logMessage.AppendLine($"  Chunk Coords: ({chunkData.ChunkX}, {chunkData.ChunkY})");
        logMessage.AppendLine($"  Tiles: {chunkData.Tiles.Length} bytes");
        logMessage.AppendLine($"  Total Chunks Received: {_snapshot.TotalChunksReceived}");
        logMessage.AppendLine("  > Snapshot updated.");
        CallDeferred(nameof(UpdateSnapshotDisplay));
        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnMoveConfirmReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode} (SC_MOVE_CONFIRM), Size: {packet.Size}");

        var data = BinaryProtocol.DecodeMoveConfirm(packet.Payload);
        if (data != null)
        {
            logMessage.AppendLine($"  Success: {data.Success}");
            logMessage.AppendLine($"  Confirmed Pos: ({data.X:F2}, {data.Y:F2}, {data.Z})");
            logMessage.AppendLine($"  Sequence Echo: {data.Seq}");

            if (data.Success)
            {
                // Update local debug position to the server-confirmed position
                _currentConfirmedPos = ((int)Math.Round(data.X), (int)Math.Round(data.Y), data.Z);
            }
        }
        else
        {
            logMessage.AppendLine("  Error: Failed to decode JSON payload.");
        }

        var resultText = data != null ? $"success={data.Success} confirmed=({data.X:F2}, {data.Y:F2}, {data.Z}) seq={data.Seq}" : "Error: payload decode failed";
        CallDeferred(nameof(SetMoveResultText), "Last Move Result: " + resultText);

        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnPlayerUpdateReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode} (SC_PLAYER_UPDATE), Size: {packet.Size}");
        var data = BinaryProtocol.DecodePlayerUpdate(packet.Payload);
        if (data != null)
        {
            logMessage.AppendLine($"  Player '{data.PlayerID}' moved to ({data.X:F2}, {data.Y:F2}, {data.Z})");
        }
        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnUnknownPacketReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode}, Size: {packet.Size}");
        logMessage.AppendLine($"  Type: Unknown (Opcode {packet.Opcode})");
        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void SetMoveResultText(string text)
    {
        _lastMoveResultLabel!.Text = text;
    }

    private void LogPacketInfo(string message)
    {
        _packetLogTextEdit!.Text += message + "\n";
    }

    private void UpdateSnapshotDisplay()
    {
        if (!IsNodeReady()) return;

        _invSyncValueLabel!.Text = _snapshot.HasReceivedInventorySync.ToString();
        _levelValueLabel!.Text = _snapshot.Level.ToString();
        _hpValueLabel!.Text = $"{_snapshot.CurrentHealth:F2} / {_snapshot.MaxHealth:F2}";
        _manaValueLabel!.Text = $"{_snapshot.CurrentMana:F2} / {_snapshot.MaxMana:F2}";
        _chunksValueLabel!.Text = _snapshot.TotalChunksReceived.ToString();
        _lastChunkValueLabel!.Text = $"({_snapshot.LastChunkX}, {_snapshot.LastChunkY})";
        _lastTimestampValueLabel!.Text = _snapshot.LastPacketTimestamp.ToString("HH:mm:ss.fff");
    }
}