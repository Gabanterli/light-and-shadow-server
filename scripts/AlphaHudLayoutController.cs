using Godot;

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

    public override void _Ready()
    {
        _topBar = GetNodeOrNull<AlphaTopBarPanel>("Root/TopBar");
        _worldPanel = GetNodeOrNull<AlphaWorldPanel>("Root/Main/WorldPanel");
        _battlePanel = GetNodeOrNull<AlphaBattlePanel>("Root/Main/SidePanel/BattlePanel");
        _backpackPanel = GetNodeOrNull<AlphaBackpackPanel>("Root/Main/SidePanel/BackpackPanel");
        _combatLogPanel = GetNodeOrNull<AlphaFeedbackLogPanel>("Root/Logs/CombatLogPanel");
        _systemLogPanel = GetNodeOrNull<AlphaFeedbackLogPanel>("Root/Logs/SystemLogPanel");

        if (UseEditorPreviewState)
        {
            ApplyEditorPreviewState();
        }
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

        _worldPanel?.QueueWorldRedraw();
    }
}