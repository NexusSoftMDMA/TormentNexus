# Top 100 Hermes Agent Add-ons, Plugins, & Tools

Compiled based on Nous Research Hermes Agent ecosystem, community repositories, and high-value MCP (Model Context Protocol) servers mapped for agentic use.

## Category 1: Official & Core Plugins (1-10)
1. **Honcho Memory Provider**: Official memory abstraction provider for episodic context.
2. **Context Engine Plugin**: Manages semantic workspace boundaries for token efficiency.
3. **Nous Research CLI Hook**: Integrates terminal notifications for human-in-the-loop gates.
4. **Local Shell Runner**: Secure sandboxed python/bash execution.
5. **Dynamic MCP Adapter**: Automatically registers and prefixes arbitrary MCP tools.
6. **Filesystem Workspace Watcher**: Syncs changes in current directory as active context.
7. **Cost Control & Budgeting**: Limits daily token costs and API calls.
8. **Goal Tracker & Decomposition**: Displays dynamic progress bars for recursive tasks.
9. **GitHub Sync Plugin**: Natively commits, pulls, and pushes code changes.
10. **Aider Dev Helper**: Integrates aider git-diff style edits natively.

## Category 2: Community Extensions (11-30)
11. **wondelai/skills**: Public library of shared agent instructions.
12. **42-evey/cost-controller**: Custom plugin for real-time cost calculation.
13. **browser-use-hermes**: Connects the browser-use python toolset for automated browsing.
14. **tavily-web-search**: Web search engine integration for factual grounding.
15. **arxiv-scholar**: Biomedical and scientific paper search.
16. **sqlite-memory-bank**: Local SQLite-backed vector index.
17. **slack-alert-bridge**: Posts progress updates to Slack channels.
18. **excalidraw-viewer**: Renders UI prototypes on an Excalidraw canvas.
19. **desktop-control-adb**: Controls Android Emulators using adb command tools.
20. **openai-tts-say**: Text-to-speech speaker for interactive shell output.
21. **github-copilot-relay**: Interfaces with Copilot CLI tool for suggestions.
22. **dbhub-sql-explorer**: Connects to multiple database engines to execute queries.
23. **chunkhound-semantic**: Code indexing and symbol searching engine.
24. **notebooklm-notebooks**: Creates and queries source notebooks.
25. **vibe-check-observer**: Analyzes user tone and adjusts agent communication.
26. **supermemory-store**: Stores page clips and thoughts to SuperMemory.
27. **prism-quality-gate**: Automatically scans codebase for linting issues.
28. **windows-registry-inspect**: Inspects Windows registry and services for diagnostics.
29. **mindsdb-predictor**: Generates ML predictions from database tables.
30. **playwright-stealth**: Stealth headless browser driver.

## Category 3: High-Value MCP Adaptations (31-100)
Exposed to Hermes via the Dynamic MCP Adapter prefixing:

31. **mcp_github_list_repos**: Lists repositories for the authenticated user.
32. **mcp_github_get_repo**: Retrieves details of a specific repository.
33. **mcp_github_create_issue**: Creates an issue in a repository.
34. **mcp_github_list_issues**: Lists issues in a repository.
35. **mcp_github_create_pr**: Creates a pull request.
36. **mcp_github_code_search**: Searches code across GitHub.
37. **mcp_github_get_file_contents**: Reads file content from GitHub.
38. **mcp_github_create_or_update_file**: Edits files on GitHub.
39. **mcp_github_list_branches**: Lists branches.
40. **mcp_github_list_workflows**: Lists GitHub Action workflows.
41. **mcp_github_trigger_workflow**: Runs a GitHub Action workflow.
42. **mcp_github_copilot_chat**: Chats with Copilot API.
43. **mcp_supabase_list_projects**: Lists Supabase database projects.
44. **mcp_supabase_execute_sql**: Executes raw SQL queries.
45. **mcp_supabase_select_rows**: Queries rows from a table.
46. **mcp_supabase_insert_rows**: Inserts new rows.
47. **mcp_supabase_update_rows**: Modifies existing rows.
48. **mcp_supabase_delete_rows**: Deletes rows.
49. **mcp_supabase_list_tables**: Lists tables.
50. **mcp_supabase_invoke_function**: Runs a Supabase edge function.
51. **mcp_desktop_execute_command**: Runs shell commands.
52. **mcp_desktop_read_file**: Reads files from desktop.
53. **mcp_desktop_write_file**: Writes files to desktop.
54. **mcp_desktop_list_directory**: Lists directory contents.
55. **mcp_desktop_search_files**: Searches local files.
56. **mcp_desktop_get_system_info**: Retrieves local specs.
57. **mcp_desktop_list_processes**: Lists running processes.
58. **mcp_desktop_kill_process**: Kills a process.
59. **mcp_desktop_execute_script**: Runs python/node scripts.
60. **mcp_desktop_open_file**: Opens a file with default editor.
61. **mcp_gemini_chat**: Connects to Gemini models.
62. **mcp_gemini_vision**: Connects to Gemini vision models.
63. **mcp_gemini_embeddings**: Generates embeddings.
64. **mcp_dbhub_list_databases**: Lists databases.
65. **mcp_dbhub_execute_query**: Executes SQL queries.
66. **mcp_conport_get_context**: Fetches context from ConPort.
67. **mcp_conport_update_context**: Updates ConPort context.
68. **mcp_conport_log_decision**: Logs architecture decisions.
69. **mcp_conport_get_decisions**: Gets logged decisions.
70. **mcp_chunkhound_index**: Indexes code for ChunkHound.
71. **mcp_chunkhound_search**: Searches ChunkHound vector db.
72. **mcp_notebooklm_create_notebook**: Creates a notebook.
73. **mcp_notebooklm_query_notebook**: Queries a notebook.
74. **mcp_vibe_check_analyze**: Analyzes text emotions.
75. **mcp_supermemory_add**: Adds clips to memory.
76. **mcp_supermemory_search**: Searches clips.
77. **mcp_probe_search_code**: Searches code via Probe.
78. **mcp_probe_find_symbol**: Locates code symbols.
79. **mcp_cipher_add_memory**: Adds memories in Cipher.
80. **mcp_cipher_search_memory**: Searches Cipher memories.
81. **mcp_deepcontext_analyze**: Analyzes codebases in DeepContext.
82. **mcp_windows_mcp_get_system_info**: Diagnostic utility.
83. **mcp_prism_analyze_quality**: Inspects code quality.
84. **mcp_prism_suggest_refactor**: Proposes code refactors.
85. **mcp_taskmaster_create_task**: Adds tasks in TaskMaster.
86. **mcp_taskmaster_list_tasks**: Lists tasks.
87. **mcp_taskmaster_update_status**: Changes task status.
88. **mcp_arxiv_search**: Queries scientific papers.
89. **mcp_semantic_scholar_search**: Queries Semantic Scholar.
90. **mcp_mem0_add_memory**: Stores memory entries.
91. **mcp_alpaca_place_order**: Places stock orders.
92. **mcp_alpha_vantage_quote**: Gets stock quotes.
93. **mcp_huggingface_search_models**: Searches models.
94. **mcp_semgrep_scan**: Runs security scans.
95. **mcp_octagon_research**: Research intelligence utility.
96. **mcp_chroma_query**: Queries vector database.
97. **mcp_basic_memory_write**: Writes key-value memory.
98. **mcp_mindsdb_predict**: Generates predictions.
99. **mcp_nws_get_forecast**: Gets local weather forecasts.
100. **mcp_say_tts**: Natively speaks prompt results.
