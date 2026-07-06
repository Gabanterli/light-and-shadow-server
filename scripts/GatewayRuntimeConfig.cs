using Godot;
using System;
using System.IO;
using System.Text.Json;

/// <summary>
/// Carrega a configuração do gateway (host e porta) em tempo de execução a partir de um arquivo JSON.
/// Se o arquivo não existir ou for inválido, utiliza valores padrão.
/// </summary>
public class GatewayRuntimeConfig
{
    private const string ConfigFileName = "gateway-config.json";
    private const string DefaultHost = "127.0.0.1";
    private const int DefaultPort = 8080;

    public string Host { get; private set; } = DefaultHost;
    public int Port { get; private set; } = DefaultPort;
    public string SourceDescription { get; private set; } = "default";
    public bool IsDefault { get; private set; } = true;

    // Classe interna para deserialização do JSON.
    private class ConfigData
    {
        public string? Host { get; set; }
        public int Port { get; set; }
    }

    public GatewayRuntimeConfig()
    {
        LoadConfig();
    }

    private void LoadConfig()
    {
        string configPath = string.Empty;

        // Determina o caminho do arquivo de configuração dependendo se está no editor ou em um build exportado.
        if (OS.HasFeature("editor"))
        {
            // No editor, procura na raiz do projeto (res://).
            configPath = ProjectSettings.GlobalizePath($"res://{ConfigFileName}");
        }
        else
        {
            // Em um build exportado, procura ao lado do executável.
            var exePath = OS.GetExecutablePath();
            var exeDir = Path.GetDirectoryName(exePath);
            if (!string.IsNullOrEmpty(exeDir))
            {
                configPath = Path.Combine(exeDir, ConfigFileName);
            }
        }

        if (string.IsNullOrEmpty(configPath) || !File.Exists(configPath))
        {
            // Se o arquivo não for encontrado, mantém os valores padrão.
            SourceDescription = "default (config file not found)";
            return;
        }

        try
        {
            string jsonContent = File.ReadAllText(configPath);
            var configData = JsonSerializer.Deserialize<ConfigData>(jsonContent);

            // Valida os dados carregados.
            if (configData != null &&
                !string.IsNullOrWhiteSpace(configData.Host) &&
                configData.Port > 0 && configData.Port <= 65535)
            {
                Host = configData.Host;
                Port = configData.Port;
                SourceDescription = ConfigFileName;
                IsDefault = false;
            }
            else
            {
                SourceDescription = $"default (invalid data in {ConfigFileName})";
            }
        }
        catch (Exception ex)
        {
            // Em caso de erro de leitura ou parsing, mantém os valores padrão.
            SourceDescription = $"default (error reading config: {ex.GetType().Name})";
        }
    }

    public override string ToString()
    {
        return $"Host: {Host}, Port: {Port}, Source: {SourceDescription}";
    }
}