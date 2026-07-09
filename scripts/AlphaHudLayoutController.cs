using Godot;

[Tool]
public partial class AlphaHudLayoutController : Control
{
    [Export]
    public bool UseEditorPreviewState { get; set; } = true;

    private AlphaTopBarPanel? _topBar;
    private AlphaWorldPanel? _worldPanel;
    private AlphaBattlePanel? _battlePanel;
    private AlphaBackpackPanel? _backpackPanel;
    private AlphaFeedbackLogPanel? _combatLogPanel;
    private AlphaFeedbackLogPanel? _systemLogPanel;
    private bool _editorPreviewApplied;

    public override void _Ready()
    {
        BindComponentReferences();

        if (UseEditorPreviewState)
        {
            ApplyEditorPreviewState();
        }
    }

    public override void _Process(double delta)
    {
        if (!Engine.IsEditorHint() || !UseEditorPreviewState || _editorPreviewApplied)
        {
            return;
        }

        BindComponentReferences();
        ApplyEditorPreviewState();
        _editorPreviewApplied = true;
    }

    private void BindComponentReferences()
    {
        _topBar = GetNodeOrNull<AlphaTopBarPanel>("Root/TopBar");
        _worldPanel = GetNodeOrNull<AlphaWorldPanel>("Root/Main/WorldPanel");
        _battlePanel = GetNodeOrNull<AlphaBattlePanel>("Root/Main/SidePanel/BattlePanel");
        _backpackPanel = GetNodeOrNull<AlphaBackpackPanel>("Root/Main/SidePanel/BackpackPanel");
        _combatLogPanel = GetNodeOrNull<AlphaFeedbackLogPanel>("Root/Logs/CombatLogPanel");
        _systemLogPanel = GetNodeOrNull<AlphaFeedbackLogPanel>("Root/Logs/SystemLogPanel");
    }

    public void ApplyEditorPreviewState()
    {
        _topBar?.BindPlayerStatus("Gabriela", 1, 100, 100, 50, 50);
        _battlePanel?.BindTargetState("Orc_Elite", "Alive", true);
        _backpackPanel?.BindBackpackSummary(0);
        _combatLogPanel?.BindMessages("Combat", new[]
        {
            "Target selected: Orc_Elite.",
            "Ready for Alpha combat probe."
        });
        _systemLogPanel?.BindMessages("System", new[]
        {
            "Editable Alpha HUD preview.",
            "Runtime binding remains controller-authoritative."
        });

        SetLabelText("Root/TopBar/Content/NameLabel", "Name: Gabriela");
        SetLabelText("Root/TopBar/Content/LevelLabel", "Level: 1");
        SetLabelText("Root/TopBar/Content/HealthLabel", "HP: 100/100");
        SetLabelText("Root/TopBar/Content/ManaLabel", "Mana: 50/50");
        SetLabelText("Root/Main/SidePanel/BattlePanel/Content/TargetLabel", "Target: Orc_Elite [selected]");
        SetLabelText("Root/Main/SidePanel/BattlePanel/Content/StateLabel", "State: Alive");
        SetLabelText("Root/Main/SidePanel/BackpackPanel/Content/SummaryLabel", "Backpack: 0 item(s)");
        SetLabelText("Root/Logs/CombatLogPanel/Content/MessagesLabel", "- Target selected: Orc_Elite.\n- Ready for Alpha combat probe.");
        SetLabelText("Root/Logs/SystemLogPanel/Content/MessagesLabel", "- Editable Alpha HUD preview.\n- Runtime binding remains controller-authoritative.");

        _worldPanel?.QueueWorldRedraw();
    }

    private void SetLabelText(string nodePath, string text)
    {
        var label = GetNodeOrNull<Label>(nodePath);
        if (label != null)
        {
            label.Text = text;
        }
    }
}