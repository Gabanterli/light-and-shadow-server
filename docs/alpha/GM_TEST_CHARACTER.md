# GM/Test Character - Alpha Environment Guide

This document outlines the server-authoritative system for GM (Game Master) and Test characters, designed exclusively for local development, QA, and the Alpha test environment.

## 1. What is a GM/Test Character?

A GM/Test character is an account with special privileges that can only be activated in a development environment. These privileges allow developers and testers to bypass certain game rules (like level gates) to accelerate testing of specific features.

**This system is production-safe by default.** All GM capabilities are completely disabled unless explicitly turned on via a server environment variable.

## 2. How to Activate GM/Test Mode

Activation requires two steps:

### Step 1: Enable the Dev GM Flag on the Backend

The entire system is controlled by the `LS_ENABLE_DEV_GM` environment variable. To enable it, set it when running the gateway server.

**Using Docker Compose:**

Add the following to the `environment` section of the `gateway-server` service in your local `docker-compose.yml` file:

```yaml
services:
  gateway-server:
    environment:
      - LS_ENABLE_DEV_GM=true
```

### Step 2: Mark an Account as GM

Connect to your local PostgreSQL database and update the `role` for a specific account.

```sql
-- Find your account ID if you don't know it
SELECT id, username FROM accounts;

-- Update the role for the desired user
UPDATE accounts SET role='gm' WHERE username='your_test_username';
-- Or 'admin'
UPDATE accounts SET role='admin' WHERE username='your_test_username';
```

## 3. How to Validate

1.  **GM Character (Dev Mode ON):**
    -   Log in with the account marked as `gm` or `admin`.
    -   Select a character that is **below level 10**.
    -   Interact with "Mentor Arion" and choose a class.
    -   **Expected Result:** The class selection should succeed. The client will show a "GM TEST" badge.

2.  **Normal Character (Dev Mode ON):**
    -   Log in with a normal account (role `player`).
    -   Select a character below level 10.
    -   Interact with "Mentor Arion".
    -   **Expected Result:** The class selection should be rejected with a "level insufficient" error.

3.  **GM Character (Dev Mode OFF):**
    -   Stop the backend, set `LS_ENABLE_DEV_GM=false` (or remove it), and restart.
    -   Log in with the `gm` account and select a character below level 10.
    -   Interact with "Mentor Arion".
    -   **Expected Result:** The class selection should be rejected. The GM bypass is disabled by the server configuration. The "GM TEST" badge will not appear.

## 4. Security Warnings

- **NEVER** enable `LS_ENABLE_DEV_GM` in a production environment.
- **NEVER** hardcode or commit passwords.
- **NEVER** trust the client to determine GM status. Authority is 100% server-side.
- The client-side "GM TEST" badge is purely a visual indicator and grants no special permissions.