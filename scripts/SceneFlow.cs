using System;
using Godot;

namespace LightAndShadow.Client;

/// <summary>
/// Utility class for debug scene transitions.
/// It does not store session state, does not persist tokens, and is not an Autoload.
/// </summary>
public static class SceneFlow
{
    public const string DebugAuthScenePath = "res://scenes/DebugAuthScene.tscn";
    public const string DebugWorldEntryScenePath = "res://scenes/DebugWorldEntryScene.tscn";

    public static void ToWorldEntry(Node caller, AuthSession session)
    {
        if (session == null)
        {
            throw new ArgumentNullException(nameof(session));
        }

        ChangeScene(
            caller,
            DebugWorldEntryScenePath,
            instance =>
            {
                if (instance is DebugWorldEntryController worldEntryController)
                {
                    worldEntryController.Session = session;
                    return;
                }

                throw new InvalidOperationException($"Scene '{DebugWorldEntryScenePath}' does not use DebugWorldEntryController.");
            });
    }

    public static void ToDebugAuth(Node caller)
    {
        ChangeScene(caller, DebugAuthScenePath, null);
    }

    private static void ChangeScene(Node caller, string scenePath, Action<Node>? configure)
    {
        if (caller == null)
        {
            throw new ArgumentNullException(nameof(caller));
        }

        var tree = caller.GetTree();
        var packedScene = ResourceLoader.Load<PackedScene>(scenePath);
        if (packedScene == null)
        {
            throw new InvalidOperationException($"Failed to load scene '{scenePath}'.");
        }

        var nextScene = packedScene.Instantiate();
        configure?.Invoke(nextScene);

        var previousScene = tree.CurrentScene;
        tree.Root.AddChild(nextScene);
        tree.CurrentScene = nextScene;

        if (previousScene != null && previousScene != nextScene)
        {
            previousScene.QueueFree();
        }
        else if (caller != nextScene)
        {
            caller.QueueFree();
        }
    }
}
