using Godot;

public partial class AlphaWorldPanel : PanelContainer
{
    [Export]
    public NodePath WorldViewPath { get; set; } = new("WorldView");

    public DebugTileWorldView? WorldView { get; private set; }

    public override void _Ready()
    {
        WorldView = GetNodeOrNull<DebugTileWorldView>(WorldViewPath);
    }

    public void QueueWorldRedraw()
    {
        WorldView?.QueueRedraw();
    }
}