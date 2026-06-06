<a href="https://www.uzi.sh">
  <img src="https://trieve.b-cdn.net/www.uzi.sh_.png">
</a>
<p align="center">
  <a
    href="https://cdn.trieve.ai/uzi-whitepaper.pdf"
    target="_blank"
    rel="noopener noreferrer"
    >Whitepaper
  </a>
</p>

## Installation

```bash
go install github.com/devflowinc/uzi@latest
```

Make sure that your GOBIN is in your PATH.

```sh
export PATH="$PATH:$HOME/go/bin"
```

## Features

- 🤖 Run multiple AI coding agents in parallel
- 🌳 Automatic Git worktree management for isolated development
- 🖥️ Tmux session management for each agent
- 🚀 Automatic development server setup with port management
- 📊 Real-time monitoring of agent status and code changes
- 🔄 Automatic handling of agent prompts and confirmations
- 🎯 Easy checkpoint and merge of agent changes

## Prerequisites

- **Git**: For version control and worktree management
- **Tmux**: For terminal session management
- **Go**: For installing
- **Your AI tool of choice**: Such as `claude`, `codex`, etc.

## Configuration

### uzi.yaml

Create a `uzi.yaml` file in your project root to configure Uzi:

```yaml
devCommand: cd astrobits && yarn && yarn dev --port $PORT
portRange: 3000-3010
```

#### Configuration Options

- **`devCommand`**: The command to start your development server. Use `$PORT` as a placeholder for the port number.
  - Example for Next.js: `npm install && npm run dev -- --port $PORT`
  - Example for Vite: `npm install && npm run dev -- --port $PORT`
  - Example for Django: `pip install -r requirements.txt && python manage.py runserver 0.0.0.0:$PORT`
- **`portRange`**: The range of ports Uzi can use for development servers (format: `start-end`)

**Important**: The `devCommand` should include all necessary setup steps (like `npm install`, `pip install`, etc.) as each agent runs in an isolated worktree with its own dependencies.

## Basic Workflow

1. **Start agents with a task:**

   ```bash
   uzi prompt --agents claude:3,codex:2 "Implement a REST API for user management with authentication"
   ```

2. **Run uzi auto**

   uzi auto automatically presses Enter to confirm all tool calls

   ```
   uzi auto
   ```

3. **Monitor agent progress:**

   ```bash
   uzi ls -w  # Watch mode
   ```

4. **Send additional instructions:**

   ```bash
   uzi broadcast "Make sure to add input validation"
   ```

5. **Merge completed work:**
   ```bash
   uzi checkpoint funny-elephant "feat: add user management API"
   ```

## Commands

### `uzi prompt` (alias: `uzi p`)

Creates new agent sessions with the specified prompt.

```bash
uzi prompt --agents claude:2,codex:1 "Build a todo app with React"
```

**Options:**

- `--agents`: Specify agents and counts in format `agent:count[,agent:count...]`
  - Use `random` as agent name for random agent names
  - Example: `--agents claude:2,random:3`

### `uzi ls` (alias: `uzi l`)

Lists all active agent sessions with their status.

```bash
uzi ls       # List active sessions
uzi ls -w    # Watch mode - refreshes every second
```

