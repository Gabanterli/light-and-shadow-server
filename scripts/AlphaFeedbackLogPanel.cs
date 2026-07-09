using Godot;
using System.Collections.Generic;
using System.Linq;

public partial class AlphaFeedbackLogPanel : PanelContainer
{
    private Label? _titleLabel;
    private Label? _messagesLabel;

    public override void _Ready()
    {
        _titleLabel = GetNodeOrNull<Label>("Content/TitleLabel");
        _messagesLabel = GetNodeOrNull<Label>("Content/MessagesLabel");
    }

    public void BindMessages(string title, IEnumerable<string> messages)
    {
        if (_titleLabel != null)
        {
            _titleLabel.Text = title;
        }

        if (_messagesLabel != null)
        {
            var lines = messages.Select(message => $"- {message}");
            _messagesLabel.Text = string.Join("\n", lines);
        }
    }
}