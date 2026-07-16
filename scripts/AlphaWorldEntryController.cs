using Godot;
using LightAndShadow.Client;
using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
public partial class AlphaWorldEntryController : Control
{
    private bool _isAlphaOrcEliteSelected;
    public AuthSession? Session { get; set; }
    public GatewayTcpClient? GatewayClient { get; set; }

    private Button? _backButton;
    private Label? _topBarLabel;
    private Label? _worldStatusLabel;
    private Label? _systemFeedbackLabel;
    private Label? _combatFeedbackLabel;
    private Label? _battleLabel;
    private Label? _backpackLabel;
    private PanelContainer? _alphaSpellbookPanel;
    private Label? _alphaSpellbookStatusLabel;
    private Button? _alphaFireBoltButton;
    private Button? _alphaHolySparkButton;
    private Button? _alphaShadowDartButton;
    private CanvasLayer? _alphaDialogueCanvasLayer;
    private Label? _alphaDevGmBadgeLabel;
    private PanelContainer? _alphaDialoguePanel;
    private Label? _alphaDialogueTitleLabel;
    private Label? _alphaDialogueTextLabel;
    private Button? _alphaDialogueCloseButton;
    private VBoxContainer? _alphaDialogueChoicesContainer;
    private DebugTileWorldView? _worldView;

    private Control? _legacyHudRoot;
    private Control? _editableHudRoot;
    private AlphaTopBarPanel? _editableTopBarPanel;
    private AlphaWorldPanel? _editableWorldPanel;
    private AlphaBattlePanel? _editableBattlePanel;
    private AlphaBackpackPanel? _editableBackpackPanel;
    private AlphaFeedbackLogPanel? _editableCombatLogPanel;
    private AlphaFeedbackLogPanel? _editableSystemLogPanel;

    private const int MaxSystemFeedbackMessages = 5;
    private const int MaxCombatFeedbackMessages = 5;
    private const string AlphaRealAttackWeaponType = "debug_sword";
    private const string AlphaMentorArionNpcId = "npc_class_trainer";
    private static readonly Vector2I AlphaMentorArionTilePosition = new(217, 146);

    private readonly DebugChunkStore _chunkStore = new();
    private readonly Queue<string> _systemFeedbackMessages = new();
    private readonly Queue<string> _combatFeedbackMessages = new();
    private readonly Queue<ChunkData> _pendingChunkData = new();
    private readonly object _pendingChunkDataLock = new();
    private CancellationTokenSource? _packetLoopCts;
    private int _ignoredPacketCount;
    private bool _isAlphaDialogueOpen;
    private string _openAlphaDialogueNpcId = string.Empty;
    private string _openAlphaDialogueNodeId = string.Empty;
    private string _confirmedAlphaClassId = string.Empty;
    private bool _isAlphaDevGM;
    private string _alphaAccountRole = string.Empty;

    private bool _hasInventorySync;
    private uint _syncedLevel;
    private double _syncedHealth;
    private double _syncedMaxHealth;
    private double _syncedMana;
    private double _syncedMaxMana;
    private int _syncedItemCount;
    private bool _hasAlphaProgressionSync;
    private ulong _syncedGold;
    private ulong _syncedExperience;

    private bool _hasWorldChunks;
    private int _syncedChunkCount;

    private string _selectedCharacterNameForWorldEntry = string.Empty;
    private bool _hasLocalPlayerPosition;
    private Vector2I _currentPlayerTilePosition;
    private int _currentPlayerTileZ;

    private bool _isAlphaMovePending;
    private Vector2I? _alphaPendingMoveTarget;
    private DateTime _lastAlphaMoveRequestSentUtc = DateTime.MinValue;
    private static readonly TimeSpan MinimumAlphaMoveRequestInterval = TimeSpan.FromMilliseconds(275);

    private string _alphaBattleTargetState = "Pending backend event";
    private bool _hasAlphaOrcEliteVisualPosition;
    private Vector2I _alphaOrcEliteVisualPosition;
    private string _alphaOrcEliteRuntimeEntityId = string.Empty;
    private bool _pendingCombatRewardConfirmation;
    private CancellationTokenSource? _alphaAutoAttackCts;
    private bool _isAlphaAttackRequestInFlight;
    private static readonly TimeSpan AlphaAutoAttackInterval = TimeSpan.FromMilliseconds(1000);
    private const string LoginOverlapOffensiveActionBlockedReason =
        "Você precisa sair de cima de outro jogador para usar ações ofensivas.";
    private const double AlphaDebugSwordPreviewRangeTiles = 1.05;

    public override void _Ready()
    {
        _legacyHudRoot = GetNodeOrNull<Control>("Root");
        _topBarLabel = GetNodeOrNull<Label>("Root/TopBar/TopBarHBox/TopBarLabel");
        _worldStatusLabel = GetNodeOrNull<Label>("Root/MainArea/WorldPanel/WorldVBox/WorldStatusLabel");
        _systemFeedbackLabel = GetNodeOrNull<Label>("Root/BottomTabs/System");
        _combatFeedbackLabel = GetNodeOrNull<Label>("Root/BottomTabs/Combat");
        _battleLabel = GetNodeOrNull<Label>("Root/MainArea/SideTabs/Battle");
        _backpackLabel = GetNodeOrNull<Label>("Root/MainArea/SideTabs/Backpack");
        _worldView = GetNodeOrNull<DebugTileWorldView>("Root/MainArea/WorldPanel/WorldVBox/AlphaWorldView");
        _backButton = GetNodeOrNull<Button>("Root/TopBar/TopBarHBox/BackButton");
        BindOptionalEditableHudComponents();

        if (_backButton != null)
        {
            _backButton.Pressed += OnBackButtonPressed;
        }

        _selectedCharacterNameForWorldEntry = Session?.IsCharacterSelected == true
            ? Session.SelectedCharacterName
            : string.Empty;

        if (_worldView != null)
        {
            _worldView.ChunkStore = _chunkStore;
            _worldView.UseFocusedViewport = true;
            _worldView.MinimumFocusedViewportTilesWide = 24;
            _worldView.FocusedViewportTilesHigh = 18;
            _worldView.ShowFixedCombatDebugOverlay = false;
            _worldView.UseOneTileEntityMarkers = true;
            _worldView.ShowAlphaCombatReadabilityHud = true;
            _worldView.MentorArionPosition = AlphaMentorArionTilePosition;
            _worldView.MouseFilter = Control.MouseFilterEnum.Stop;
            _worldView.GuiInput += OnAlphaWorldViewGuiInput;
        }

        RefreshTopBarShellState();
        RefreshBattleTargetState();
        RefreshCombatFeedbackState();
        MountAlphaSpellbookShell();
        MountAlphaDialogueWindowShell();
        RefreshBackpackShellState();
        MountAlphaDevGmBadge();
        RefreshAlphaSpellbookShellState();
        RefreshWorldShellState();
        StartAlphaWorldBootstrapPacketLoop();

        GD.Print("AlphaWorldEntryController loaded: world bootstrap packet loop boundary active.");
    }

    public override void _ExitTree()
    {
        StopAlphaAutoAttackLoop("scene exit");
        if (_worldView != null)
        {
            _worldView.GuiInput -= OnAlphaWorldViewGuiInput;
        }

        StopAlphaPacketLoop();
        _packetLoopCts?.Dispose();
        _packetLoopCts = null;
        GatewayClient?.Dispose();
    }

    public override void _UnhandledInput(InputEvent inputEvent)
    {
        if (inputEvent is not InputEventKey keyEvent || !keyEvent.Pressed || keyEvent.IsEcho())
        {
            return;
        }

        var (deltaX, deltaY) = keyEvent.Keycode switch
        {
            Key.W or Key.Up => (0, -1),
            Key.S or Key.Down => (0, 1),
            Key.A or Key.Left => (-1, 0),
            Key.D or Key.Right => (1, 0),
            _ => (0, 0)
        };

        if (deltaX == 0 && deltaY == 0)
        {
            return;
        }

        _ = SendAlphaMoveAsync(deltaX, deltaY, "keyboard");
        GetViewport().SetInputAsHandled();
    }

