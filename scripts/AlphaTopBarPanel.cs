using Godot;

public partial class AlphaTopBarPanel : PanelContainer
{
    private Label? _nameLabel;
    private Label? _levelLabel;
    private Label? _healthLabel;
    private Label? _manaLabel;

    public override void _Ready()
    {
        _nameLabel = GetNodeOrNull<Label>("Content/NameLabel");
        _levelLabel = GetNodeOrNull<Label>("Content/LevelLabel");
        _healthLabel = GetNodeOrNull<Label>("Content/HealthLabel");
        _manaLabel = GetNodeOrNull<Label>("Content/ManaLabel");
    }

    public void BindPlayerStatus(string playerName, uint level, double health, double maxHealth, double mana, double maxMana)
    {
        if (_nameLabel != null)
        {
            _nameLabel.Text = $"Name: {playerName}";
        }

        if (_levelLabel != null)
        {
            _levelLabel.Text = $"Level: {level}";
        }

        if (_healthLabel != null)
        {
            _healthLabel.Text = $"HP: {health:F0}/{maxHealth:F0}";
        }

        if (_manaLabel != null)
        {
            _manaLabel.Text = $"Mana: {mana:F0}/{maxMana:F0}";
        }
    }
}