# Zed Assistant MCP

Zed Assistant MCP is a remote [Model Context Protocol](https://modelcontextprotocol.io) server that lets an AI assistant (Claude, ChatGPT, or any other MCP-compatible client) manage one or more **Project Zomboid servers**: check server status, moderate players, edit server/sandbox configuration, tail the game log, and run raw admin commands - all through natural-language conversation instead of RCON consoles or hand-edited `.ini`/`.lua` files.

It works with any Project Zomboid server that has RCON enabled and writes the usual `Server/*.ini` / `Server/*_SandboxVars.lua` files - that includes both servers run via the standalone **dedicated server** and servers **hosted in-game** ("Host" from the game's main menu). Both produce similar on-disk layout and RCON interface, so Zed Assistant MCP treats them identically.

It runs as a single Go binary that exposes:

- an **OAuth 2.1 authorization server** (login + token issuing), so only accounts you approve can connect, and
- an **MCP endpoint** (`/mcp`) that speaks Streamable HTTP and requires a valid bearer token on every request.

The server itself holds no game logic - every tool call is translated into either an RCON command sent to your running Project Zomboid server, or a direct read/write of the server's own config files and player database on disk.

---

## Table of contents

- [How it fits together](#how-it-fits-together)
- [Features](#features)
- [Available tools](#available-tools)
- [Requirements](#requirements)
- [Installation](#installation)
- [Preparing your Project Zomboid server](#preparing-your-project-zomboid-server)
- [Configuration](#configuration)
  - [Config file location & overrides](#config-file-location--overrides)
  - [Full reference](#full-reference)
  - [Minimal working example](#minimal-working-example)
- [Running the server](#running-the-server)
- [Connecting an AI / MCP client](#connecting-an-ai--mcp-client)
- [Access control model](#access-control-model)
- [Security notes](#security-notes)
- [Troubleshooting](#troubleshooting)
- [License](#license)

---

## How it fits together

```
                     ┌────────────────────────────┐
  AI client ────────►│     Zed Assistant MCP      │
 (Claude, etc.)      │  /auth   OAuth 2.1 server  │
      ▲              │  /mcp    MCP tool endpoint │
      │              └─────────────┬──────────────┘
      └── OAuth login (browser) ───┘
                                  │  RCON (admin commands)
                                  │  filesystem (ini / lua / sqlite)
                                  ▼
                          ┌─────────────────────┐
                          │  Project Zomboid    │
                          │  server(s)          │
                          │  (dedicated or      │
                          │   in-game hosted)   │
                          └─────────────────────┘
```

Zed Assistant MCP must run somewhere that has:
1. **Filesystem access** to each Project Zomboid server's data/"home" directory (to read/write `Server/<name>.ini`, `Server/<name>_SandboxVars.lua`, `db/<name>.db`, and the console log), and
2. **Network access** to each server's RCON port.

In practice this usually means running it on the same machine as your Project Zomboid server(s), though a remote setup works too as long as RCON is reachable and the home directories are mounted/shared.

---

## Features

- **Multi-instance management** - configure any number of Project Zomboid servers, each independently, and manage all of them through one deployment.
- **Per-user, per-instance access control** - every instance has its own allow-list of user emails; a user only sees and can act on the instances they're allowed to use.
- **Live server status** - online/offline state, connected players (with role and Steam ID when known), player count and max players.
- **Configuration read/write** - read and update both the flat `ServerOptions` `.ini` (ports, player limits, mods, etc.) and the nested `SandboxVars.lua` (zombie settings, loot, weather, time, and anything mods add), with:
  - key filtering (with `*` wildcards) so you don't have to fetch the entire file,
  - type-checked, all-or-nothing updates - if any key/value in a batch update is invalid, nothing is written,
  - optional **live apply** of server `.ini` changes via RCON `reloadoptions` (sandbox changes always require a restart - there's no safe hot-reload for those).
- **Player moderation** - kick, ban, and unban by username or Steam ID.
- **Player state management** - give items, grant XP, toggle invisibility/noclip.
- **User account management** - allow-list Steam IDs, add server login accounts, change access levels, set passwords.
- **Server lifecycle control** - reload Lua, save the world, or quit the server (quit requires an explicit confirmation flag, since it disconnects everyone).
- **Broadcast messages** to all connected players.
- **Game log tailing** - read the most recent console log lines, newest-first, with case-insensitive substring/wildcard filtering.
- **Raw admin command escape hatch** - for anything not covered by a dedicated tool.
- **Built-in login system** - no external identity provider required; users and passwords are defined directly in the config file. This `local` provider is the only one available today and is best suited to testing/small setups; a future version is planned to add login via external identity providers (e.g. Google), which will be the recommended option for production use once available.
- **Standards-based auth** - OAuth 2.1 with mandatory PKCE and dynamic client registration via Client ID Metadata Documents (CIMD), so modern MCP clients can connect without manual client registration.
- **Safe by construction in a few places**: destructive actions (server quit) require explicit confirmation, and RCON password/port are always read live from the Zomboid server's own `.ini` file - never duplicated/stored in this server's config.

---

## Available tools

These are the tools exposed to the connected AI assistant over MCP. The assistant decides when to call them based on your conversation - you don't call them directly.

| Tool | Purpose | Notes |
|---|---|---|
| `list-zomboid-instances` | List the Project Zomboid instances the current user has access to. | Read-only. |
| `get-server-status` | Online state, connected players, player count/max. | Read-only. Cheap - call before targeting a player, to confirm they're online and get their exact username. |
| `read-zomboid-config` | Read `server` (.ini) or `sandbox` (SandboxVars.lua) settings, optionally filtered by key (wildcards supported). | Read-only. |
| `update-zomboid-config` | Partially update `server` or `sandbox` settings. | Validates the whole batch before writing anything. `applyLive` (server-only) hot-applies via RCON `reloadoptions`. |
| `execute-raw-admin-command` | Run any raw Project Zomboid admin command string. | Escape hatch; unvalidated. Run the `help` admin command (via this same tool) to see available commands. |
| `broadcast-server-message` | Send a message to all connected players. | |
| `moderate-player` | `kick`, `ban`, or `unban` a player, by username or Steam ID. | Returns updated server status after the action. |
| `manage-player-state` | `additem`, `addxp`, `invisible`, `noclip` for a specific player. | Each action has its own required fields (item module ID, perk name, XP amount, or on/off). |
| `manage-user-account` | `addsteamid`, `removesteamid`, `adduser`, `setaccesslevel`, `setpassword`. | Server-side account/whitelist management, not this MCP server's own login accounts. |
| `manage-server-lifecycle` | `reloadalllua`, `save`, `quit`. | `quit` requires `confirm: true` - it disconnects every player. |
| `get-zomboid-game-logs` | Tail the console log, most recent first, with optional filter. | Log lines have no real timestamps (only a per-boot counter), so "since" filtering isn't supported. |

All tools require selecting an `instanceId` (except `list-zomboid-instances`), and every call is checked against that instance's user allow-list before anything happens.

---

## Requirements

- One or more running **Project Zomboid servers** (dedicated server or in-game hosted), each with RCON enabled.
- To run a prebuilt binary: nothing extra - Linux (amd64) and Windows (amd64) binaries are published on the project's GitHub Releases page.
- To build from source: **Go 1.26+**.
- Network reachability from this server to each Project Zomboid server's RCON port, and filesystem access to each server's data directory (see [Preparing your Project Zomboid server](#preparing-your-project-zomboid-server)).

---

## Installation

### Option A: download a prebuilt binary

Grab the binary for your platform from the Releases page (`zed-assistant-mcp-linux-amd64` or `zed-assistant-mcp-windows-amd64.exe`), verify it against the published `checksums.txt` if you like, and make it executable.

### Option B: build from source

```bash
git clone https://github.com/zed-assistant/mcp.git
cd mcp
go build -o zed-assistant-mcp ./cmd/server
```

Either way, you get a single self-contained binary that only needs a config file to run.

---

## Preparing your Project Zomboid server

Zed Assistant MCP doesn't run or install Project Zomboid - it manages a server that's already set up and (usually) running, whether that's the standalone **dedicated server** or a session **hosted in-game** via "Host" in the game's main menu. Both write similar `Server/` file layout and support the same RCON interface, so everything below applies equally to either. For each server instance you want to manage, you need to know:

1. **Home directory** - the Zomboid data directory the server writes to (the folder containing `Server/`, `db/`, and the console log). For the dedicated server binary this is typically the `Zomboid` folder under the account that runs it, or wherever you pointed it with `-cachedir`-style startup options; for an in-game hosted server it's the same `Zomboid` folder under the hosting player's own user profile - on a typical Windows install that's `C:/Users/<username>/Zomboid`. This is the `home_dir` config value.
2. **Server name** - the name given to the server (via `-servername <name>` for the dedicated binary, or the server name chosen in the in-game "Host" dialog; default `servertest` if never changed), which determines the config file names: `Server/<name>.ini` and `Server/<name>_SandboxVars.lua`. This is the `server_name` config value.
3. **RCON enabled** - RCON must be turned on, with a **strong password** set, in that server's own `<name>.ini` (`RCONPort` / `RCONPassword`) - this option exists for both dedicated and in-game hosted servers. Set a long, unique `RCONPassword` before connecting Zed Assistant MCP to it: this password grants full admin console access to the server, so a weak or reused one undermines every other access control in this document. Zed Assistant MCP reads the RCON port and password **directly from that file** every time it needs them - you never enter them in this server's own config.
4. **RCON reachability** - the host/IP where RCON is reachable from wherever Zed Assistant MCP runs. This is the `rcon_host` config value. RCON has no transport encryption of its own, so keep it on a trusted network (localhost, private LAN, or a VPN/SSH tunnel) - don't expose the RCON port to the public internet.

---

## Configuration

### Config file location & overrides

The binary takes a single flag:

```bash
./zed-assistant-mcp --config /path/to/config.yml
```

(default: `config.yml` in the current directory.)

Configuration is layered, in this order (each layer overrides the one before it):

1. **Built-in defaults** (ports, timeouts, etc.).
2. **Your YAML config file.** Any `${VAR_NAME}` in the file is expanded from the process environment before parsing - this is how you keep secrets like passwords out of the file itself (see the example below).
3. **Environment variables** with the prefix `ZED_ASSISTANT_MCP_`. Nesting is expressed with a double underscore, matching the YAML structure - e.g. `ZED_ASSISTANT_MCP_LOGGER__JSON_FORMAT=true` sets `logger.json_format`. This layer can override *any* config value directly, not just fill in placeholders.

### Full reference

```yaml
logger:
  json_format: false      # true = structured JSON logs (good for log aggregators); false = colored console output
  disable_color: false    # true = plain-text console output (no ANSI colors), only relevant when json_format is false

server:
  port: 8080              # TCP port the HTTP server listens on
  external_url: ""         # REQUIRED. Public base URL this server is reached at, e.g. "https://zomboid-mcp.example.com".
                           # Used to build OAuth endpoints, redirect targets, and the MCP resource identifier - it must
                           # match how your AI client will actually reach this server, including scheme and port.
                           # For a local/single-machine instance this can be something like "http://localhost:8080"
                           # (see the local development note under "Minimal working example" below).

oauth2:
  signing_secret: ""       # Base64-encoded 32-byte secret used to sign tokens. If left empty, a random one is generated
                           # on every startup, which invalidates all previously issued tokens on restart. Set a fixed
                           # value for production so logins/tokens survive restarts.
  id_token_signing_key: "" # Base64-encoded PEM-encoded RSA private key (PKCS#1 or PKCS#8), used to sign OIDC ID tokens.
                           # Same restart caveat as signing_secret if left empty.
  pending_auth_ttl: 10m    # How long a browser login flow has to complete before it expires.
  idp:
    type: "local"          # Identity provider type. Only "local" is currently supported. It's intended for testing/small
                           # setups; support for external identity providers (e.g. Google) is planned for a future
                           # version and will be the recommended option for production use once available.
    secure_cookie: true    # Marks the short-lived login cookie as Secure (HTTPS-only). Only set to false for local
                           # http:// testing - never in production.
    local:
      users:                       # Accounts allowed to log in to this MCP server at all.
        - username: "you@example.com"  # Must be a valid email address; this becomes the user's identity everywhere else
                                        # in the config (e.g. zomboid.instances.*.users).
          password: "a-strong-password" # Compared as plain text against what's typed at login - use a long, unique
                                         # value, and prefer injecting it via ${ENV_VAR} rather than committing it.

zomboid:
  instances:                       # Map of instance ID -> instance config. The ID (map key) is an internal identifier
                                   # used in tool calls and URLs; it never needs to match anything on the Zomboid side.
    "my-server-id":
      name: "My Server"            # Human-readable display name shown to the AI/user.
      home_dir: "/srv/zomboid/my-server"  # Absolute path to that server's Zomboid data directory (see "Preparing your
                                           # Project Zomboid server" above). Must be readable/writable by this process.
      server_name: "servertest"    # Matches the server's name (dedicated server's -servername argument, or the name
                                   # chosen when hosting in-game). Defaults to "servertest" if omitted, matching
                                   # Project Zomboid's own default.
      rcon_host: "localhost"       # Hostname/IP where this instance's RCON is reachable. The port and password are
                                   # read live from that instance's own <server_name>.ini - not configured here.
      users:                       # Emails allowed to use this specific instance (subset of, or matching,
                                   # oauth2.idp.local.users). Must contain at least one entry.
        - "you@example.com"
```

### Minimal working example

```yaml
server:
  external_url: "https://zomboid-mcp.example.com" # or http://localhost:8080 for local testing

oauth2:
  idp:
    type: "local"
    local:
      users:
        - username: "admin@example.com"
          password: "${MCP_ADMIN_PASSWORD}"

zomboid:
  instances:
    "main":
      name: "Main Server"
      home_dir: "/home/pzuser/Zomboid"
      rcon_host: "localhost"
      users:
        - "admin@example.com"
```

Run it with the password supplied out-of-band:

```bash
MCP_ADMIN_PASSWORD='a-strong-password' ./zed-assistant-mcp --config config.yml
```

For local development/testing without HTTPS, also set `oauth2.idp.secure_cookie: false` and use an `external_url` like `http://localhost:8080`.

---

## Running the server

```bash
./zed-assistant-mcp --config /path/to/config.yml
```

- Logs go to stdout (console-formatted or JSON, per `logger.json_format`).
- The process shuts down gracefully on `SIGINT`/`SIGTERM` (e.g. Ctrl-C, or a service manager's stop signal), finishing in-flight requests before exiting.
- Run it under your process supervisor of choice (systemd, a container, a process manager) like any other long-running HTTP service; there's no bundled Docker image or service unit yet (an official Docker image is planned for a future version), so set one up appropriately for your environment in the meantime.
- Put it behind a reverse proxy that terminates TLS (recommended) so `external_url` can be an `https://` address; OAuth cookies and dynamic client registration both expect HTTPS in production.

---

## Connecting an AI / MCP client

Zed Assistant MCP exposes a standard **remote MCP server** at `<external_url>/mcp`, using Streamable HTTP transport and OAuth 2.1 for authorization. Any MCP client that supports remote HTTP servers with OAuth (including dynamic/CIMD-based client registration) can connect:

1. In your MCP client, add a new remote server pointing at `https://<external_url>/mcp`.
2. The client discovers auth requirements automatically from:
   - `https://<external_url>/.well-known/oauth-protected-resource/mcp`
   - `https://<external_url>/.well-known/oauth-authorization-server`
3. The client registers itself using its own **Client ID Metadata Document** (an `https://` URL that serves a small JSON document describing the client) - no manual "create an OAuth app" step is needed on your side. Redirect URIs in that document must be `https://` or a loopback address (`localhost`/`127.0.0.1`).
4. Your browser opens to a login prompt (HTTP Basic Auth dialog) - enter one of the `oauth2.idp.local.users` username/password pairs.
5. On success, the client receives an access token (and, if it requested the `offline_access` scope, a refresh token) tied to that user's email. Every subsequent MCP tool call is authorized against that identity, per-instance, as described below.

If your client doesn't support CIMD-based dynamic registration or remote OAuth-protected MCP servers, it won't be able to connect - this server does not support static/manual client registration or unauthenticated access.

---

## Access control model

There are two independent layers of access control:

1. **Who can log in at all** - controlled by `oauth2.idp.local.users`. Only these username/password pairs can obtain a token.
2. **Who can use which Zomboid instance** - controlled by each entry's `zomboid.instances.<id>.users` list. A logged-in user only sees the instances (via `list-zomboid-instances`) and can only call tools against the instances where their email appears in that instance's `users` list.

A user can be a valid login without having access to any instance (they'd authenticate successfully but every tool call would be rejected), and instance access lists can overlap or differ freely between instances - there's no separate "admin" role; access is purely per-instance, per-email.

---

## Security notes

- **Passwords are compared in plain text** from the config file (there's no hashing) - keep the config file's permissions restricted, use long/unique passwords, and prefer injecting them via `${ENV_VAR}` substitution rather than storing them in the file directly.
- **Set `oauth2.signing_secret` and `oauth2.id_token_signing_key` explicitly** in production. If left blank they're regenerated randomly on every startup, silently invalidating every previously issued token (users will need to log in again after every restart).
- **RCON has no transport encryption.** Keep `rcon_host` reachable only over a trusted network path (localhost, private network, VPN, or SSH tunnel) - never expose a Project Zomboid RCON port directly to the internet.
- **Client ID Metadata Document fetches are SSRF-guarded**: this server only dials public IP addresses when resolving a client's metadata document, refusing loopback/private/link-local targets, so a malicious `client_id` can't be used to reach your internal network.
- **`execute-raw-admin-command` is unvalidated** - anything reachable by an authorized instance user's assistant can be sent verbatim to the game server's RCON console. Treat instance access lists as equivalent to "trusted with full admin console access to that server."
- The `manage-server-lifecycle` `quit` action requires an explicit `confirm: true` precisely because it's irreversible mid-session (disconnects every player); the AI assistant is expected to ask for confirmation before setting it.

---

## Troubleshooting

- **Server fails to start with a validation error** (`server.external_url must be set`, `zomboid.instances must contain at least one instance`, etc.) - the printed message names the exact missing/invalid field; check it against the [Full reference](#full-reference).
- **"User has no access to the requested Project Zomboid server instance"** - the authenticated user's email isn't in that instance's `users` list (or the instance ID doesn't exist).
- **`get-server-status` reports `online: false`** instead of erroring - this means Zed Assistant MCP couldn't open an RCON connection (server is stopped, RCON is disabled, or `rcon_host`/network path is wrong); it's used as the "is it running" signal rather than a hard error.
- **"RCON password not found/empty in server config"** - RCON isn't enabled (or has no password) in that instance's `<server_name>.ini`; enable it in-game/server config and restart the Zomboid server.
- **Config update rejected with a list of problems, nothing changed** - `update-zomboid-config` validates the entire batch before writing anything; fix every listed key/value and retry.
- **Game log tool returns "Game log file was not found"** - neither `server-console.txt` nor `console.txt` exists directly under `home_dir`; double check `home_dir` points at the right directory.
- **MCP client can't connect / auth fails** - confirm `external_url` exactly matches the URL your client uses (scheme, host, and port), that it's reachable over HTTPS in production, and that your client supports CIMD-based dynamic client registration (see [Connecting an AI / MCP client](#connecting-an-ai--mcp-client)).

---

## License

MIT - see [LICENSE](LICENSE).