    private void BindOptionalEditableHudComponents()
    {
        _editableHudRoot =
            GetNodeOrNull<Control>("Root/EditableAlphaHud") ??
            GetNodeOrNull<Control>("Root/EditableHud") ??
            GetNodeOrNull<Control>("EditableAlphaHud") ??
            GetNodeOrNull<Control>("EditableHud") ??
            GetNodeOrNull<Control>("AlphaHudLayout");

        if (_editableHudRoot == null || !_editableHudRoot.Visible)
        {
            _editableHudRoot = null;
            return;
        }

        _editableTopBarPanel =
            _editableHudRoot.GetNodeOrNull<AlphaTopBarPanel>("Root/TopBar") ??
            _editableHudRoot.GetNodeOrNull<AlphaTopBarPanel>("TopBar");

        _editableWorldPanel =
            _editableHudRoot.GetNodeOrNull<AlphaWorldPanel>("Root/Main/WorldPanel") ??
            _editableHudRoot.GetNodeOrNull<AlphaWorldPanel>("Main/WorldPanel");

        _editableBattlePanel =
            _editableHudRoot.GetNodeOrNull<AlphaBattlePanel>("Root/Main/SidePanel/BattlePanel") ??
            _editableHudRoot.GetNodeOrNull<AlphaBattlePanel>("Main/SidePanel/BattlePanel");

        _editableBackpackPanel =
            _editableHudRoot.GetNodeOrNull<AlphaBackpackPanel>("Root/Main/SidePanel/BackpackPanel") ??
            _editableHudRoot.GetNodeOrNull<AlphaBackpackPanel>("Main/SidePanel/BackpackPanel");

        _editableCombatLogPanel =
            _editableHudRoot.GetNodeOrNull<AlphaFeedbackLogPanel>("Root/Logs/CombatLogPanel") ??
            _editableHudRoot.GetNodeOrNull<AlphaFeedbackLogPanel>("Logs/CombatLogPanel");

        _editableSystemLogPanel =
            _editableHudRoot.GetNodeOrNull<AlphaFeedbackLogPanel>("Root/Logs/SystemLogPanel") ??
            _editableHudRoot.GetNodeOrNull<AlphaFeedbackLogPanel>("Logs/SystemLogPanel");

        var editableBackButton =
            _editableHudRoot.GetNodeOrNull<Button>("Root/TopBar/Content/BackButton") ??
            _editableHudRoot.GetNodeOrNull<Button>("TopBar/Content/BackButton");

        if (editableBackButton != null)
        {
            _backButton = editableBackButton;
        }

        if (_editableHudRoot != null)
        {
            _editableHudRoot.MouseFilter = MouseFilterEnum.Pass;
        }

        if (_editableWorldPanel != null)
        {
            _editableWorldPanel.MouseFilter = MouseFilterEnum.Pass;
        }

        if (_editableWorldPanel?.WorldView != null)
        {
            _worldView = _editableWorldPanel.WorldView;
            _legacyHudRoot?.Hide();
        }

        GD.Print("Alpha optional editable HUD bridge bound.");
    }
    private void RefreshAlphaMentorArionDebugPosition()
    {
        if (_worldView == null)
        {
            return;
        }

        _worldView.MentorArionPosition = AlphaMentorArionTilePosition;
    }
    private void RefreshTopBarShellState()
    {
        var sessionState = Session != null ? "session received" : "session missing";
        var characterState = Session?.IsCharacterSelected == true
            ? Session.SelectedCharacterName
            : "pending character selection";
        var clientState = GatewayClient?.IsConnected == true ? "client connected" : "client disconnected";
        var levelState = _hasInventorySync ? _syncedLevel.ToString() : "pending sync";
        var hpState = _hasInventorySync ? $"{_syncedHealth:F0}/{_syncedMaxHealth:F0}" : "pending sync";
        var manaState = _hasInventorySync ? $"{_syncedMana:F0}/{_syncedMaxMana:F0}" : "pending sync";
        var goldState = _hasAlphaProgressionSync ? _syncedGold.ToString() : "pending sync";
        var xpState = _hasAlphaProgressionSync ? _syncedExperience.ToString() : "pending sync";

        if (_topBarLabel != null)
        {
            _topBarLabel.Text = $"Level: {levelState} | XP: {xpState} | Gold: {goldState} | HP: {hpState} | Mana: {manaState} | {sessionState} | {clientState}";
        }

        _editableTopBarPanel?.BindPlayerStatus(
            characterState,
            _hasInventorySync ? _syncedLevel : 0,
            _hasInventorySync ? _syncedHealth : 0,
            _hasInventorySync ? _syncedMaxHealth : 0,
            _hasInventorySync ? _syncedMana : 0,
            _hasInventorySync ? _syncedMaxMana : 0
        );
    }

    private void RefreshBattleTargetState()
    {
        if (_battleLabel != null)
        {
            _battleLabel.Text = $"Battle\n\nTarget: Orc_Elite\nState: {_alphaBattleTargetState}\nHP: real backend only";
        }

        _editableBattlePanel?.BindTargetState("Orc_Elite", _alphaBattleTargetState, _isAlphaOrcEliteSelected);
        RefreshAlphaSpellbookShellState();
    }
    private void RefreshCombatFeedbackState()
    {
        if (_combatFeedbackLabel != null)
        {
            if (_combatFeedbackMessages.Count == 0)
            {
                _combatFeedbackLabel.Text = "Combat\n- No combat events yet\n- Real backend events only";
            }
            else
            {
                var lines = new List<string> { "Combat" };

                foreach (var feedbackMessage in _combatFeedbackMessages)
                {
                    lines.Add($"- {feedbackMessage}");
                }

                _combatFeedbackLabel.Text = string.Join("\n", lines);
            }
        }

        _editableCombatLogPanel?.BindMessages("Combat", _combatFeedbackMessages);
    }
    private void MountAlphaDialogueWindowShell()
    {
        if (_alphaDialoguePanel != null)
        {
            RefreshAlphaDialogueWindowLayout();
            return;
        }

        _alphaDialogueCanvasLayer = new CanvasLayer
        {
            Name = "AlphaDialogueCanvasLayer",
            Layer = 50
        };

        _alphaDialoguePanel = new PanelContainer
        {
            Name = "AlphaDialoguePanel",
            Visible = false,
            MouseFilter = MouseFilterEnum.Stop,
            ZIndex = 100,
            CustomMinimumSize = new Vector2(560.0f, 210.0f)
        };

        var margin = new MarginContainer
        {
            Name = "Margin"
        };

        margin.AddThemeConstantOverride("margin_left", 14);
        margin.AddThemeConstantOverride("margin_top", 12);
        margin.AddThemeConstantOverride("margin_right", 14);
        margin.AddThemeConstantOverride("margin_bottom", 12);

        var content = new VBoxContainer
        {
            Name = "Content",
            SizeFlagsHorizontal = SizeFlags.ExpandFill,
            SizeFlagsVertical = SizeFlags.ExpandFill
        };

        _alphaDialogueTitleLabel = new Label
        {
            Name = "TitleLabel",
            Text = "Dialogue",
            HorizontalAlignment = HorizontalAlignment.Center
        };

        _alphaDialogueTextLabel = new Label
        {
            Name = "TextLabel",
            Text = "No dialogue open.",
            AutowrapMode = TextServer.AutowrapMode.WordSmart,
            SizeFlagsHorizontal = SizeFlags.ExpandFill
        };


        _alphaDialogueCloseButton = new Button
        {
            Name = "CloseButton",
            Text = "Fechar",
            MouseFilter = MouseFilterEnum.Stop,
            SizeFlagsHorizontal = SizeFlags.ShrinkEnd
        };

        _alphaDialogueCloseButton.Pressed += CloseAlphaDialogueWindowFromButton;
        _alphaDialogueChoicesContainer = new VBoxContainer
        {
            Name = "ChoicesContainer",
            SizeFlagsHorizontal = SizeFlags.ExpandFill,
            SizeFlagsVertical = SizeFlags.ExpandFill
        };

        content.AddChild(_alphaDialogueTitleLabel);
                content.AddChild(_alphaDialogueCloseButton);
content.AddChild(_alphaDialogueTextLabel);
        content.AddChild(_alphaDialogueChoicesContainer);
        margin.AddChild(content);
        _alphaDialoguePanel.AddChild(margin);
        _alphaDialogueCanvasLayer.AddChild(_alphaDialoguePanel);
        AddChild(_alphaDialogueCanvasLayer);

        RefreshAlphaDialogueWindowLayout();

        GD.Print("Alpha dialogue window mounted on CanvasLayer overlay.");
    }

    private void RefreshAlphaDialogueWindowLayout()
    {
        if (_alphaDialoguePanel == null)
        {
            return;
        }

        var viewportSize = GetViewportRect().Size;
        var worldRect = _worldView != null
            ? _worldView.GetGlobalRect()
            : new Rect2(Vector2.Zero, viewportSize);

        if (worldRect.Size.X <= 0.0f || worldRect.Size.Y <= 0.0f)
        {
            worldRect = new Rect2(Vector2.Zero, viewportSize);
        }

        var panelWidth = Math.Min(760.0f, Math.Max(560.0f, worldRect.Size.X * 0.58f));
        var panelHeight = Math.Min(360.0f, Math.Max(260.0f, worldRect.Size.Y * 0.42f));
        var panelX = worldRect.Position.X + Math.Max(0.0f, (worldRect.Size.X - panelWidth) * 0.5f);
        var panelY = worldRect.Position.Y + Math.Max(0.0f, (worldRect.Size.Y - panelHeight) * 0.5f);

        _alphaDialoguePanel.SetAnchorsPreset(Control.LayoutPreset.TopLeft);
        _alphaDialoguePanel.OffsetLeft = panelX;
        _alphaDialoguePanel.OffsetTop = panelY;
        _alphaDialoguePanel.OffsetRight = panelX + panelWidth;
        _alphaDialoguePanel.OffsetBottom = panelY + panelHeight;
        _alphaDialoguePanel.Position = new Vector2(panelX, panelY);
        _alphaDialoguePanel.Size = new Vector2(panelWidth, panelHeight);
        _alphaDialoguePanel.CustomMinimumSize = new Vector2(panelWidth, panelHeight);

        GD.Print($"Alpha dialogue window centered on world view: worldRect={worldRect}, position={_alphaDialoguePanel.Position}, size={_alphaDialoguePanel.Size}");
    }

    private void MountAlphaDevGmBadge()
    {
        if (_alphaDevGmBadgeLabel != null)
        {
            return;
        }

        _alphaDevGmBadgeLabel = new Label
        {
            Name = "AlphaDevGmBadgeLabel",
            Text = "GM TEST",
            Visible = false,
            HorizontalAlignment = HorizontalAlignment.Right,
            VerticalAlignment = VerticalAlignment.Top,
            AnchorsPreset = (int)LayoutPreset.TopRight,
            ZIndex = 200
        };

        AddChild(_alphaDevGmBadgeLabel);
        GD.Print("Alpha Dev GM badge mounted.");
    }

    private void RefreshAlphaDevGmBadge()
    {
        if (_alphaDevGmBadgeLabel != null)
        {
            _alphaDevGmBadgeLabel.Visible = _isAlphaDevGM;
        }
    }

