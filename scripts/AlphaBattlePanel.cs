using Godot;

public partial class AlphaBattlePanel : PanelContainer
{
    private Label? _targetLabel;
    private Label? _stateLabel;

    public override void _Ready()
    {
        _targetLabel = GetNodeOrNull<Label>("Content/TargetLabel");
        _stateLabel = GetNodeOrNull<Label>("Content/StateLabel");
    }

    public void BindTargetState(string targetName, string state, bool isSelected)
    {
        if (_targetLabel != null)
        {
            _targetLabel.Text = isSelected ? $"Target: {targetName} [selected]" : $"Target: {targetName}";
        }

        if (_stateLabel != null)
        {
            _stateLabel.Text = $"State: {state}";
        }
    }
}