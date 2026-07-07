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
    private readonly DebugChunkStore _chunkStore = new();

    // UI Node references
    private Label? _statusLabel;
    private Button? _backButton;
    private Label? _isAuthenticatedValueLabel;
    private Label? _isCharacterSelectedValueLabel;
    private Label? _accountIdValueLabel;
    private Label? _selectedCharacterNameValueLabel;
    private TextEdit? _packetLogTextEdit;
    private Button? _sendMoveButton;
    private Label? _lastActionResultLabel; // Renamed for clarity
    private Button? _attackOrcEliteButton;
    private DebugTileWorldView? _worldView;

    // Snapshot UI Node references
    private Label? _invSyncValueLabel;
    private Label? _levelValueLabel;
    private Label? _hpValueLabel;
    private Label? _manaValueLabel;
    private Label? _chunksValueLabel;
    private Label? _lastChunkValueLabel;
    private Label? _lastTimestampValueLabel;

    // Move State UI Node references
    private Label? _confirmedPosValueLabel;
    private Label? _lastTargetValueLabel;
    private Label? _movePendingValueLabel;

    // Local state for debug movement
    private (int x, int y, int z) _currentConfirmedPos = (103, 102, 0);
    private string _selectedCharacterNameForWorldEntry = string.Empty;
    private bool _initialPositionSet = false;
    private Vector2I? _orcElitePosition;
    private bool _isOrcEliteDead = false;
    private string _orcEliteRuntimeEntityId = string.Empty;

    private bool _isMovePending = false;
    private Vector2I? _lastSentTarget;

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
        _lastActionResultLabel = GetNode<Label>("VBoxContainer/LastMoveResultLabel"); // Renamed for clarity
        _attackOrcEliteButton = GetNode<Button>("VBoxContainer/AttackOrcEliteButton");
        _worldView = GetNode<DebugTileWorldView>("VBoxContainer/DebugTileWorldView");

        // Get snapshot node references
        _invSyncValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/InvSyncValueLabel");
        _levelValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/LevelValueLabel");
        _hpValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/HPValueLabel");
        _manaValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/ManaValueLabel");
        _chunksValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/ChunksValueLabel");
        _lastChunkValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/LastChunkValueLabel");
        _lastTimestampValueLabel = GetNode<Label>("VBoxContainer/SnapshotGridContainer/LastTimestampValueLabel");

        // Get move state node references
        _confirmedPosValueLabel = GetNode<Label>("VBoxContainer/MoveStateGridContainer/ConfirmedPosValueLabel");
        _lastTargetValueLabel = GetNode<Label>("VBoxContainer/MoveStateGridContainer/LastTargetValueLabel");
        _movePendingValueLabel = GetNode<Label>("VBoxContainer/MoveStateGridContainer/MovePendingValueLabel");

        _backButton.Pressed += OnBackButtonPressed;
        _sendMoveButton.Pressed += OnSendMoveButtonPressed;
        _attackOrcEliteButton.Pressed += OnAttackOrcEliteButtonPressed;

        // Pass the chunk store to the view
        _worldView!.ChunkStore = _chunkStore;

        // Set initial player marker position
        _worldView.PlayerTilePosition = new Vector2I(_currentConfirmedPos.x, _currentConfirmedPos.y);

        // Set initial move state UI
        _confirmedPosValueLabel!.Text = $"({_currentConfirmedPos.x}, {_currentConfirmedPos.y}, {_currentConfirmedPos.z})";

        // Populate UI with session data
        if (Session != null)
        {
            _isAuthenticatedValueLabel.Text = Session.IsAuthenticated.ToString();
            _isCharacterSelectedValueLabel.Text = Session.IsCharacterSelected.ToString();
            _accountIdValueLabel.Text = Session.AccountId.ToString();
            _selectedCharacterNameValueLabel.Text = Session.SelectedCharacterName;
            _selectedCharacterNameForWorldEntry = Session.SelectedCharacterName;

            // Setup the packet router
            _router = new DebugIncomingPacketRouter();
            _router.RegisterHandler(4001, OnInventorySyncReceived);
            _router.RegisterHandler(2006, OnChunkDataReceived);
            _router.RegisterHandler(2005, OnMoveConfirmReceived);
            _router.RegisterHandler(2001, OnPlayerUpdateReceived);
            _router.RegisterHandler(3002, OnDamageEventReceived);
            _router.RegisterHandler(3003, OnTargetDeadEventReceived);
            _router.RegisterHandler(3004, OnCreatureRespawnEventReceived);
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

    public override void _UnhandledInput(InputEvent @event)
    {
        if (@event is InputEventKey keyEvent && keyEvent.Pressed && !keyEvent.IsEcho())
        {
            if (_isMovePending)
            {
                return; // Ignore new movement input while one is pending
            }

            var (deltaX, deltaY) = keyEvent.Keycode switch
            {
                Key.W or Key.Up => (0, -1),
                Key.S or Key.Down => (0, 1),
                Key.A or Key.Left => (-1, 0),
                Key.D or Key.Right => (1, 0),
                _ => (0, 0)
            };

            if (deltaX != 0 || deltaY != 0)
            {
                _ = SendDebugMoveAsync(deltaX, deltaY, "keyboard");
                GetViewport().SetInputAsHandled();
            }
        }
    }

    private void OnSendMoveButtonPressed()
    {
        _ = SendDebugMoveAsync(1, 0, "button");
    }

    // Debug-only technical combat validation. Not final gameplay combat input.
    private async void OnAttackOrcEliteButtonPressed()
    {
        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            LogPacketInfo("Cannot send attack: Not connected.");
            SetActionResultText("Last Action Result: cannot send attack - not connected");
            return;
        }

        if (Session == null)
        {
            LogPacketInfo("Cannot send attack: Session is null.");
            SetActionResultText("Last Action Result: cannot send attack - session null");
            return;
        }
        if (string.IsNullOrWhiteSpace(_selectedCharacterNameForWorldEntry))
        {
            LogPacketInfo("Cannot send attack: cached selected character name is empty.");
            SetActionResultText("Last Action Result: cannot send attack - cached selected character empty");
            return;
        }

        if (_cts == null || _cts.IsCancellationRequested)
        {
            LogPacketInfo("Cannot send attack: Packet listener is not active.");
            return;
        }

        var logMessage = new StringBuilder();
        var sessionAuth = Session?.IsAuthenticated ?? false;
        var sessionCharSelected = Session?.IsCharacterSelected ?? false;
        var sessionCharName = Session?.SelectedCharacterName ?? "[null]";
        LogPacketInfo(
            $"Debug attack session state: Auth={sessionAuth}, CharSelected={sessionCharSelected}, " +
            $"CurrentSessionName='{sessionCharName}', CachedWorldEntryName='{_selectedCharacterNameForWorldEntry}'"
        );
        var attackTargetId = string.IsNullOrWhiteSpace(_orcEliteRuntimeEntityId)
            ? "Orc_Elite"
            : _orcEliteRuntimeEntityId;
        var attackTargetMode = attackTargetId == "Orc_Elite" ? "fallback-static" : "runtime-entity";

        logMessage.AppendLine("[SEND] Opcode: 3000 (CS_ATTACK_REQUEST)");
        logMessage.AppendLine($"  Target: {attackTargetId}");
        logMessage.AppendLine($"  TargetMode: {attackTargetMode}");
        logMessage.AppendLine("  WeaponType: debug_sword");
        if (_isOrcEliteDead)
        {
            logMessage.AppendLine("  RetryFlow: Orc_Elite is locally dead; next attack requests server-side debug respawn.");
        }
        LogPacketInfo(logMessage.ToString());

        try
        {
            var wasOrcEliteDeadBeforeAttack = _isOrcEliteDead;
            await GatewayClient.SendAttackRequestAsync(attackTargetId, "debug_sword", _cts.Token);

            if (wasOrcEliteDeadBeforeAttack)
            {
                _isOrcEliteDead = false;
                if (_worldView != null)
                {
                    _worldView.IsOrcEliteDead = false;
                    _worldView.QueueRedraw();
                }

                SetActionResultText($"Last Action Result: retry attack sent to {attackTargetId}; Orc_Elite respawn requested");
                LogPacketInfo("Debug retry flow: Orc_Elite local visual state reset after attack send.");
            }
            else
            {
                SetActionResultText($"Last Action Result: attack request sent to {attackTargetId}");
            }
        }
        catch (Exception ex)
        {
            LogPacketInfo($"Error sending attack request: {ex.Message}");
            SetActionResultText($"Last Action Result: send error - {ex.GetType().Name}");
        }
    }

    private async Task SendDebugMoveAsync(int deltaX, int deltaY, string source)
    {
        if (_isMovePending)
        {
            LogPacketInfo($"Cannot send move from {source}: a move is already pending confirmation.");
            return;
        }

        SetActionResultText($"Last Move Result: move initiated by {source}");

        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            LogPacketInfo("Cannot send move: Not connected.");
            SetActionResultText("Last Move Result: cannot send - not connected");
            return;
        }
        if (_cts == null || _cts.IsCancellationRequested)
        {
            LogPacketInfo("Cannot send move: Packet listener is not active.");
            SetActionResultText("Last Move Result: cannot send - listener inactive");
            return;
        }

        var targetX = _currentConfirmedPos.x + deltaX;
        var targetY = _currentConfirmedPos.y + deltaY;
        var targetZ = (sbyte)_currentConfirmedPos.z;
        var clientTimestamp = (ulong)DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();
        _lastSentTarget = new Vector2I(targetX, targetY);

        // Enter pending state
        _isMovePending = true;
        _sendMoveButton!.Disabled = true;
        _movePendingValueLabel!.Text = "Yes";
        _lastTargetValueLabel!.Text = $"({targetX}, {targetY})";
        _worldView!.TargetPosition = _lastSentTarget;
        _worldView.QueueRedraw();

        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[SEND] Opcode: 2004 (CS_MOVE_REQUEST)");
        logMessage.AppendLine($"  Target: ({targetX}, {targetY}, {targetZ})");
        LogPacketInfo(logMessage.ToString());
        SetActionResultText($"Last Move Result: sending move request to ({targetX}, {targetY}, {targetZ})");

        try
        {
            await GatewayClient.SendMoveRequestAsync(targetX, targetY, targetZ, 0, clientTimestamp, _cts.Token);
            SetActionResultText($"Last Move Result: move request sent to ({targetX}, {targetY}, {targetZ}), waiting confirm");
        }
        catch (OperationCanceledException)
        {
            LogPacketInfo("Move request was canceled.");
            SetActionResultText("Last Move Result: request canceled");
            CallDeferred(nameof(ResetMoveState));
        }
        catch (Exception ex)
        {
            LogPacketInfo($"Error sending move request: {ex.Message}");
            SetActionResultText($"Last Move Result: send error - {ex.GetType().Name}");
            CallDeferred(nameof(ResetMoveState));
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
                CallDeferred(nameof(LogPacketInfo), $"[RAW RECV] Opcode: {packet.Opcode}, Size: {packet.Size}, Seq: {packet.Sequence}");

                if (packet.Opcode == 3003)
                {
                    CallDeferred(nameof(SetActionResultText), "Last Action Result: RAW 3003 received from Gateway");
                }

                try
                {
                    _router?.Dispatch(packet);
                }
                catch (Exception ex)
                {
                    // This prevents a faulty handler from crashing the entire packet listener loop.
                    CallDeferred(nameof(LogPacketInfo), $"Packet handler error for opcode {packet.Opcode}: {ex.Message}");
                }
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
        _chunkStore.AddChunk(chunkData.ChunkX, chunkData.ChunkY, chunkData.Tiles);
        _snapshot.UpdateFromChunkData(chunkData);

        logMessage.AppendLine("  Type: Chunk Data");
        logMessage.AppendLine($"  Chunk Coords: ({chunkData.ChunkX}, {chunkData.ChunkY})");
        logMessage.AppendLine($"  Tiles: {chunkData.Tiles.Length} bytes");
        logMessage.AppendLine($"  Total Chunks Received: {_snapshot.TotalChunksReceived}");
        logMessage.AppendLine("  > Snapshot updated.");
        CallDeferred(nameof(RequestWorldViewRedraw));
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
            // Always apply the server-confirmed position.
            // On Success=false, this is an authoritative rubberband correction.
            _currentConfirmedPos = ((int)Math.Round(data.X), (int)Math.Round(data.Y), data.Z);

            if (data.Success)
            {
                logMessage.AppendLine("  > Applied accepted authoritative move.");
            }
            else
            {
                logMessage.AppendLine("  > Applied authoritative rubberband correction.");
            }

            // Safely schedule the visual update for the world view.
            CallDeferred(nameof(UpdateWorldViewMarkers));
        }
        else
        {
            logMessage.AppendLine("  Error: Failed to decode JSON payload.");
        }

        var resultText = data != null ? $"success={data.Success} confirmed=({data.X:F2}, {data.Y:F2}, {data.Z}) seq={data.Seq}" : "Error: payload decode failed";
        CallDeferred(nameof(SetActionResultText), "Last Move Result: " + resultText);

        // Always reset the pending state after receiving a response.
        CallDeferred(nameof(ResetMoveState));

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

            // Check if this update is for our own player (e.g., initial position sync)
            if (!string.IsNullOrEmpty(_selectedCharacterNameForWorldEntry) && data.PlayerID == _selectedCharacterNameForWorldEntry)
            {
                logMessage.AppendLine("  > Applied authoritative position for local player.");
                _currentConfirmedPos = ((int)Math.Round(data.X), (int)Math.Round(data.Y), data.Z);

                // If this is the first time we get the player's position, calculate the debug Orc's position.
                if (!_initialPositionSet)
                {
                    _orcElitePosition = new Vector2I(_currentConfirmedPos.x + 2, _currentConfirmedPos.y + 2);
                    _initialPositionSet = true;
                }

                CallDeferred(nameof(UpdateWorldViewMarkers));
                CallDeferred(nameof(UpdateConfirmedPositionLabel));
            }
        }
        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnDamageEventReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode} (SC_DAMAGE_EVENT), Size: {packet.Size}");

        try
        {
            var data = BinaryProtocol.DecodeDamageEvent(packet.Payload);
            logMessage.AppendLine($"  Attacker: {data.AttackerID}, Target: {data.TargetID}");
            logMessage.AppendLine($"  Damage: {data.Damage:F2}, Crit: {data.IsCrit}, Hit: {data.IsHit}, Success: {data.Success}");
            logMessage.AppendLine($"  Skill: '{data.SkillName}'");
        }
        catch (Exception ex)
        {
            logMessage.AppendLine($"  Error decoding DamageEvent: {ex.Message}");
        }

        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnTargetDeadEventReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode} (SC_TARGET_DEAD), Size: {packet.Size}");

        try
        {
            var data = BinaryProtocol.DecodeTargetDeadEvent(packet.Payload);
            CallDeferred(nameof(SetActionResultText), $"Last Action Result: target dead received - {data.TargetID} (3003)");
            logMessage.AppendLine($"  Target Dead: {data.TargetID}");

            // If the specific debug target is dead, update its state and redraw the view.
            if (data.TargetID == "Orc_Elite" || (!string.IsNullOrWhiteSpace(data.RuntimeEntityID) && data.RuntimeEntityID == _orcEliteRuntimeEntityId))
            {
                _isOrcEliteDead = true;
                if (_worldView != null)
                {
                    _worldView.IsOrcEliteDead = true;
                    _worldView.QueueRedraw();
                }
            }
        }
        catch (Exception ex)
        {
            logMessage.AppendLine($"  Error decoding TargetDeadEvent: {ex.Message}");
        }

        CallDeferred(nameof(LogPacketInfo), logMessage.ToString());
    }

    private void OnCreatureRespawnEventReceived(Packet packet)
    {
        var logMessage = new StringBuilder();
        logMessage.AppendLine($"[RECV] Opcode: {packet.Opcode} (SC_CREATURE_RESPAWN), Size: {packet.Size}");

        try
        {
            var data = BinaryProtocol.DecodeCreatureRespawnEvent(packet.Payload);
            CallDeferred(nameof(SetActionResultText), $"Last Action Result: creature respawn received - {data.TargetID} runtime={data.RuntimeEntityID} (3004)");
            logMessage.AppendLine($"  Creature Respawn: {data.TargetID}");
            logMessage.AppendLine($"  RuntimeEntityID: {data.RuntimeEntityID}");

            if (data.TargetID == "Orc_Elite" || (!string.IsNullOrWhiteSpace(data.RuntimeEntityID) && data.RuntimeEntityID == _orcEliteRuntimeEntityId))
            {
                _orcEliteRuntimeEntityId = data.RuntimeEntityID;
                _isOrcEliteDead = false;
                if (_worldView != null)
                {
                    _worldView.IsOrcEliteDead = false;
                    _worldView.QueueRedraw();
                }

                logMessage.AppendLine("  Orc_Elite visual state reset to alive/red.");
            }
        }
        catch (Exception ex)
        {
            logMessage.AppendLine($"  Error decoding CreatureRespawnEvent: {ex.Message}");
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

    private void SetActionResultText(string text) // Renamed for clarity
    {
        _lastActionResultLabel!.Text = text;
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

    private void RequestWorldViewRedraw()
    {
        _worldView?.QueueRedraw();
    }

    private void UpdateWorldViewMarkers()
    {
        if (_worldView != null)
        {
            _worldView.PlayerTilePosition = new Vector2I(_currentConfirmedPos.x, _currentConfirmedPos.y);
            _worldView.OrcElitePosition = _orcElitePosition;
            _worldView.IsOrcEliteDead = _isOrcEliteDead;
            _worldView.QueueRedraw();
        }
    }

    private void ResetMoveState()
    {
        _isMovePending = false;
        _sendMoveButton!.Disabled = false;
        _movePendingValueLabel!.Text = "No";
        _worldView!.TargetPosition = null;
        _worldView.QueueRedraw();
        UpdateConfirmedPositionLabel();
    }

    private void UpdateConfirmedPositionLabel()
    {
        _confirmedPosValueLabel!.Text = $"({_currentConfirmedPos.x}, {_currentConfirmedPos.y}, {_currentConfirmedPos.z})";
    }
}

