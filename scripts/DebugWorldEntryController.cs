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

    // UI Node references
    private Label? _statusLabel;
    private Button? _backButton;
    private Label? _isAuthenticatedValueLabel;
    private Label? _isCharacterSelectedValueLabel;
    private Label? _accountIdValueLabel;
    private Label? _selectedCharacterNameValueLabel;
    private TextEdit? _packetLogTextEdit;

    private int _chunksReceived = 0;

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

        _backButton.Pressed += OnBackButtonPressed;

        // Populate UI with session data
        if (Session != null)
        {
            _isAuthenticatedValueLabel.Text = Session.IsAuthenticated.ToString();
            _isCharacterSelectedValueLabel.Text = Session.IsCharacterSelected.ToString();
            _accountIdValueLabel.Text = Session.AccountId.ToString();
            _selectedCharacterNameValueLabel.Text = Session.SelectedCharacterName;

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
                HandlePacket(packet);
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

    private void HandlePacket(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode}, Size: {packet.Size}");

        switch (packet.Opcode)
        {
            case 4001: // SC_INVENTORY_SYNC
                var inventoryData = BinaryProtocol.DecodeInventorySync(packet.Payload);
                logMessage.AppendLine("  Type: Inventory Sync");
                logMessage.AppendLine($"  Item Count: {inventoryData.Items.Count}");
                logMessage.AppendLine($"  Level: {inventoryData.Level}");
                logMessage.AppendLine($"  HP: {inventoryData.Health:F2} / {inventoryData.MaxHealth:F2}");
                logMessage.AppendLine($"  Mana: {inventoryData.Mana:F2} / {inventoryData.MaxMana:F2}");
                break;

            case 2006: // SC_CHUNK_DATA
                var chunkData = BinaryProtocol.DecodeChunkData(packet.Payload);
                _chunksReceived++;
                logMessage.AppendLine("  Type: Chunk Data");
                logMessage.AppendLine($"  Chunk Coords: ({chunkData.ChunkX}, {chunkData.ChunkY})");
                logMessage.AppendLine($"  Tiles: {chunkData.Tiles.Length} bytes");
                logMessage.AppendLine($"  Total Chunks Received: {_chunksReceived}");
                break;

            default:
                logMessage.AppendLine($"  Type: Unknown (Opcode {packet.Opcode})");
                break;
        }

        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void LogPacketInfo(string message)
    {
        _packetLogTextEdit!.Text += message + "\n";
    }
}