    private void BindAlphaDialogueWindow(string npcId, string nodeId, string nodeText, string choicesText, int choiceCount)
    {
        MountAlphaDialogueWindowShell();

        if (_alphaDialoguePanel == null)
        {
            GD.PrintErr("Alpha dialogue window skipped: panel missing.");
            return;
        }

        RefreshAlphaDialogueWindowLayout();

        _isAlphaDialogueOpen = true;
        _openAlphaDialogueNpcId = npcId;
        _openAlphaDialogueNodeId = nodeId;
        _alphaDialoguePanel.Visible = true;
        _alphaDialoguePanel.Show();

        if (_alphaDialogueTitleLabel != null)
        {
            _alphaDialogueTitleLabel.Text = $"Mentor Arion | {nodeId}";
        }

        if (_alphaDialogueTextLabel != null)
        {
            _alphaDialogueTextLabel.Text = nodeText;
        }

        if (_alphaDialogueChoicesContainer == null)
        {
            GD.PrintErr("Alpha dialogue window choices skipped: container missing.");
            return;
        }

        foreach (var child in _alphaDialogueChoicesContainer.GetChildren())
        {
            child.QueueFree();
        }

        if (choiceCount <= 0 || string.IsNullOrWhiteSpace(choicesText) || choicesText == "none")
        {
            _alphaDialogueChoicesContainer.AddChild(new Label
            {
                Text = "No dialogue choices available.",
                AutowrapMode = TextServer.AutowrapMode.WordSmart
            });

            GD.Print($"Alpha dialogue window opened without choices: npc={npcId}, node={nodeId}, text={nodeText}");
            return;
        }

        var choices = choicesText.Split(new[] { " | " }, StringSplitOptions.RemoveEmptyEntries);
        foreach (var rawChoice in choices)
        {
            var optionText = rawChoice.Trim();
            if (string.IsNullOrWhiteSpace(optionText))
            {
                continue;
            }

            var optionButton = new Button
            {
                Text = $"› {optionText}",
                MouseFilter = MouseFilterEnum.Stop,
                SizeFlagsHorizontal = SizeFlags.ExpandFill
            };

            var parsedChoice = ParseAlphaDialogueChoice(optionText);
            var choiceNextNodeId = parsedChoice.NextNodeId;
            var choiceDisplayText = parsedChoice.Text;
            optionButton.Text = $"Choice: {choiceDisplayText}";

            optionButton.Pressed += () =>
            {
                _ = SendAlphaDialogueResponseAsync(choiceNextNodeId, choiceDisplayText);
            };

            _alphaDialogueChoicesContainer.AddChild(optionButton);
        }

        GD.Print($"Alpha dialogue window opened: npc={npcId}, node={nodeId}, choices={choiceCount}, text={nodeText}, rawChoices={choicesText}");
    }

    private static (string NextNodeId, string Text) ParseAlphaDialogueChoice(string rawChoice)
    {
        var safeRawChoice = rawChoice ?? string.Empty;
        var separatorIndex = safeRawChoice.IndexOf(": ", StringComparison.Ordinal);

        if (separatorIndex <= 0)
        {
            var fallback = safeRawChoice.Trim();
            return (fallback, fallback);
        }

        var nextNodeId = safeRawChoice.Substring(0, separatorIndex).Trim();
        var text = safeRawChoice.Substring(separatorIndex + 2).Trim();

        return (nextNodeId, text);
    }

    private async Task SendAlphaDialogueResponseAsync(string nextNodeId, string displayText)
    {
        if (!_isAlphaDialogueOpen)
        {
            SetAlphaSystemMessage("Cannot respond: no dialogue is open.");
            return;
        }

        if (string.IsNullOrWhiteSpace(_openAlphaDialogueNpcId) || string.IsNullOrWhiteSpace(_openAlphaDialogueNodeId))
        {
            SetAlphaSystemMessage("Cannot respond: dialogue state is incomplete.");
            return;
        }

        if (string.IsNullOrWhiteSpace(nextNodeId))
        {
            SetAlphaSystemMessage("Cannot respond: next node id is empty.");
            return;
        }

        if (_packetLoopCts == null || _packetLoopCts.IsCancellationRequested)
        {
            SetAlphaSystemMessage("Cannot respond: listener inactive.");
            return;
        }

        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            SetAlphaSystemMessage("Cannot respond: client disconnected.");
            return;
        }

