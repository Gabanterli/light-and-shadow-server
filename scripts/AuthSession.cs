using System;

namespace LightAndShadow.Client;

public sealed class AuthSession
{
    public uint AccountId { get; private set; }
    public string SessionToken { get; private set; } = string.Empty;
    public string SelectedCharacterName { get; private set; } = string.Empty;

    public bool IsAuthenticated => !string.IsNullOrEmpty(SessionToken) && AccountId > 0;
    public bool IsCharacterSelected => !string.IsNullOrEmpty(SelectedCharacterName);

    public void SetLogin(uint accountId, string token)
    {
        if (string.IsNullOrWhiteSpace(token))
        {
            throw new ArgumentException("Session token cannot be null or whitespace.", nameof(token));
        }
        if (accountId == 0)
        {
            throw new ArgumentException("Account ID cannot be zero.", nameof(accountId));
        }

        AccountId = accountId;
        SessionToken = token;
        SelectedCharacterName = string.Empty; // Reset character on new login
    }

    public void SetSelectedCharacter(string characterName)
    {
        if (!IsAuthenticated)
        {
            throw new InvalidOperationException("Cannot select a character without an authenticated session.");
        }
        if (string.IsNullOrWhiteSpace(characterName))
        {
            throw new ArgumentException("Character name cannot be null or whitespace.", nameof(characterName));
        }

        SelectedCharacterName = characterName;
    }

    public void Clear()
    {
        AccountId = 0;
        SessionToken = string.Empty;
        SelectedCharacterName = string.Empty;
    }
}