```
AGENT    MODEL  STATUS    DIFF  ADDR                     PROMPT
brian    codex  ready  +0/-0  http://localhost:3003  make a component that looks similar to @astrobits/src/components/Button/ that creates a Tooltip in the same style. Ensure that you include a reference to it and examples on the main page.
gregory  codex  ready  +0/-0  http://localhost:3001  make a component that `
```

### `uzi auto` (alias: `uzi a`)

Monitors all agent sessions and automatically handles prompts.

```bash
uzi auto
```

**Features:**

- Auto-presses Enter for trust prompts
- Handles continuation confirmations
- Runs in the background until interrupted (Ctrl+C)

### `uzi kill` (alias: `uzi k`)

Terminates agent sessions and cleans up resources.

```bash
uzi kill agent-name    # Kill specific agent
uzi kill all          # Kill all agents
```

### `uzi run` (alias: `uzi r`)

Executes a command in all active agent sessions.

```bash
uzi run "git status"              # Run in all agents
uzi run --delete "npm test"       # Run and delete the window after
```

**Options:**

- `--delete`: Remove the tmux window after running the command

### `uzi broadcast` (alias: `uzi b`)

Sends a message to all active agent sessions.

```bash
uzi broadcast "Please add error handling to all API calls"
```

### `uzi checkpoint` (alias: `uzi c`)

Makes a commit and rebases changes from an agent's worktree into your current branch.

```bash
uzi checkpoint agent-name "feat: implement user authentication"
```

### `uzi reset`

Removes all Uzi data and configuration.

```bash
uzi reset
```

**Warning**: This deletes all data in `~/.local/share/uzi`

### Advanced Usage

**Running different AI tools:**

```bash
uzi prompt --agents=claude:2,aider:2,cursor:1 "Refactor the authentication system"
```

**Using random agent names:**

```bash
uzi prompt --agents=random:5 "Fix all TypeScript errors"
```

**Running tests across all agents:**

```bash
uzi run "npm test"
```


## FAQ

### What is Uzi?

Uzi is a CLI tool for running large numbers of coding agents in parallel with git worktrees. It enables you to spin up multiple AI coding assistants (Claude Code, Cursor, Codex, etc.) to work on different tasks simultaneously, then merge their results back together.

### Key Features

| Feature | Description |
|---------|-------------|
| **Parallel Execution** | Run multiple agents concurrently on separate worktrees |
| **Git Worktrees** | Each agent works in isolated worktree to avoid conflicts |
| **Multi-Agent Support** | Works with Claude Code, Cursor, Codex, Windsurf, and more |
| **Task Queue** | Manage and prioritize multiple development tasks |
| **Merge Automation** | Automatically merge completed work back to main |
| **Conflict Resolution** | Handles merge conflicts with AI assistance |

### Installation

```bash
# Install via cargo (Rust)
cargo install uzi

# Or build from source
git clone https://github.com/devflowinc/uzi.git
cd uzi
cargo build --release
```

### Requirements

| Requirement | Version |
|-------------|---------|
| **Rust** | 1.70+ |
| **Git** | 2.17+ (worktrees support) |
| **AI Agent CLI** | Claude Code, Cursor, or compatible |

### Quick Start

```bash
# Initialize uzi in your project
uzi init

# Add tasks for agents
uzi add "Implement feature X"
uzi add "Fix bug Y"
uzi add "Add tests for module Z"

# Run all agents in parallel
uzi run --parallel 3

# Monitor progress
uzi status

# Merge completed work
uzi merge
```

### Workflow Example

1. **Initialize**: Setup uzi configuration in your repo
2. **Add Tasks**: Define what each agent should work on
3. **Run Parallel**: Spawn multiple agents in isolated worktrees
4. **Monitor**: Track progress and catch errors early
5. **Merge**: Integrate completed work back to main branch

### Benefits vs Sequential Approach

| Metric | Sequential | Uzi (Parallel) |
|--------|-----------|----------------|
| **Time** | 3 hours | 1 hour (3x faster) |
| **Context Switching** | High | Low (isolated worktrees) |
| **Conflicts** | Frequent | Rare (separate branches) |
| **Productivity** | 1 agent | N agents simultaneously |

### Is Uzi Free?

Yes! Uzi is open-source. Free to use in personal and commercial projects.

### Help & Resources

| Resource | Link |
|----------|------|
| **Documentation** | [uzi.sh](https://www.uzi.sh) |
| **Whitepaper** | [cdn.trieve.ai/uzi-whitepaper.pdf](https://cdn.trieve.ai/uzi-whitepaper.pdf) |
| **GitHub Issues** | [github.com/devflowinc/uzi/issues](https://github.com/devflowinc/uzi/issues) |
| **Discord** | [discord.gg/uzi](https://discord.gg/uzi) |