        try
        {
            SetAlphaSystemMessage($"Sending dialogue response: {_openAlphaDialogueNodeId} -> {nextNodeId}.");
            SetAlphaCombatMessage($"Dialogue option sent: {displayText}");

            await GatewayClient.SendDialogueResponseRequestAsync(
                _openAlphaDialogueNpcId,
                _openAlphaDialogueNodeId,
                nextNodeId,
                _packetLoopCts.Token
            );

            GD.Print($"Alpha dialogue response sent: npc={_openAlphaDialogueNpcId}, node={_openAlphaDialogueNodeId}, next={nextNodeId}, text={displayText}");

            if (string.Equals(nextNodeId, "end", StringComparison.OrdinalIgnoreCase))
            {
                CloseAlphaDialogueWindow("dialogue ended");
            }
        }
        catch (OperationCanceledException)
        {
            SetAlphaSystemMessage("Dialogue response cancelled.");
            GD.Print("Alpha dialogue response cancelled.");
        }
        catch (Exception ex)
        {
            SetAlphaSystemMessage($"Dialogue response failed: {ex.GetType().Name}.");
            GD.PrintErr($"Alpha dialogue response failed: {ex.Message}");
        }
    }

    private void CloseAlphaDialogueWindowFromButton()
    {
        CloseAlphaDialogueWindow("close button");
    }

    private void CloseAlphaDialogueWindow(string reason)
    {
        if (_alphaDialoguePanel != null)
        {
            _alphaDialoguePanel.Hide();
        }

        _isAlphaDialogueOpen = false;
        _openAlphaDialogueNpcId = string.Empty;

        SetAlphaSystemMessage($"Dialogue closed: {reason}.");
        GD.Print($"Alpha dialogue window closed: reason={reason}");
    }

    private void HandleAlphaDialogueClosePacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeDialogueClose(packet.Payload);
            CallDeferred(nameof(ApplyAlphaDialogueClose), data.NpcId, data.ReasonCode, data.Message);
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha DialogueClose decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"DialogueClose decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyAlphaDialogueClose(string npcId, byte reasonCode, string message)
    {
        var feedback = string.IsNullOrWhiteSpace(message) ? "Diálogo encerrado." : message;
        CloseAlphaDialogueWindow(feedback);
    }

    private void RefreshAlphaDialogueProximityState()
    {
        if (!_isAlphaDialogueOpen)
        {
            return;
        }

        var deltaX = Math.Abs(_currentPlayerTilePosition.X - AlphaMentorArionTilePosition.X);
        var deltaY = Math.Abs(_currentPlayerTilePosition.Y - AlphaMentorArionTilePosition.Y);

        if (deltaX > 4 || deltaY > 4)
        {
            CloseAlphaDialogueWindow("moved away from Mentor Arion");
        }
    }

    private void MountAlphaSpellbookShell()
    {
        if (_alphaSpellbookPanel != null)
        {
            return;
        }

        var sidePanel =
            _editableHudRoot?.GetNodeOrNull<Control>("Root/Main/SidePanel") ??
            _editableHudRoot?.GetNodeOrNull<Control>("Main/SidePanel") ??
            GetNodeOrNull<Control>("Root/MainArea/SideTabs");

        if (sidePanel == null)
        {
            GD.PrintErr("Alpha Spellbook shell skipped: side panel node not found.");
            return;
        }

        _alphaSpellbookPanel = new PanelContainer
        {
            Name = "AlphaSpellbookPanel",
            MouseFilter = MouseFilterEnum.Stop,
            SizeFlagsHorizontal = SizeFlags.ExpandFill,
            SizeFlagsVertical = SizeFlags.ShrinkBegin
        };

        var content = new VBoxContainer
        {
            Name = "Content",
            SizeFlagsHorizontal = SizeFlags.ExpandFill
        };

        var titleLabel = new Label
        {
            Name = "TitleLabel",
            Text = "Spellbook",
            HorizontalAlignment = HorizontalAlignment.Center
        };

        _alphaSpellbookStatusLabel = new Label
        {
            Name = "StatusLabel",
            Text = "Select a target, then click a spell. Backend cast request enabled.",
            AutowrapMode = TextServer.AutowrapMode.WordSmart
        };

        _alphaFireBoltButton = CreateAlphaSpellButton("Fire Bolt");
        _alphaHolySparkButton = CreateAlphaSpellButton("Holy Spark");
        _alphaShadowDartButton = CreateAlphaSpellButton("Shadow Dart");

        _alphaFireBoltButton.Pressed += () => OnAlphaSpellbookSpellPressed("Fire Bolt", 1001);
        _alphaHolySparkButton.Pressed += () => OnAlphaSpellbookSpellPressed("Holy Spark", 1002);
        _alphaShadowDartButton.Pressed += () => OnAlphaSpellbookSpellPressed("Shadow Dart", 1003);

        content.AddChild(titleLabel);
        content.AddChild(_alphaSpellbookStatusLabel);
        content.AddChild(_alphaFireBoltButton);
        content.AddChild(_alphaHolySparkButton);
        content.AddChild(_alphaShadowDartButton);

        _alphaSpellbookPanel.AddChild(content);
        sidePanel.AddChild(_alphaSpellbookPanel);

        GD.Print("Alpha Spellbook shell mounted.");
    }

    private static Button CreateAlphaSpellButton(string spellName)
    {
        return new Button
        {
            Text = spellName,
            SizeFlagsHorizontal = SizeFlags.ExpandFill,
            FocusMode = FocusModeEnum.None
        };
    }

    private void RefreshAlphaSpellbookShellState()
    {
        if (_alphaSpellbookStatusLabel == null)
        {
            return;
        }

        var targetState = _isAlphaOrcEliteSelected
            ? $"Target: Orc_Elite ({_alphaBattleTargetState})"
            : $"Target: not selected ({_alphaBattleTargetState})";

        var identityState = HasAlphaSafeTargetIdentity()
            ? "Target identity ready"
            : "Target identity pending";

        _alphaSpellbookStatusLabel.Text = $"{targetState}. {identityState}. Spell cast sends CS_CAST_SKILL 3001.";
    }

    private void OnAlphaSpellbookSpellPressed(string spellName, uint skillId)
    {
        RefreshAlphaSpellbookShellState();

        if (_alphaBattleTargetState == "Dead")
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: target is dead.");
            return;
        }

        if (_alphaBattleTargetState != "Alive")
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: target is not ready.");
            return;
        }

        if (!_isAlphaOrcEliteSelected)
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: select Orc_Elite first.");
            return;
        }

        if (!HasAlphaSafeTargetIdentity())
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: target identity pending.");
            return;
        }

        _ = SendAlphaCastSkillOnceAsync(spellName, skillId);
    }

    private async Task SendAlphaCastSkillOnceAsync(string spellName, uint skillId, CancellationToken cancellationToken = default)
    {
        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: client disconnected.");
            return;
        }

        if (_packetLoopCts == null || _packetLoopCts.IsCancellationRequested)
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: listener inactive.");
            return;
        }

        if (!HasAlphaSafeTargetIdentity())
        {
            SetAlphaCombatMessage($"Cannot cast {spellName}: target identity pending.");
            return;
        }

        var linkedToken = cancellationToken == default
            ? _packetLoopCts.Token
            : cancellationToken;

        var targetX = _hasAlphaOrcEliteVisualPosition ? _alphaOrcEliteVisualPosition.X : 0;
        var targetY = _hasAlphaOrcEliteVisualPosition ? _alphaOrcEliteVisualPosition.Y : 0;

        SetAlphaCombatMessage($"Casting {spellName} [skill {skillId}] on Orc_Elite.");
        SetAlphaSystemMessage($"CS_CAST_SKILL 3001 queued: {spellName} -> {_alphaOrcEliteRuntimeEntityId}");

        try
        {
            await GatewayClient.SendCastSkillRequestAsync(
                skillId,
                _alphaOrcEliteRuntimeEntityId,
                targetX,
                targetY,
                linkedToken
            );

            SetAlphaCombatMessage($"Cast request sent: {spellName}. Waiting for server result.");
            GD.Print($"Alpha spell cast sent: spell={spellName}, skill={skillId}, target={_alphaOrcEliteRuntimeEntityId}, targetX={targetX}, targetY={targetY}");
        }
        catch (OperationCanceledException)
        {
            SetAlphaCombatMessage($"Cast cancelled: {spellName}.");
        }
        catch (Exception ex)
        {
            SetAlphaCombatMessage($"Cast failed locally: {spellName} ({ex.GetType().Name}).");
            GD.PrintErr($"Alpha spell cast send failed: {ex.Message}");
        }
    }

    private void RefreshBackpackShellState()
    {
        var itemCountState = _hasInventorySync ? $"{_syncedItemCount} synced" : "pending sync";

        if (_backpackLabel != null)
        {
            var goldState = _hasAlphaProgressionSync ? _syncedGold.ToString() : "pending sync";
            var xpState = _hasAlphaProgressionSync ? _syncedExperience.ToString() : "pending sync";
            _backpackLabel.Text = $"Backpack\n\nItems: {itemCountState}\nGold: {goldState}\nXP: {xpState}\nReal inventory sync only.";
        }

        _editableBackpackPanel?.BindBackpackSummary(_hasInventorySync ? _syncedItemCount : 0);
    }

    private void SyncAlphaPlayerVitalsHud()
    {
        if (_worldView == null)
        {
            return;
        }

        _worldView.HasPlayerVitals = _hasInventorySync;
        _worldView.PlayerHealth = _syncedHealth;
        _worldView.PlayerMaxHealth = _syncedMaxHealth;
        _worldView.PlayerMana = _syncedMana;
        _worldView.PlayerMaxMana = _syncedMaxMana;
        RequestAlphaWorldViewRedraw();
    }

    private void ApplyAlphaOrcEliteDamageHudFeedback(double damage, bool isCritical, string runtimeEntityId)
    {
        if (_worldView == null)
        {
            return;
        }

        _worldView.ApplyAlphaOrcEliteConfirmedDamage(damage, runtimeEntityId);
        var critText = isCritical ? " CRIT" : string.Empty;
        _worldView.AddAlphaOrcEliteFloatingCombatText($"-{damage:F0}{critText}", isCritical, false);
        RequestAlphaWorldViewRedraw();
    }

    private void ApplyAlphaOrcEliteMissFloatingText()
    {
        if (_worldView == null)
        {
            return;
        }

        _worldView.AddAlphaOrcEliteFloatingCombatText("Miss", false, true);
        RequestAlphaWorldViewRedraw();
    }

    private void ApplyAlphaConfirmedSpellVisualFeedback(string skillName)
    {
        if (_worldView == null || !IsAlphaConfirmedSpellSkillName(skillName))
        {
            return;
        }

        GD.Print($"Alpha confirmed spell visual feedback: {skillName}");
        _worldView.AddAlphaConfirmedSpellVisual(skillName.Trim());
        RequestAlphaWorldViewRedraw();
    }

    private static bool IsAlphaConfirmedSpellSkillName(string? skillName)
    {
        var normalizedSkillName = skillName?.Trim() ?? string.Empty;

        return string.Equals(normalizedSkillName, "Fire Bolt", StringComparison.OrdinalIgnoreCase)
            || string.Equals(normalizedSkillName, "Holy Spark", StringComparison.OrdinalIgnoreCase)
            || string.Equals(normalizedSkillName, "Shadow Dart", StringComparison.OrdinalIgnoreCase);
    }
    private void RefreshWorldShellState()
    {
        if (_worldStatusLabel != null)
        {
            var viewState = _worldView != null ? "world view mounted" : "world view missing";
            var chunkState = _hasWorldChunks ? $"{_syncedChunkCount} chunks synced" : "chunks pending sync";
            var playerMarkerState = _hasLocalPlayerPosition ? "player marker synced" : "player marker pending sync";
            _worldStatusLabel.Text = $"World sync: {chunkState}. {playerMarkerState}. {viewState}. Focused Alpha viewport. Packet loop: InventorySync + world chunks + player position + WASD move confirm.";
        }
    }

    private async void StartAlphaWorldBootstrapPacketLoop()
    {
        if (GatewayClient == null)
        {
            SetAlphaSystemMessage("Alpha world bootstrap listener not started: client missing.");
            return;
        }

        if (!GatewayClient.IsConnected)
        {
            SetAlphaSystemMessage("Alpha world bootstrap listener not started: client disconnected.");
            return;
        }

        if (_packetLoopCts != null && !_packetLoopCts.IsCancellationRequested)
        {
            SetAlphaSystemMessage("Alpha world bootstrap listener already running.");
            return;
        }

        _packetLoopCts = new CancellationTokenSource();
        var token = _packetLoopCts.Token;

        SetAlphaSystemMessage("Alpha world bootstrap listener started. Waiting for inventory, world chunks, player position, target state, and combat feedback.");

        try
        {
            while (!token.IsCancellationRequested)
            {
                var packet = await GatewayClient.ReceivePacketAsync(token);

                if (packet.Opcode == 4001)
                {
                    HandleAlphaInventorySyncPacket(packet);
                }
                else if (packet.Opcode == 2006)
                {
                    HandleAlphaChunkDataPacket(packet);
                }
                else if (packet.Opcode == 2001)
                {
                    HandleAlphaPlayerUpdatePacket(packet);
                }
                else if (packet.Opcode == 2005)
                {
                    HandleAlphaMoveConfirmPacket(packet);
                }
                else if (packet.Opcode == 3002)
                {
                    HandleAlphaDamageEventPacket(packet);
                }
                else if (packet.Opcode == 3003)
                {
                    HandleAlphaTargetDeadPacket(packet);
                }
                else if (packet.Opcode == 3004)
                {
                    HandleAlphaCreatureRespawnPacket(packet);
                }
                else if (packet.Opcode == 3005)
                {
                    HandleAlphaLootResultPacket(packet);
                }
                else if (packet.Opcode == BinaryProtocol.SC_CAST_SKILL_RESULT)
                {
                    HandleAlphaCastSkillResultPacket(packet);
                }
                else if (packet.Opcode == BinaryProtocol.SC_CHOOSE_VOCATION_RESP)
                {
                    HandleAlphaChooseVocationResponsePacket(packet);
                }
                else if (packet.Opcode == BinaryProtocol.SC_DIALOGUE_OPEN)
                {
                    HandleAlphaDialogueOpenPacket(packet);
                }
                else if (packet.Opcode == BinaryProtocol.SC_DIALOGUE_CLOSE)
                {
                    HandleAlphaDialogueClosePacket(packet);
                }
                else if (packet.Opcode == BinaryProtocol.SC_ALPHA_CAPABILITIES)
                {
                    HandleAlphaCapabilitiesPacket(packet);
                }
                else
                {
                    _ignoredPacketCount++;
                    if (_ignoredPacketCount == 1 || _ignoredPacketCount % 10 == 0)
                    {
                        CallDeferred(nameof(SetAlphaSystemMessage), $"Alpha listener ignoring non-bootstrap packets. Ignored: {_ignoredPacketCount}");
                    }
                }
            }
        }
        catch (OperationCanceledException)
        {
            GD.Print("Alpha world bootstrap listener stopped.");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha world bootstrap listener stopped with error: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Alpha world bootstrap listener error: {ex.GetType().Name}");
        }
    }

    private void StopAlphaPacketLoop()
    {
        if (_packetLoopCts == null)
        {
            return;
        }

        if (!_packetLoopCts.IsCancellationRequested)
        {
            _packetLoopCts.Cancel();
        }
    }

    private void HandleAlphaCapabilitiesPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeAlphaCapabilities(packet.Payload);
            CallDeferred(nameof(ApplyAlphaCapabilities), data.IsDevGM, data.Role);
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha Capabilities decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"GM capabilities decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyAlphaCapabilities(bool isDevGM, string role)
    {
        _isAlphaDevGM = isDevGM;
        _alphaAccountRole = role;
        SetAlphaSystemMessage($"GM capabilities received: IsDevGM={isDevGM}, Role={role}");
        RefreshAlphaDevGmBadge();
    }

    private void HandleAlphaChooseVocationResponsePacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeChooseVocationResponse(packet.Payload);
            CallDeferred(
                nameof(ApplyAlphaChooseVocationResponseValues),
                data.Success,
                data.ErrorMessage,
                data.ClassName
            );
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha ChooseVocation response decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Class selection feedback decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyAlphaChooseVocationResponseValues(bool success, string errorMessage, string className)
    {
        var displayClass = FormatAlphaClassName(className);

        if (success)
        {
            _confirmedAlphaClassId = className ?? string.Empty;

            if (_isAlphaDialogueOpen)
            {
                CloseAlphaDialogueWindow("class selection confirmed");
            }

            SetAlphaSystemMessage($"Class selected: {displayClass}.");
            SetAlphaCombatMessage($"Mentor Arion confirmed your class: {displayClass}.");
            RefreshTopBarShellState();
            GD.Print($"Alpha class selection confirmed: class={className}");
            return;
        }

        var safeError = string.IsNullOrWhiteSpace(errorMessage)
            ? "class selection rejected"
            : errorMessage.Trim();

        SetAlphaSystemMessage($"Class selection failed: {safeError}");
        SetAlphaCombatMessage($"Mentor Arion rejected class selection: {safeError}");
        GD.Print($"Alpha class selection rejected: error={safeError}, class={className}");
    }

    private static string FormatAlphaClassName(string className)
    {
        return (className ?? string.Empty).Trim().ToLowerInvariant() switch
        {
            "knight" => "Cavaleiro",
            "mage" => "Mago",
            "archer" => "Arqueiro",
            "assassin" => "Assassino",
            "cleric" => "Clérigo",
            "novice" => "Novato",
            "" => "unknown",
            var other => other
        };
    }

    private void HandleAlphaDialogueOpenPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeDialogueOpen(packet.Payload);
            var choiceParts = new List<string>();

            foreach (var choice in data.Choices)
            {
                choiceParts.Add($"{choice.NextNodeId}: {choice.Text}");
            }

            var choicesText = choiceParts.Count == 0 ? "none" : string.Join(" | ", choiceParts);

            CallDeferred(
                nameof(ApplyAlphaDialogueOpenValues),
                data.NpcId,
                data.NodeId,
                data.NodeText,
                choicesText,
                data.Choices.Count
            );
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha DialogueOpen decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"DialogueOpen decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyAlphaDialogueOpenValues(string npcId, string nodeId, string nodeText, string choicesText, int choiceCount)
    {
        BindAlphaDialogueWindow(npcId, nodeId, nodeText, choicesText, choiceCount);
        SetAlphaSystemMessage($"Dialogue opened: npc={npcId}, node={nodeId}, choices={choiceCount}.");
        SetAlphaCombatMessage($"NPC says: {nodeText}");
        GD.Print($"Alpha dialogue open decoded: npc={npcId}, node={nodeId}, text={nodeText}, choices={choicesText}");
    }
    private void HandleAlphaInventorySyncPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeInventorySync(packet.Payload);
            CallDeferred(
                nameof(ApplyInventorySyncValues),
                data.Level,
                data.Health,
                data.MaxHealth,
                data.Mana,
                data.MaxMana,
                data.Items.Count,
                data.HasAlphaProgression,
                data.Gold,
                data.Experience
            );
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha InventorySync decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"InventorySync decode failed: {ex.GetType().Name}");
        }
    }

    private void HandleAlphaChunkDataPacket(Packet packet)
    {
        try
        {
            var chunkData = BinaryProtocol.DecodeChunkData(packet.Payload);

            lock (_pendingChunkDataLock)
            {
                _pendingChunkData.Enqueue(chunkData);
            }

            CallDeferred(nameof(ApplyPendingChunkData));
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha ChunkData decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"World chunk sync decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyPendingChunkData()
    {
        var appliedCount = 0;

        while (true)
        {
            ChunkData? chunkData;

            lock (_pendingChunkDataLock)
            {
                if (_pendingChunkData.Count == 0)
                {
                    break;
                }

                chunkData = _pendingChunkData.Dequeue();
            }

            _chunkStore.AddChunk(chunkData.ChunkX, chunkData.ChunkY, chunkData.Tiles);
            appliedCount++;
        }

        if (appliedCount == 0)
        {
            return;
        }

        _hasWorldChunks = true;
        _syncedChunkCount = _chunkStore.Chunks.Count;

        RefreshWorldShellState();
        RequestAlphaWorldViewRedraw();
        SetAlphaSystemMessage($"World chunk sync received. Chunks synced: {_syncedChunkCount}");

        GD.Print($"Alpha world chunks applied: applied={appliedCount}, total={_syncedChunkCount}");
    }

    private void RequestAlphaWorldViewRedraw()
    {
        _worldView?.QueueRedraw();
    }

    private void HandleAlphaPlayerUpdatePacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodePlayerUpdate(packet.Payload);
            if (data == null)
            {
                CallDeferred(nameof(SetAlphaSystemMessage), "Player position sync decode returned empty data.");
                return;
            }

            if (!string.IsNullOrEmpty(_selectedCharacterNameForWorldEntry) && data.PlayerID != _selectedCharacterNameForWorldEntry)
            {
                return;
            }

            CallDeferred(
                nameof(ApplyLocalPlayerPositionValues),
                data.X,
                data.Y,
                data.Z,
                "Local player position synced."
            );
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha PlayerUpdate decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Player position sync decode failed: {ex.GetType().Name}");
        }
    }

    private void HandleAlphaMoveConfirmPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeMoveConfirm(packet.Payload);
            if (data == null)
            {
                CallDeferred(nameof(SetAlphaSystemMessage), "Authoritative position correction decode returned empty data.");
                return;
            }

            CallDeferred(
                nameof(ApplyAlphaMoveConfirmValues),
                data.X,
                data.Y,
                data.Z,
                data.Success
            );
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha MoveConfirm decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Authoritative position correction decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyLocalPlayerPositionValues(double x, double y, int z, string feedbackMessage)
    {
        _hasLocalPlayerPosition = true;
        _currentPlayerTilePosition = new Vector2I((int)Math.Round(x), (int)Math.Round(y));
        _currentPlayerTileZ = z;

        if (_worldView != null)
        {
            _worldView.PlayerTilePosition = _currentPlayerTilePosition;
            RefreshAlphaMentorArionDebugPosition();
            RefreshAlphaDialogueProximityState();
            SyncAlphaOrcEliteNearbyVisualMarker();
        }

        RefreshWorldShellState();
        RequestAlphaWorldViewRedraw();
        SetAlphaSystemMessage(feedbackMessage);

        GD.Print($"Alpha local player marker synced: z={z}");
    }

    private async Task SendAlphaMoveAsync(int deltaX, int deltaY, string source)
    {
        if (_isAlphaMovePending)
        {
            SetAlphaSystemMessage("Cannot move: waiting for server confirmation.");
            return;
        }

        if (!_hasLocalPlayerPosition)
        {
            SetAlphaSystemMessage("Cannot move: player position pending sync.");
            return;
        }

        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            SetAlphaSystemMessage("Cannot move: client disconnected.");
            return;
        }

        if (_packetLoopCts == null || _packetLoopCts.IsCancellationRequested)
        {
            SetAlphaSystemMessage("Cannot move: listener inactive.");
            return;
        }

        var nowUtc = DateTime.UtcNow;
        var elapsedSinceLastMove = nowUtc - _lastAlphaMoveRequestSentUtc;
        if (elapsedSinceLastMove < MinimumAlphaMoveRequestInterval)
        {
            var waitMs = Math.Max(0, MinimumAlphaMoveRequestInterval.TotalMilliseconds - elapsedSinceLastMove.TotalMilliseconds);
            SetAlphaSystemMessage($"Cannot move: waiting {waitMs:F0}ms for movement cooldown.");
            return;
        }

        var targetX = _currentPlayerTilePosition.X + deltaX;
        var targetY = _currentPlayerTilePosition.Y + deltaY;

        if (IsAlphaTileBlockedByOrcElite(targetX, targetY))
        {
            SetAlphaSystemMessage("Movement blocked: Orc_Elite occupies that tile.");
            GD.Print("Alpha movement blocked by Orc_Elite client-side collision preview.");
            return;
        }

        var targetZ = (sbyte)_currentPlayerTileZ;
        var clientTimestamp = (ulong)DateTimeOffset.UtcNow.ToUnixTimeMilliseconds();

        _isAlphaMovePending = true;
        _alphaPendingMoveTarget = new Vector2I(targetX, targetY);
        _lastAlphaMoveRequestSentUtc = nowUtc;

        if (_worldView != null)
        {
            _worldView.TargetPosition = _alphaPendingMoveTarget;
            RequestAlphaWorldViewRedraw();
        }

        SetAlphaSystemMessage($"Move requested by {source}.");

        try
        {
            await GatewayClient.SendMoveRequestAsync(targetX, targetY, targetZ, 0, clientTimestamp, _packetLoopCts.Token);
            SetAlphaSystemMessage("Move request sent. Waiting for server confirmation.");
            GD.Print($"Alpha move request sent: target=({targetX}, {targetY}, {targetZ})");
        }
        catch (OperationCanceledException)
        {
            ClearAlphaMovePendingState();
            SetAlphaSystemMessage("Move request cancelled.");
            GD.Print("Alpha move request cancelled.");
        }
        catch (Exception ex)
        {
            ClearAlphaMovePendingState();
            SetAlphaSystemMessage($"Move request failed: {ex.GetType().Name}.");
            GD.PrintErr($"Alpha move request failed: {ex.Message}");
        }
    }

    private bool IsAlphaTileBlockedByOrcElite(int tileX, int tileY)
    {
        if (_alphaBattleTargetState == "Dead")
        {
            return false;
        }

        if (_worldView?.OrcElitePosition is not Vector2I orcElitePosition)
        {
            return false;
        }

        return orcElitePosition.X == tileX && orcElitePosition.Y == tileY;
    }

    private void ApplyAlphaMoveConfirmValues(double x, double y, int z, bool success)
    {
        ClearAlphaMovePendingState();
        ApplyLocalPlayerPositionValues(
            x,
            y,
            z,
            success ? "Move confirmed by server." : "Move corrected by server."
        );
    }

    private void ClearAlphaMovePendingState()
    {
        _isAlphaMovePending = false;
        _alphaPendingMoveTarget = null;

        if (_worldView != null)
        {
            _worldView.TargetPosition = null;
            RequestAlphaWorldViewRedraw();
        }
    }

    private void SyncAlphaOrcEliteNearbyVisualMarker()
    {
        if (_worldView == null || !_hasAlphaOrcEliteVisualPosition)
        {
            return;
        }

        _worldView.OrcElitePosition = _alphaOrcEliteVisualPosition;
    }

    private static string BuildAlphaCombatFailureFeedback(string reason)
    {
        if (string.IsNullOrWhiteSpace(reason))
        {
            return "Combat action failed: server rejected attack.";
        }

        var trimmedReason = reason.Trim();
        if (trimmedReason.Contains("fora de alcance", StringComparison.OrdinalIgnoreCase) || trimmedReason.Contains("out of range", StringComparison.OrdinalIgnoreCase))
        {
            return $"Combat action failed: target out of debug sword range. Server: {trimmedReason}";
        }

        return $"Combat action failed: {trimmedReason}";
    }

    private void ApplyAlphaCombatFailureFeedback(string message)
    {
        var hasRunningAutoAttack = _alphaAutoAttackCts != null && !_alphaAutoAttackCts.IsCancellationRequested;
        if (hasRunningAutoAttack)
        {
            StopAlphaAutoAttackLoop("server rejected attack");
            SetAlphaCombatMessage($"{message} Auto-attack stopped.");
            return;
        }

        SetAlphaCombatMessage(message);
    }

    private static bool IsTemporaryLoginOverlapCombatBlock(string reason)
    {
        return string.Equals(
            reason?.Trim(),
            LoginOverlapOffensiveActionBlockedReason,
            StringComparison.Ordinal
        );
    }

    private void ApplyAlphaTemporaryCombatBlockFeedback(string reason)
    {
        SetAlphaCombatMessage(reason);
    }

    private void HandleAlphaDamageEventPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeDamageEvent(packet.Payload);
            GD.Print($"Alpha damage event decoded: target={data.TargetID}, skill={data.SkillName}, success={data.Success}, hit={data.IsHit}, damage={data.Damage:F0}");

            if (!data.Success)
            {
                if (IsTemporaryLoginOverlapCombatBlock(data.SkillName))
                {
                    CallDeferred(
                        nameof(ApplyAlphaTemporaryCombatBlockFeedback),
                        data.SkillName
                    );
                    return;
                }

                var failureMessage =
                    BuildAlphaCombatFailureFeedback(data.SkillName);

                CallDeferred(
                    nameof(ApplyAlphaCombatFailureFeedback),
                    failureMessage
                );
                return;
            }

            var isOrcEliteTarget = data.TargetID == "Orc_Elite" || data.TargetID == _alphaOrcEliteRuntimeEntityId;
            var isAlphaConfirmedSpell = isOrcEliteTarget && IsAlphaConfirmedSpellSkillName(data.SkillName);

            if (isAlphaConfirmedSpell)
            {
                CallDeferred(nameof(ApplyAlphaConfirmedSpellVisualFeedback), data.SkillName);
            }

            if (!data.IsHit)
            {
                CallDeferred(nameof(SetAlphaCombatMessage), "Combat event: attack missed.");
                CallDeferred(nameof(ApplyAlphaOrcEliteMissFloatingText));
                return;
            }

            var critText = data.IsCrit ? " Critical." : string.Empty;
            if (isOrcEliteTarget)
            {
                CallDeferred(nameof(ApplyAlphaOrcEliteDamageHudFeedback), data.Damage, data.IsCrit, _alphaOrcEliteRuntimeEntityId);
            }

            CallDeferred(nameof(SetAlphaCombatMessage), $"Combat event: {data.Damage:F0} damage.{critText}");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha DamageEvent decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaCombatMessage), $"Combat feedback decode failed: {ex.GetType().Name}");
        }
    }

    private void HandleAlphaTargetDeadPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeTargetDeadEvent(packet.Payload);

            var isOrcEliteTarget = data.TargetID == "Orc_Elite" || data.TargetID == _alphaOrcEliteRuntimeEntityId;
            if (!isOrcEliteTarget)
            {
                return;
            }

            CallDeferred(nameof(ApplyAlphaSafeTargetIdentity), data.RuntimeEntityID, true);
            CallDeferred(nameof(ApplyAlphaBattleTargetState), "Dead", "Orc_Elite defeated.");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha TargetDead decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Target state sync decode failed: {ex.GetType().Name}");
        }
    }

    private void HandleAlphaCastSkillResultPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeCastSkillResult(packet.Payload);
            if (data.Success)
            {
                var message = $"Cast confirmed: {data.Reason}.";
                CallDeferred(nameof(SetAlphaCombatMessage), message);
                GD.Print($"Alpha cast success: skill={data.SkillId}, reason='{data.Reason}', cooldown={data.CooldownRemainingMs}ms");
            }
            else
            {
                var message = $"Cast failed: {data.Reason}.";
                CallDeferred(nameof(SetAlphaCombatMessage), message);
                GD.Print($"Alpha cast failed: skill={data.SkillId}, reason='{data.Reason}', cooldown={data.CooldownRemainingMs}ms");
            }
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha CastSkillResult decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaCombatMessage), $"Cast feedback decode failed: {ex.GetType().Name}");
        }
    }

    private void HandleAlphaLootResultPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeLootResultEvent(packet.Payload);
            var message = BuildAlphaLootResultFeedback(data);
            CallDeferred(nameof(SetAlphaCombatMessage), message);
            GD.Print($"Alpha LootResult received: table={data.TableID}, item={data.ItemID}, quantity={data.Quantity}, dropped={data.Dropped}, granted={data.Granted}, reason={data.Reason}");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha LootResult decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaCombatMessage), $"Loot feedback decode failed: {ex.GetType().Name}");
        }
    }

    private static string BuildAlphaLootResultFeedback(LootResultEventData data)
    {
        var itemText = string.IsNullOrWhiteSpace(data.ItemID) ? "unknown item" : data.ItemID;
        var quantityText = data.Quantity > 1 ? $" x{data.Quantity}" : string.Empty;

        if (!data.Dropped)
        {
            return $"Loot result: no item dropped from {data.TableID}.";
        }

        if (data.Granted)
        {
            return $"Loot dropped: {itemText}{quantityText}. Loot granted to backpack.";
        }

        if (data.Reason == "inventory_full")
        {
            return $"Loot dropped: {itemText}{quantityText}. Loot blocked: inventory full.";
        }

        var reason = string.IsNullOrWhiteSpace(data.Reason) ? "unknown reason" : data.Reason;
        return $"Loot dropped: {itemText}{quantityText}. Loot blocked: {reason}.";
    }
    private void HandleAlphaCreatureRespawnPacket(Packet packet)
    {
        try
        {
            var data = BinaryProtocol.DecodeCreatureRespawnEvent(packet.Payload);

            var isOrcEliteTarget = data.TargetID == "Orc_Elite";
            if (!isOrcEliteTarget)
            {
                return;
            }

            CallDeferred(nameof(ApplyAlphaSafeTargetIdentity), data.RuntimeEntityID, false);
            CallDeferred(nameof(ApplyAlphaOrcEliteBackendPosition), data.HasPosition, data.X, data.Y, data.Z);
            CallDeferred(nameof(ApplyAlphaBattleTargetState), "Alive", "Orc_Elite respawned.");
        }
        catch (Exception ex)
        {
            GD.PrintErr($"Alpha CreatureRespawn decode failed: {ex.Message}");
            CallDeferred(nameof(SetAlphaSystemMessage), $"Target respawn sync decode failed: {ex.GetType().Name}");
        }
    }

    private void ApplyAlphaOrcEliteBackendPosition(bool hasPosition, double x, double y, int z)
    {
        if (!hasPosition)
        {
            return;
        }

        _alphaOrcEliteVisualPosition = new Vector2I((int)Math.Round(x), (int)Math.Round(y));
        _hasAlphaOrcEliteVisualPosition = true;

        if (_worldView != null)
        {
            _worldView.OrcElitePosition = _alphaOrcEliteVisualPosition;
        }

        SetAlphaSystemMessage($"Orc_Elite backend position synced: {_alphaOrcEliteVisualPosition.X},{_alphaOrcEliteVisualPosition.Y},z={z}.");
        RequestAlphaWorldViewRedraw();
    }
    private void ApplyAlphaSafeTargetIdentity(string runtimeEntityId, bool isDead)
    {
        if (string.IsNullOrWhiteSpace(runtimeEntityId))
        {
            return;
        }

        var oldId = _alphaOrcEliteRuntimeEntityId;
        _alphaOrcEliteRuntimeEntityId = runtimeEntityId.Trim();
        SetAlphaCombatMessage("Target identity synced.");

        if (isDead)
        {
            _worldView?.MarkAlphaOrcEliteDead(_alphaOrcEliteRuntimeEntityId);
        }
        else if (oldId != _alphaOrcEliteRuntimeEntityId)
        {
            _worldView?.ResetAlphaOrcEliteHealthForRespawn(_alphaOrcEliteRuntimeEntityId);
        }
        GD.Print("Alpha safe target identity synced for Orc_Elite.");
    }

    private bool HasAlphaSafeTargetIdentity()
    {
        return !string.IsNullOrWhiteSpace(_alphaOrcEliteRuntimeEntityId);
    }

    private void ApplyAlphaBattleTargetState(string state, string feedbackMessage)
    {
        _alphaBattleTargetState = state;

        if (state == "Dead")
        {
            StopAlphaAutoAttackLoop("target dead");
            _isAlphaOrcEliteSelected = false;
        }

        if (_worldView != null)
        {
            _worldView.IsOrcEliteDead = state == "Dead";
            _worldView.IsOrcEliteSelected = _isAlphaOrcEliteSelected && state == "Alive";
            SyncAlphaOrcEliteNearbyVisualMarker();
        }

        RefreshBattleTargetState();
        RequestAlphaWorldViewRedraw();
        SetAlphaSystemMessage(feedbackMessage);

        if (state == "Dead")
        {
            _pendingCombatRewardConfirmation = true;
            SetAlphaCombatMessage("Target defeated. Waiting for reward sync.");
        }
        else if (state == "Alive")
        {
            SetAlphaCombatMessage("Target respawned.");
        }

        GD.Print($"Alpha Battle target state updated: Orc_Elite={state}");
    }

    private void ApplyInventorySyncData(InventorySyncData data)
    {
        ApplyInventorySyncValues(
            data.Level,
            data.Health,
            data.MaxHealth,
            data.Mana,
            data.MaxMana,
            data.Items.Count,
            data.HasAlphaProgression,
            data.Gold,
            data.Experience
        );
    }

    private void ApplyInventorySyncValues(uint level, double health, double maxHealth, double mana, double maxMana, int itemCount, bool hasAlphaProgression, ulong gold, ulong experience)
    {
        var hadPreviousInventorySync = _hasInventorySync; var previousLevel = _syncedLevel; var previousItemCount = _syncedItemCount; var hadPreviousAlphaProgressionSync = _hasAlphaProgressionSync; var previousGold = _syncedGold; var previousExperience = _syncedExperience;
        _hasInventorySync = true;
        _syncedLevel = level;
        _syncedHealth = health;
        _syncedMaxHealth = maxHealth;
        _syncedMana = mana;
        _syncedMaxMana = maxMana;
        _syncedItemCount = itemCount;
        _hasAlphaProgressionSync = hasAlphaProgression;
        if (hasAlphaProgression)
        {
            _syncedGold = gold;
            _syncedExperience = experience;
        }

        RefreshTopBarShellState();
        SyncAlphaPlayerVitalsHud();
        RefreshBackpackShellState();
        SetAlphaSystemMessage($"InventorySync 4001 received. Items: {_syncedItemCount}");

        if (_pendingCombatRewardConfirmation) { _pendingCombatRewardConfirmation = false; var rewardMessage = BuildAlphaRewardSyncFeedback( hadPreviousInventorySync, previousLevel, previousItemCount, hadPreviousAlphaProgressionSync, previousGold, previousExperience, level, itemCount, hasAlphaProgression, gold, experience ); SetAlphaCombatMessage(rewardMessage); }

        GD.Print($"Alpha inventory sync applied: level={_syncedLevel}, hp={_syncedHealth:F2}/{_syncedMaxHealth:F2}, mana={_syncedMana:F2}/{_syncedMaxMana:F2}, items={_syncedItemCount}, hasProgression={_hasAlphaProgressionSync}, gold={_syncedGold}, xp={_syncedExperience}");
    }

    private static string BuildAlphaRewardSyncFeedback(
        bool hadPreviousInventorySync,
        uint previousLevel,
        int previousItemCount,
        bool hadPreviousAlphaProgressionSync,
        ulong previousGold,
        ulong previousExperience,
        uint level,
        int itemCount,
        bool hasAlphaProgression,
        ulong gold,
        ulong experience)
    {
        var parts = new List<string> { "Reward sync confirmed." };

        if (hasAlphaProgression && hadPreviousAlphaProgressionSync)
        {
            parts.Add(gold >= previousGold ? $"Gold +{gold - previousGold}" : $"Gold synced: {gold}");
            parts.Add(experience >= previousExperience ? $"XP +{experience - previousExperience}" : $"XP synced: {experience}");
        }
        else if (hasAlphaProgression)
        {
            parts.Add($"Gold: {gold}");
            parts.Add($"XP: {experience}");
        }
        else
        {
            parts.Add("Gold/XP pending sync");
        }

        if (hadPreviousInventorySync)
        {
            if (level > previousLevel)
            {
                parts.Add($"Level {previousLevel}->{level}");
            }

            if (itemCount > previousItemCount)
            {
                parts.Add($"Item slots +{itemCount - previousItemCount}");
            }
            else if (itemCount == previousItemCount)
            {
                parts.Add("No new item slot synced");
            }
        }

        return string.Join(" | ", parts);
    }

    private void SetAlphaSystemMessage(string message)
    {
        if (!string.IsNullOrWhiteSpace(message))
        {
            _systemFeedbackMessages.Enqueue(message.Trim());

            while (_systemFeedbackMessages.Count > MaxSystemFeedbackMessages)
            {
                _systemFeedbackMessages.Dequeue();
            }
        }

        if (_systemFeedbackLabel != null)
        {
            var lines = new List<string> { "System" };

            foreach (var feedbackMessage in _systemFeedbackMessages)
            {
                lines.Add($"- {feedbackMessage}");
            }

            _systemFeedbackLabel.Text = string.Join("\n", lines);
        }

        _editableSystemLogPanel?.BindMessages("System", _systemFeedbackMessages);

        GD.Print($"Alpha System: {message}");
    }
    private void SetAlphaCombatMessage(string message)
    {
        if (!string.IsNullOrWhiteSpace(message))
        {
            _combatFeedbackMessages.Enqueue(message.Trim());

            while (_combatFeedbackMessages.Count > MaxCombatFeedbackMessages)
            {
                _combatFeedbackMessages.Dequeue();
            }
        }

        RefreshCombatFeedbackState();

        GD.Print($"Alpha Combat: {message}");
    }

    private void OnAlphaWorldViewGuiInput(InputEvent inputEvent)
    {
        if (inputEvent is not InputEventMouseButton mouseButton)
        {
            return;
        }

        if (!mouseButton.Pressed)
        {
            return;
        }

        if (mouseButton.ButtonIndex == MouseButton.Left)
        {
            OnAlphaLeftClickMoveRequested(mouseButton);
            return;
        }

        if (mouseButton.ButtonIndex == MouseButton.Right)
        {
            if (TryHandleAlphaMentorArionRightClick(mouseButton))
            {
                return;
            }

            if (TryHandleAlphaOrcEliteRightClick(mouseButton))
            {
                return;
            }

            SetAlphaCombatMessage("No valid target."); // B3-F: Feedback for empty right-click
        }
    }
    private void OnAlphaLeftClickMoveRequested(InputEventMouseButton mouseButton)
    {
        if (_worldView == null || !_worldView.TryGetFocusedTileAtLocalPosition(mouseButton.Position, out var clickedTile))
        {
            SetAlphaSystemMessage("Cannot move: clicked tile could not be resolved.");
            return;
        }

        if (!_hasLocalPlayerPosition)
        {
            SetAlphaSystemMessage("Cannot move: player position pending sync.");
            return;
        }

        var deltaX = Math.Clamp(clickedTile.X - _currentPlayerTilePosition.X, -1, 1);
        var deltaY = Math.Clamp(clickedTile.Y - _currentPlayerTilePosition.Y, -1, 1);

        if (deltaX == 0 && deltaY == 0)
        {
            SetAlphaSystemMessage("Cannot move: clicked current tile.");
            return;
        }

        var targetTile = new Vector2I(_currentPlayerTilePosition.X + deltaX, _currentPlayerTilePosition.Y + deltaY);
        if (targetTile.X == AlphaMentorArionTilePosition.X && targetTile.Y == AlphaMentorArionTilePosition.Y)
        {
            SetAlphaSystemMessage($"Cannot move: Mentor Arion blocks tile {targetTile.X},{targetTile.Y}.");
            SetAlphaCombatMessage("Movement blocked: NPCs are solid.");
            GD.Print($"Alpha movement blocked by Mentor Arion: targetTile={targetTile}, mentor={AlphaMentorArionTilePosition}");
            return;
        }

        GD.Print($"Alpha left-click movement requested: clickedTile={clickedTile}, player={_currentPlayerTilePosition}, step=({deltaX},{deltaY})");
        _ = SendAlphaMoveAsync(deltaX, deltaY, "left-click");
    }

    private bool TryHandleAlphaOrcEliteRightClick(InputEventMouseButton mouseButton)
    {
        if (_worldView == null || !_worldView.TryGetFocusedTileAtLocalPosition(mouseButton.Position, out var clickedTile))
        {
            return false;
        }

        if (!_worldView.OrcElitePosition.HasValue)
        {
            return false;
        }

        var orcTile = _worldView.OrcElitePosition.Value;
        if (clickedTile.X != orcTile.X || clickedTile.Y != orcTile.Y)
        {
            return false;
        }

        SelectAlphaOrcEliteTargetAndStartAutoAttack();
        return true;
    }
    private bool TryHandleAlphaMentorArionRightClick(InputEventMouseButton mouseButton)
    {
        if (_worldView == null || !_worldView.TryGetFocusedTileAtLocalPosition(mouseButton.Position, out var clickedTile))
        {
            return false;
        }

        if (clickedTile.X != AlphaMentorArionTilePosition.X || clickedTile.Y != AlphaMentorArionTilePosition.Y)
        {
            return false;
        }

        SetAlphaSystemMessage($"Interacting with Mentor Arion..."); // B3-F: Interaction feedback
        SetAlphaCombatMessage($"Interacting with Mentor Arion at {clickedTile.X},{clickedTile.Y}.");
        GD.Print($"Alpha Mentor Arion interaction requested: npc={AlphaMentorArionNpcId}, tile={clickedTile}");
        _ = SendAlphaNpcInteractRequestAsync(AlphaMentorArionNpcId);
        return true;
    }

    private async Task SendAlphaNpcInteractRequestAsync(string npcId)
    {
        if (string.IsNullOrWhiteSpace(npcId))
        {
            SetAlphaSystemMessage("Cannot interact: NPC id is empty.");
            return;
        }

        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            SetAlphaSystemMessage("Cannot interact: client disconnected.");
            return;
        }

        if (_packetLoopCts == null || _packetLoopCts.IsCancellationRequested)
        {
            SetAlphaSystemMessage("Cannot interact: listener inactive.");
            return;
        }

        try
        {
            SetAlphaSystemMessage($"Interacting with NPC: {npcId}.");
            await GatewayClient.SendNpcInteractRequestAsync(npcId, _packetLoopCts.Token);
            SetAlphaSystemMessage($"NPC interaction request sent for {npcId} at tile {AlphaMentorArionTilePosition.X},{AlphaMentorArionTilePosition.Y}. Waiting for dialogue open.");
            SetAlphaCombatMessage($"NPC interaction sent: Mentor Arion at {AlphaMentorArionTilePosition.X},{AlphaMentorArionTilePosition.Y}.");
            GD.Print($"Alpha NPC interact request sent: npc={npcId}");
        }
        catch (OperationCanceledException)
        {
            SetAlphaSystemMessage("NPC interaction request cancelled.");
            GD.Print("Alpha NPC interact request cancelled.");
        }
        catch (Exception ex)
        {
            SetAlphaSystemMessage($"NPC interaction failed: {ex.GetType().Name}.");
            GD.PrintErr($"Alpha NPC interact request failed: {ex.Message}");
        }
    }
    private void SelectAlphaOrcEliteTargetAndStartAutoAttack()
    {
        if (_alphaBattleTargetState == "Dead")
        {
            _isAlphaOrcEliteSelected = false;

            if (_worldView != null)
            {
                _worldView.IsOrcEliteSelected = false;
                RequestAlphaWorldViewRedraw();
            }

            SetAlphaCombatMessage("Cannot select target: Orc_Elite is dead.");
            return;
        }

        if (_alphaBattleTargetState != "Alive")
        {
            SetAlphaCombatMessage("Cannot select target: Orc_Elite is not ready.");
            return;
        }

        _isAlphaOrcEliteSelected = true;

        if (_worldView != null)
        {
            _worldView.IsOrcEliteSelected = true;
            SyncAlphaOrcEliteNearbyVisualMarker();
            RequestAlphaWorldViewRedraw();
        }

        RefreshBattleTargetState();
        RefreshAlphaSpellbookShellState(); // B3-F: Interaction feedback
        SetAlphaCombatMessage("Target selected: Orc_Elite. Auto-attack started.");
        SetAlphaSystemMessage("Alpha target selected.");
        StartAlphaAutoAttackLoop();
    }

    private void StartAlphaAutoAttackLoop()
    {
        if (_alphaBattleTargetState != "Alive")
        {
            SetAlphaCombatMessage("Cannot auto-attack: target not ready.");
            return;
        }

        if (!_isAlphaOrcEliteSelected)
        {
            SetAlphaCombatMessage("Cannot auto-attack: target not selected.");
            return;
        }

        if (!HasAlphaSafeTargetIdentity())
        {
            SetAlphaCombatMessage("Cannot auto-attack: target identity pending.");
            return;
        }

        if (_alphaAutoAttackCts != null && !_alphaAutoAttackCts.IsCancellationRequested)
        {
            SetAlphaCombatMessage("Auto-attack already running.");
            return;
        }

        _alphaAutoAttackCts?.Dispose();
        _alphaAutoAttackCts = new CancellationTokenSource();

        SetAlphaCombatMessage("Auto-attack started.");
        GD.Print("Alpha auto-attack loop started for Orc_Elite.");
        _ = RunAlphaAutoAttackLoopAsync(_alphaAutoAttackCts.Token);
    }

    private void StopAlphaAutoAttackLoop(string reason)
    {
        if (_alphaAutoAttackCts == null)
        {
            return;
        }

        if (!_alphaAutoAttackCts.IsCancellationRequested)
        {
            _alphaAutoAttackCts.Cancel();
        }

        _alphaAutoAttackCts.Dispose();
        _alphaAutoAttackCts = null;
        _isAlphaAttackRequestInFlight = false;

        if (!string.IsNullOrWhiteSpace(reason))
        {
            GD.Print($"Alpha auto-attack loop stopped: {reason}");
        }
    }

    private async Task RunAlphaAutoAttackLoopAsync(CancellationToken cancellationToken)
    {
        try
        {
            while (!cancellationToken.IsCancellationRequested)
            {
                if (!_isAlphaOrcEliteSelected || _alphaBattleTargetState != "Alive")
                {
                    StopAlphaAutoAttackLoop("target no longer attackable");
                    return;
                }

                await SendAlphaAttackOnceAsync("auto-attack", cancellationToken);
                await Task.Delay(AlphaAutoAttackInterval, cancellationToken);
            }
        }
        catch (OperationCanceledException)
        {
            GD.Print("Alpha auto-attack loop cancelled.");
        }
        catch (Exception ex)
        {
            SetAlphaCombatMessage($"Auto-attack stopped: {ex.GetType().Name}.");
            GD.PrintErr($"Alpha auto-attack loop failed: {ex.Message}");
            StopAlphaAutoAttackLoop("loop error");
        }
    }

    private bool IsAlphaDebugSwordPreviewRangeReady(out double distance)
    {
        distance = 0;

        if (!_hasLocalPlayerPosition)
        {
            return true;
        }

        var deltaX = _currentPlayerTilePosition.X - _alphaOrcEliteVisualPosition.X;
        var deltaY = _currentPlayerTilePosition.Y - _alphaOrcEliteVisualPosition.Y;
        distance = Math.Max(Math.Abs(deltaX), Math.Abs(deltaY));

        return distance <= AlphaDebugSwordPreviewRangeTiles;
    }

    private async Task SendAlphaAttackOnceAsync(string source, CancellationToken cancellationToken = default)
    {
        if (GatewayClient == null || !GatewayClient.IsConnected)
        {
            SetAlphaCombatMessage("Cannot attack: client disconnected.");
            StopAlphaAutoAttackLoop("client disconnected");
            return;
        }

        if (_alphaBattleTargetState == "Dead")
        {
            SetAlphaCombatMessage("Cannot attack: target is dead.");
            StopAlphaAutoAttackLoop("target dead");
            return;
        }

        if (_alphaBattleTargetState != "Alive")
        {
            SetAlphaCombatMessage("Cannot attack: target not ready.");
            return;
        }

        if (!HasAlphaSafeTargetIdentity())
        {
            SetAlphaCombatMessage("Cannot attack: target identity pending.");
            GD.Print("Alpha attack blocked: safe target identity is pending.");
            return;
        }

        if (_packetLoopCts == null || _packetLoopCts.IsCancellationRequested)
        {
            SetAlphaCombatMessage("Cannot attack: listener inactive.");
            StopAlphaAutoAttackLoop("listener inactive");
            return;
        }

        if (!IsAlphaDebugSwordPreviewRangeReady(out var previewDistance)) { var rangeMessage = $"Cannot attack: move into debug sword range. Distance: {previewDistance:F2}, Range: {AlphaDebugSwordPreviewRangeTiles:F2}."; SetAlphaCombatMessage(rangeMessage); if (source == "auto-attack") { StopAlphaAutoAttackLoop("target outside debug sword preview range"); } return; }
        if (_isAlphaAttackRequestInFlight)
        {
            if (source == "right-click")
            {
                SetAlphaCombatMessage("Attack already in flight.");
            }

            return;
        }

        var effectiveToken = cancellationToken.CanBeCanceled ? cancellationToken : _packetLoopCts.Token;

        try
        {
            _isAlphaAttackRequestInFlight = true;
            if (source != "auto-attack") { SetAlphaCombatMessage("Sending attack request."); }
            await GatewayClient.SendAttackRequestAsync(_alphaOrcEliteRuntimeEntityId, AlphaRealAttackWeaponType, effectiveToken);
            if (source != "auto-attack") { SetAlphaCombatMessage("Attack request sent."); }
            GD.Print($"Alpha {source} attack request sent with safe target identity.");
        }
        catch (OperationCanceledException)
        {
            SetAlphaCombatMessage("Attack cancelled.");
            GD.Print($"Alpha {source} attack request cancelled.");
        }
        catch (Exception ex)
        {
            SetAlphaCombatMessage($"Attack send failed: {ex.GetType().Name}.");
            GD.PrintErr($"Alpha {source} attack request failed: {ex.Message}");
        }
        finally
        {
            _isAlphaAttackRequestInFlight = false;
        }
    }

    private void OnBackButtonPressed()
    {
        SetAlphaSystemMessage("Back requested. Alpha listener cancellation requested.");
        StopAlphaPacketLoop();
        SceneFlow.ToDebugAuth(this);
    }
}
