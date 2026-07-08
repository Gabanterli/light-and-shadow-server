using Godot;
using LightAndShadow.Client;
using System;

namespace LightAndShadow.Client;

/// <summary>
/// A static utility class to manage scene transitions and flow.
/// </summary>
public static class SceneFlow
{
    public const string DebugAuthScenePath = "res://scenes/DebugAuthScene.tscn";
    public const string DebugWorldEntryScenePath = "res://scenes/DebugWorldEntryScene.tscn";
    public const string AlphaWorldEntryScenePath = "res://scenes/AlphaWorldEntryScene.tscn";

    public static void ToWorldEntry(Node caller, AuthSession session, GatewayTcpClient client)
    {
        var packedScene = ResourceLoader.Load<PackedScene>(DebugWorldEntryScenePath);
        var instance = packedScene.Instantiate();

        if (instance is DebugWorldEntryController worldEntryController)
        {
            // Pass the current session object to the next scene's controller.
            worldEntryController.Session = session;
            worldEntryController.GatewayClient = client;
        }
        else
        {
            GD.PrintErr($"Error: Could not instantiate or find controller in {DebugWorldEntryScenePath}. Scene transition aborted.");
            return;
        }

        ChangeScene(caller, instance);
    }

    public static void ToAlphaWorldEntryShell(Node caller)
    {
        var packedScene = ResourceLoader.Load<PackedScene>(AlphaWorldEntryScenePath);
        var instance = packedScene.Instantiate();

        if (instance is not AlphaWorldEntryController)
        {
            GD.PrintErr($"Error: Could not instantiate or find controller in {AlphaWorldEntryScenePath}. Scene transition aborted.");
            return;
        }

        ChangeScene(caller, instance);
    }

    public static void ToDebugAuth(Node caller)
    {
        var tree = caller.GetTree();
        var previousScene = tree.CurrentScene;

        tree.ChangeSceneToFile(DebugAuthScenePath);

        if (previousScene != null)
        {
            previousScene.QueueFree();
        }
    }

    private static void ChangeScene(Node caller, Node nextScene)
    {
        var tree = caller.GetTree();
        var previousScene = tree.CurrentScene;

        tree.Root.AddChild(nextScene);
        tree.CurrentScene = nextScene;

        if (previousScene != null)
        {
            previousScene.QueueFree();
        }
        else
        {
            caller.QueueFree();
        }
    }
}
