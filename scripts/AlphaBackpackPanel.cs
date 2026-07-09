using Godot;

public partial class AlphaBackpackPanel : PanelContainer
{
    private Label? _summaryLabel;

    public override void _Ready()
    {
        _summaryLabel = GetNodeOrNull<Label>("Content/SummaryLabel");
    }

    public void BindBackpackSummary(int itemCount)
    {
        if (_summaryLabel != null)
        {
            _summaryLabel.Text = $"Backpack: {itemCount} item(s)";
        }
    }
}