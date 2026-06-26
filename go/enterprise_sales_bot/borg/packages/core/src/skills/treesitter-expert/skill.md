---
name: TreeSitter Expert
description: Expert knowledge for building high-performance Neovim plugins using nvim-treesitter.
---

# TreeSitter Expert Skill

## Purpose
You are an expert in Neovim plugin architecture, specifically regarding `nvim-treesitter`. You provide patterns for high-performance, query-based syntax analysis and virtual text display.

## Capabilities
- **Architecture**: Advises on the "Three-Tier Plugin Structure" (Minimal entry, Lazy Lua, Query separation).
- **Queries**: Write complex scheme queries for `nvim-treesitter` to extract semantic nodes.
- **Performance**: Implement advanced caching (buffer tick memoization) and throttling (debouncing) patterns.
- **Display**: Use `api.nvim_buf_set_extmark` for "overlay" or "eol" virtual text.

## Core Knowledge (Reference)

### 1. Module System Architecture

**Three-Tier Plugin Structure:**
```
your-plugin/
├── plugin/your-plugin.lua          # Minimal entry point (VimL commands only)
├── lua/your-plugin/
│   ├── init.lua                    # Main module with setup()
│   ├── config.lua                  # Configuration management
│   ├── highlight.lua               # Display logic (extmarks/virtual text)
│   └── query.lua                   # TreeSitter query execution
└── queries/
    └── go/
        └── custom-queries.scm      # TreeSitter query definitions
```

**Entry Point Pattern (plugin/your-plugin.lua):**
```lua
-- Lazy load protection
if vim.g.loaded_your_plugin then
  return
end
vim.g.loaded_your_plugin = true

-- Define commands that load on-demand
command! YourPluginToggle lua require("your-plugin").toggle()
command! YourPluginEnable lua require("your-plugin").enable()
```

### 2. Configuration Management

**Deep Merge Pattern:**
```lua
function M.setup(options)
  M.options = vim.tbl_deep_extend("force", {}, defaults, M.options or {}, options or {})
end
```

### 3. TreeSitter Query System

**Query Execution Pattern:**
```lua
function M.get_error_nodes(bufnr)
  local parser = parsers.get_parser(bufnr, "go")
  local query = ts.query.get("go", "error-collapse")
  
  parser:for_each_tree(function(tree, lang_tree)
    local root = tree:root()
    for id, node, metadata in query:iter_captures(root, bufnr, 0, -1) do
      -- Process capture
    end
  end)
end
```

### 4. Performance Optimization

**Buffer Tick Memoization:**
```lua
function M.memoize_by_buf_tick(fn)
  local cache = setmetatable({}, { __mode = "kv" })
  return function(bufnr)
    local tick = vim.api.nvim_buf_get_changedtick(bufnr)
    if cache[bufnr] and cache[bufnr].tick == tick then
      return cache[bufnr].result
    end
    -- compute...
  end
end
```
