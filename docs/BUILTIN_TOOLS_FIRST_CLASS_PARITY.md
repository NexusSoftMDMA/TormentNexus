# First-Class IDE/CLI Tool Parity (Absolute 1:1 Compatibility)

## The Philosophy
Large Language Models (LLMs) like Claude 3.7 Sonnet, GPT-4o, and others are heavily fine-tuned on the exact tool signatures used by the most popular coding environments (Claude Code, Cursor, Aider, Windsurf, OpenCode).

When an LLM sees a familiar tool signature (e.g., `str_replace_editor`, `glob`, `bash`), its performance, reasoning, and accuracy increase dramatically because it falls into well-worn cognitive grooves established during RLHF and instruction tuning. 

**tormentnexus's mandate is Absolute 1:1 Parity.** We do not rename, restructure, or "improve" these core tool signatures. If a model expects Claude Code's `Bash` tool, tormentnexus provides a tool that is byte-for-byte identical in schema and behavior.

---

## 1. Claude Code Parity (Primary Target)

Claude 3.7 Sonnet is explicitly trained to use the following internal tools. tormentnexus implements these exactly as expected:

### `bash`
- **Description**: Executes a bash command and returns the output.
- **Schema**:
  ```json
  {
    "command": "string (required)"
  }
  ```
- **tormentnexus Implementation**: Mapped directly to our isolated process execution engine, preserving standard output, standard error, and exit codes identically to Claude Code.

### `glob`
- **Description**: Searches for files matching a glob pattern.
- **Schema**:
  ```json
  {
    "pattern": "string (required)"
  }
  ```
- **tormentnexus Implementation**: Fast Rust-backed `ignore` crawler, returning absolute paths exactly formatted as the model expects.

### `grep_search`
- **Description**: Searches for a regular expression pattern within file contents.
- **Schema**:
  ```json
  {
    "pattern": "string (required)",
    "dir_path": "string (optional)",
    "include_pattern": "string (optional)"
  }
  ```
- **tormentnexus Implementation**: Backed by `ripgrep`, ensuring exact argument parsing and line-number formatting.

### `file_read` / `read_file`
- **Description**: Reads file contents with line limits.
- **Schema**:
  ```json
  {
    "file_path": "string (required)",
    "start_line": "number (optional)",
    "end_line": "number (optional)"
  }
  ```

### `str_replace_editor` / `replace`
- **Description**: Replaces exact string matches within a file.
- **Schema**:
  ```json
  {
    "file_path": "string (required)",
    "old_string": "string (required)",
    "new_string": "string (required)"
  }
  ```

---

## 2. Aider Parity

Aider relies heavily on specific diff formats (Unified Diff, Search/Replace blocks).
- **`ask_aider`**: (Supported) - allows the model to spin up a sub-agent.
- **`run_tests`**: (Supported) - maps to tormentnexus's `AutoTestReactor`.

---

## 3. Cursor & Windsurf Parity

These editors provide specific contextual tools.
- **`get_cursor_context`**: Maps to tormentnexus's `get_project_context`.
- **`list_workspace_symbols`**: Maps to tormentnexus's `lspTools` for fast symbol extraction.

---

## Implementation Strategy

tormentnexus intercepts tool requests at the router level. When an LLM sends a payload for `bash` (expecting Claude Code's implementation), the tormentnexus router recognizes the 1:1 alias and routes it to our secure `executionEnvironment`, returning the exact JSON structure the model was fine-tuned to receive.

There is no "translation layer" for the LLM to learn. To the model, tormentnexus *is* Claude Code, it *is* Cursor, it *is* Aider.
