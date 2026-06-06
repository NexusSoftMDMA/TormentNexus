package tools

/**
 * @file registry.go
 * @module go/internal/tools
 *
 * WHAT: Go-native registry for standard library and parity tools.
 * Maps tool names to their native Go implementations.
 */

import (
	"context"
	"fmt"
	"sync"
)

type ToolHandler func(ctx context.Context, args map[string]interface{}) (ToolResponse, error)

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]ToolHandler
}

func NewRegistry() *Registry {
	r := &Registry{
		handlers: make(map[string]ToolHandler),
	}
	r.registerAll()
	return r
}

func (r *Registry) registerAll() {
	// Native Handlers
	r.handlers["read_file"] = HandleRead
	r.handlers["write_file"] = HandleWrite
	r.handlers["edit_file"] = HandleEdit
	r.handlers["str_replace_editor"] = HandleEdit
	r.handlers["grep_search"] = HandleGrep
	r.handlers["search_files"] = HandleGrep
	r.handlers["glob"] = HandleGlob
	r.handlers["find_files"] = HandleGlob
	r.handlers["apply_patch"] = HandleApplyPatch
	r.handlers["multi_edit"] = HandleMultiEdit
	r.handlers["bash"] = HandleBash
	r.handlers["ls"] = HandleLS
	r.handlers["list_directory"] = HandleLS
	r.handlers["web_fetch"] = HandleWebFetch
	r.handlers["ingest_git"] = HandleGitIngest
	r.handlers["sqlite_get_catalog"] = HandleSqliteGetCatalog
	r.handlers["sqlite_execute"] = HandleSqliteExecute
	r.handlers["search"] = HandleDDGSearch
	r.handlers["fetch_content"] = HandleDDGFetchContent
	r.handlers["slack_list_channels"] = HandleSlackListChannels
	r.handlers["slack_post_message"] = HandleSlackPostMessage
	r.handlers["slack_reply_to_thread"] = HandleSlackReplyToThread
	r.handlers["slack_add_reaction"] = HandleSlackAddReaction
	r.handlers["slack_get_channel_history"] = HandleSlackGetChannelHistory
	r.handlers["slack_get_thread_replies"] = HandleSlackGetThreadReplies
	r.handlers["slack_get_users"] = HandleSlackGetUsers
	r.handlers["slack_get_user_profile"] = HandleSlackGetUserProfile

	// Filesystem MCP Tools
	r.handlers["read_text_file"] = HandleReadTextFile
	r.handlers["create_directory"] = HandleCreateDirectory
	r.handlers["list_directory"] = HandleListDirectory
	r.handlers["list_directory_with_sizes"] = HandleListDirectoryWithSizes
	r.handlers["directory_tree"] = HandleDirectoryTree
	r.handlers["move_file"] = HandleMoveFile
	r.handlers["get_file_info"] = HandleGetFileInfo
	r.handlers["search_files"] = HandleSearchFiles

	// Ollama MCP Tools (AI & LLM Integration)
	r.handlers["list_local_models"] = HandleListLocalModels
	r.handlers["local_llm_chat"] = HandleLocalLLMChat
	r.handlers["ollama_health_check"] = HandleOllamaHealthCheck
	r.handlers["system_resource_check"] = HandleSystemResourceCheck

	// TTS MCP Tools (Media & Design)
	r.handlers["say_tts"] = HandleSayTTS
	r.handlers["openai_tts"] = HandleOpenAITTS

	// Vercel MCP Tools (Cloud & DevOps)
	r.handlers["vercel_list_projects"] = HandleVercelListProjects
	r.handlers["vercel_get_project"] = HandleVercelGetProject
	r.handlers["vercel_list_deployments"] = HandleVercelListDeployments
	r.handlers["vercel_get_deployment"] = HandleVercelGetDeployment
	r.handlers["vercel_cancel_deployment"] = HandleVercelCancelDeployment
	r.handlers["vercel_list_env_vars"] = HandleVercelListEnvVars
	r.handlers["vercel_create_env_var"] = HandleVercelCreateEnvVar
	r.handlers["vercel_delete_env_var"] = HandleVercelDeleteEnvVar

	// DexPaprika MCP Tools (Finance & Crypto)
	r.handlers["getCapabilities"] = HandleDexPaprikaGetCapabilities
	r.handlers["getNetworks"] = HandleDexPaprikaGetNetworks
	r.handlers["getStats"] = HandleDexPaprikaGetStats
	r.handlers["search"] = HandleDexPaprikaSearch
	r.handlers["getNetworkDexes"] = HandleDexPaprikaGetNetworkDexes
	r.handlers["getNetworkPools"] = HandleDexPaprikaGetNetworkPools
	r.handlers["getDexPools"] = HandleDexPaprikaGetDexPools
	r.handlers["getNetworkPoolsFilter"] = HandleDexPaprikaGetNetworkPoolsFilter
	r.handlers["getPoolDetails"] = HandleDexPaprikaGetPoolDetails
	r.handlers["getPoolOHLCV"] = HandleDexPaprikaGetPoolOHLCV
	r.handlers["getPoolTransactions"] = HandleDexPaprikaGetPoolTransactions
	r.handlers["getTokenDetails"] = HandleDexPaprikaGetTokenDetails
	r.handlers["getTokenPools"] = HandleDexPaprikaGetTokenPools
	r.handlers["getTokenMultiPrices"] = HandleDexPaprikaGetTokenMultiPrices
	r.handlers["filterNetworkTokens"] = HandleDexPaprikaFilterNetworkTokens
	r.handlers["getTopTokens"] = HandleDexPaprikaGetTopTokens
	r.handlers["submitFeedback"] = HandleDexPaprikaSubmitFeedback

	// National Weather Service (NWS) MCP Tools (Weather & Location)
	r.handlers["nws_get_forecast"] = HandleNWSGetForecast
	r.handlers["nws_search_alerts"] = HandleNWSSearchAlerts
	r.handlers["nws_get_observations"] = HandleNWSGetObservations
	r.handlers["nws_find_stations"] = HandleNWSFindStations
	r.handlers["nws_list_alert_types"] = HandleNWSListAlertTypes
	r.handlers["nws_get_office_discussion"] = HandleNWSGetOfficeDiscussion
	r.handlers["nws_get_zone_forecast"] = HandleNWSGetZoneForecast

	// ast-grep-mcp Tools (Category 11)
	r.handlers["ast_grep_dump_syntax_tree"] = HandleDumpSyntaxTree
	r.handlers["ast_grep_test_match_code_rule"] = HandleTestMatchCodeRule
	r.handlers["ast_grep_find_code"] = HandleFindCode
	r.handlers["ast_grep_find_code_by_rule"] = HandleFindCodeByRule

	// PAL Tools (Category 12)
	r.handlers["pal_chat"] = HandlePalChat
	r.handlers["pal_thinkdeep"] = HandlePalThinkDeep
	r.handlers["pal_planner"] = HandlePalPlanner
	r.handlers["pal_consensus"] = HandlePalConsensus
	r.handlers["pal_codereview"] = HandlePalCodeReview
	r.handlers["pal_precommit"] = HandlePalPrecommit
	r.handlers["pal_debug"] = HandlePalDebug
	r.handlers["pal_challenge"] = HandlePalChallenge

	// Short/alias mappings for PAL tools without prefix
	r.handlers["chat"] = HandlePalChat
	r.handlers["thinkdeep"] = HandlePalThinkDeep
	r.handlers["planner"] = HandlePalPlanner
	r.handlers["consensus"] = HandlePalConsensus
	r.handlers["codereview"] = HandlePalCodeReview
	r.handlers["precommit"] = HandlePalPrecommit
	r.handlers["debug"] = HandlePalDebug
	r.handlers["challenge"] = HandlePalChallenge

	// Serena Tools (Category 13)
	r.handlers["get_symbols_overview"] = HandleGetSymbolsOverview
	r.handlers["find_symbol"] = HandleFindSymbol
	r.handlers["find_referencing_symbols"] = HandleFindReferencingSymbols
	r.handlers["find_implementations"] = HandleFindImplementations
	r.handlers["find_declaration"] = HandleFindDeclaration
	r.handlers["rename_symbol"] = HandleRenameSymbol
	r.handlers["onboarding"] = HandleOnboarding

	// Claude Code Aliases
	r.handlers["Read"] = HandleRead
	r.handlers["Write"] = HandleWrite
	r.handlers["Edit"] = HandleEdit
	r.handlers["Bash"] = HandleBash
	r.handlers["LS"] = HandleLS
	r.handlers["WebFetch"] = HandleWebFetch
	r.handlers["Glob"] = HandleGlob
	r.handlers["Grep"] = HandleGrep
	r.handlers["MultiEdit"] = HandleMultiEdit

	// Codex Aliases
	r.handlers["shell"] = HandleBash
	r.handlers["create_file"] = HandleWrite
	r.handlers["view_file"] = HandleRead
	r.handlers["apply_diff"] = HandleApplyPatch
	r.handlers["search_files_codex"] = HandleGrep

	// OpenCode / Pi Aliases
	r.handlers["read"] = HandleRead
	r.handlers["write"] = HandleWrite
	r.handlers["edit"] = HandleEdit
	r.handlers["grep"] = HandleGrep
	r.handlers["ls"] = HandleLS
	r.handlers["glob_pi"] = HandleGlob

	// Thoughtbox Tools (Category 14)
	r.handlers["thoughtbox_search"] = HandleThoughtboxSearch
	r.handlers["thoughtbox_execute"] = HandleThoughtboxExecute
	r.handlers["thoughtbox_peer_notebook"] = HandleThoughtboxPeerNotebook

	// Fetch Tool (Assimilated)
	r.handlers["fetch"] = HandleFetch

	// Tavily Tools (Assimilated)
	r.handlers["tavily-search"] = HandleTavilySearch

	// Chrome DevTools Tools (Assimilated)
	r.handlers["chrome-devtools"] = HandleChromeDevTools

	// Firecrawl Tools (Assimilated from firecrawl-mcp)
	r.handlers["firecrawl_scrape"] = HandleFirecrawl
	r.handlers["firecrawl_crawl"] = HandleFirecrawl
	r.handlers["firecrawl"] = HandleFirecrawl

	// Exa Search Tools (Assimilated from SSE exa)
	r.handlers["exa_search"] = HandleExaSearch
	r.handlers["exa_find_similar"] = HandleExaFindSimilar
	r.handlers["exa_get_contents"] = HandleExaGetContents

	// arXiv Tools (Assimilated from arxiv-mcp-server)
	r.handlers["arxiv_search"] = HandleArxivSearch
	r.handlers["arxiv_get_paper"] = HandleArxivGetPaper
	r.handlers["arxiv_list_recent"] = HandleArxivListRecent

	// Semantic Scholar Tools (Assimilated from paper_search_server)
	r.handlers["paper_search"] = HandleSemanticScholarSearch
	r.handlers["paper_details"] = HandleSemanticScholarGetPaper
	r.handlers["paper_citations"] = HandleSemanticScholarGetCitations
	r.handlers["semantic_scholar_search"] = HandleSemanticScholarSearch
	r.handlers["semantic_scholar_paper"] = HandleSemanticScholarGetPaper

	// mem0 Memory Tools (Assimilated from @mem0/mcp-server)
	r.handlers["mem0_add_memory"] = HandleMem0AddMemory
	r.handlers["mem0_search_memory"] = HandleMem0SearchMemory
	r.handlers["mem0_get_memories"] = HandleMem0GetMemories
	r.handlers["mem0_delete_memory"] = HandleMem0DeleteMemory
	r.handlers["mem0_update_memory"] = HandleMem0UpdateMemory
	r.handlers["add_memory"] = HandleMem0AddMemory
	r.handlers["search_memory"] = HandleMem0SearchMemory

	// Alpaca Trading Tools (Assimilated from alpaca-mcp-server)
	r.handlers["alpaca_get_account"] = HandleAlpacaGetAccount
	r.handlers["alpaca_get_positions"] = HandleAlpacaGetPositions
	r.handlers["alpaca_get_orders"] = HandleAlpacaGetOrders
	r.handlers["alpaca_place_order"] = HandleAlpacaPlaceOrder
	r.handlers["alpaca_cancel_order"] = HandleAlpacaCancelOrder
	r.handlers["alpaca_get_bars"] = HandleAlpacaGetBars
	r.handlers["alpaca_get_latest_quote"] = HandleAlpacaGetLatestQuote

	// Alpha Vantage Financial Tools (Assimilated from av-mcp)
	r.handlers["av_quote"] = HandleAVGlobalQuote
	r.handlers["av_time_series"] = HandleAVTimeSeries
	r.handlers["av_forex_rate"] = HandleAVForexRate
	r.handlers["av_crypto_rate"] = HandleAVCryptoRate
	r.handlers["av_symbol_search"] = HandleAVSearch
	r.handlers["av_economic_indicator"] = HandleAVEconomicIndicator
	r.handlers["alpha_vantage_quote"] = HandleAVGlobalQuote

	// Hugging Face Hub Tools (Assimilated from SSE huggingface)
	r.handlers["hf_search_models"] = HandleHFSearchModels
	r.handlers["hf_get_model"] = HandleHFGetModel
	r.handlers["hf_search_datasets"] = HandleHFSearchDatasets
	r.handlers["hf_text_generation"] = HandleHFTextGeneration
	r.handlers["hf_classify_text"] = HandleHFClassification
	r.handlers["hf_embeddings"] = HandleHFEmbeddings
	r.handlers["hf_search_spaces"] = HandleHFSearchSpaces

	// Semgrep Security Tools (Assimilated from semgrep + semgrepstream)
	r.handlers["semgrep_scan"] = HandleSemgrepScan
	r.handlers["semgrep_cloud_scan"] = HandleSemgrepCloudScan
	r.handlers["semgrep_search_rules"] = HandleSemgrepRuleSearch

	// Octagon Financial Intelligence (Assimilated from octagon + octagon-deep-research)
	r.handlers["octagon_research"] = HandleOctagonResearch
	r.handlers["octagon_company_search"] = HandleOctagonCompanySearch
	r.handlers["octagon_financials"] = HandleOctagonFinancials
	r.handlers["octagon_news"] = HandleOctagonNews

	// Browser Automation Tools (Assimilated from playwright/browser-use/browsermcp/puppeteer/browserbase)
	r.handlers["browser_navigate"] = HandleBrowserNavigate
	r.handlers["browser_screenshot"] = HandleBrowserScreenshot
	r.handlers["browser_get_html"] = HandleBrowserGetHTML
	r.handlers["browser_evaluate"] = HandleBrowserEvaluate
	r.handlers["browser_click"] = HandleBrowserClick
	r.handlers["browser_fill_form"] = HandleBrowserFillForm

	// ChromaDB Vector Store Tools (Assimilated from chroma-mcp)
	r.handlers["chroma_list_collections"] = HandleChromaListCollections
	r.handlers["chroma_create_collection"] = HandleChromaCreateCollection
	r.handlers["chroma_add_documents"] = HandleChromaAddDocuments
	r.handlers["chroma_query"] = HandleChromaQuery
	r.handlers["chroma_delete_collection"] = HandleChromaDeleteCollection
	r.handlers["chroma_get_documents"] = HandleChromaGetCollection

	// Basic Memory Tools (Assimilated from basic-memory)
	r.handlers["basic_memory_write"] = HandleBasicMemoryWrite
	r.handlers["basic_memory_read"] = HandleBasicMemoryRead
	r.handlers["basic_memory_search"] = HandleBasicMemorySearch
	r.handlers["basic_memory_list"] = HandleBasicMemoryList
	r.handlers["basic_memory_delete"] = HandleBasicMemoryDelete
	r.handlers["memory_write"] = HandleBasicMemoryWrite
	r.handlers["memory_read"] = HandleBasicMemoryRead
	r.handlers["memory_search"] = HandleBasicMemorySearch

	// MindsDB ML Database Tools (Assimilated from SSE mindsdb)
	r.handlers["mindsdb_query"] = HandleMindsDBQuery
	r.handlers["mindsdb_list_models"] = HandleMindsDBListModels
	r.handlers["mindsdb_predict"] = HandleMindsDBPredict

	// ═══════════════════════════════════════════════════════════════
	// ASSIMILATED MCP SERVERS — Phase 2: Full Native Reimplementation
	// ═══════════════════════════════════════════════════════════════

	// GitHub Copilot API Tools (Assimilated from github SSE)
	r.handlers["github_list_repos"] = HandleGithubListRepos
	r.handlers["github_get_repo"] = HandleGithubGetRepo
	r.handlers["github_create_issue"] = HandleGithubCreateIssue
	r.handlers["github_list_issues"] = HandleGithubListIssues
	r.handlers["github_create_pr"] = HandleGithubCreatePR
	r.handlers["github_code_search"] = HandleGithubCodeSearch
	r.handlers["github_get_file_contents"] = HandleGithubGetFileContents
	r.handlers["github_create_or_update_file"] = HandleGithubCreateOrUpdateFile
	r.handlers["github_list_branches"] = HandleGithubListBranches
	r.handlers["github_list_workflows"] = HandleGithubListWorkflows
	r.handlers["github_trigger_workflow"] = HandleGithubTriggerWorkflow
	r.handlers["github_copilot_chat"] = HandleGithubCopilotChat

	// Supabase Tools (Assimilated from supabase SSE)
	r.handlers["supabase_list_projects"] = HandleSupabaseListProjects
	r.handlers["supabase_get_project"] = HandleSupabaseGetProject
	r.handlers["supabase_execute_sql"] = HandleSupabaseExecuteSQL
	r.handlers["supabase_select_rows"] = HandleSupabaseSelectRows
	r.handlers["supabase_insert_rows"] = HandleSupabaseInsertRows
	r.handlers["supabase_update_rows"] = HandleSupabaseUpdateRows
	r.handlers["supabase_delete_rows"] = HandleSupabaseDeleteRows
	r.handlers["supabase_list_tables"] = HandleSupabaseListTables
	r.handlers["supabase_invoke_function"] = HandleSupabaseInvokeFunction

	// Desktop Commander Tools (Assimilated from @wonderwhy-er/desktop-commander)
	r.handlers["desktop_execute_command"] = HandleDesktopExecuteCommand
	r.handlers["desktop_read_file"] = HandleDesktopReadFile
	r.handlers["desktop_read_multiple_files"] = HandleDesktopReadMultipleFiles
	r.handlers["desktop_write_file"] = HandleDesktopWriteFile
	r.handlers["desktop_create_directory"] = HandleDesktopCreateDirectory
	r.handlers["desktop_list_directory"] = HandleDesktopListDirectory
	r.handlers["desktop_directory_tree"] = HandleDesktopDirectoryTree
	r.handlers["desktop_search_files"] = HandleDesktopSearchFiles
	r.handlers["desktop_move_file"] = HandleDesktopMoveFile
	r.handlers["desktop_get_file_info"] = HandleDesktopGetFileInfo
	r.handlers["desktop_list_processes"] = HandleDesktopListProcesses
	r.handlers["desktop_kill_process"] = HandleDesktopKillProcess
	r.handlers["desktop_get_system_info"] = HandleDesktopGetSystemInfo
	r.handlers["desktop_execute_script"] = HandleDesktopExecuteScript
	r.handlers["desktop_open_file"] = HandleDesktopOpenFile
	r.handlers["desktop_tail_file"] = HandleDesktopTailFile

	// Gemini API Tools (Assimilated from gemini-mcp)
	r.handlers["gemini_chat"] = HandleGeminiChat
	r.handlers["gemini_code_generation"] = HandleGeminiCodeGeneration
	r.handlers["gemini_vision"] = HandleGeminiVision
	r.handlers["gemini_embeddings"] = HandleGeminiEmbeddings
	r.handlers["gemini_list_models"] = HandleGeminiListModels
	r.handlers["gemini_function_calling"] = HandleGeminiFunctionCalling

	// DBHub Universal Database Tools (Assimilated from @bytebase/dbhub)
	r.handlers["dbhub_list_databases"] = HandleDBHubListDatabases
	r.handlers["dbhub_list_tables"] = HandleDBHubListTables
	r.handlers["dbhub_describe_table"] = HandleDBHubDescribeTable
	r.handlers["dbhub_execute_query"] = HandleDBHubExecuteQuery
	r.handlers["dbhub_list_schemas"] = HandleDBHubListSchemas

	// ConPort Context Portal Tools (Assimilated from context-portal-mcp)
	r.handlers["conport_get_context"] = HandleConPortGetContext
	r.handlers["conport_update_context"] = HandleConPortUpdateContext
	r.handlers["conport_log_decision"] = HandleConPortLogDecision
	r.handlers["conport_get_decisions"] = HandleConPortGetDecisions
	r.handlers["conport_add_pattern"] = HandleConPortAddPattern
	r.handlers["conport_get_patterns"] = HandleConPortGetPatterns
	r.handlers["conport_set_active_context"] = HandleConPortSetActiveContext
	r.handlers["conport_get_active_context"] = HandleConPortGetActiveContext
	r.handlers["conport_log_progress"] = HandleConPortLogProgress
	r.handlers["conport_get_progress"] = HandleConPortGetProgress

	// ChunkHound Code Search Tools (Assimilated from chunkhound)
	r.handlers["chunkhound_index"] = HandleChunkhoundIndex
	r.handlers["chunkhound_search"] = HandleChunkhoundSearch
	r.handlers["chunkhound_stats"] = HandleChunkhoundStats
	r.handlers["chunkhound_list_indexed"] = HandleChunkhoundListIndexed
	r.handlers["chunkhound_get_chunk"] = HandleChunkhoundGetChunk

	// NotebookLM Tools (Assimilated from @roomi-fields/notebooklm-mcp)
	r.handlers["notebooklm_create_notebook"] = HandleNotebookLMCreateNotebook
	r.handlers["notebooklm_query_notebook"] = HandleNotebookLMQueryNotebook
	r.handlers["notebooklm_list_notebooks"] = HandleNotebookLMListNotebooks
	r.handlers["notebooklm_add_source"] = HandleNotebookLMAddSource
	r.handlers["notebooklm_get_summary"] = HandleNotebookLMGetSummary
	r.handlers["notebooklm_upload_pdf"] = HandleNotebookLMUploadPDF

	// Vibe Check Tools (Assimilated from @pv-bhat/vibe-check-mcp)
	r.handlers["vibe_check_analyze"] = HandleVibeCheckAnalyze
	r.handlers["vibe_check_quick"] = HandleVibeCheckQuick
	r.handlers["vibe_check_review_patterns"] = HandleVibeCheckReviewPatterns

	// SuperMemory Tools (Assimilated from mcp-supermemory-ai)
	r.handlers["supermemory_add"] = HandleSuperMemoryAdd
	r.handlers["supermemory_search"] = HandleSuperMemorySearch
	r.handlers["supermemory_delete"] = HandleSuperMemoryDelete
	r.handlers["supermemory_list"] = HandleSuperMemoryList

	// Probe Code Search Tools (Assimilated from @probelabs/probe)
	r.handlers["probe_search_code"] = HandleProbeSearchCode
	r.handlers["probe_find_symbol"] = HandleProbeFindSymbol
	r.handlers["probe_get_structure"] = HandleProbeGetStructure
	r.handlers["probe_explain_code"] = HandleProbeExplainCode

	// Cipher Memory Aggregator Tools (Assimilated from @byterover/cipher)
	r.handlers["cipher_add_memory"] = HandleCipherAddMemory
	r.handlers["cipher_search_memory"] = HandleCipherSearchMemory
	r.handlers["cipher_list_memories"] = HandleCipherListMemories
	r.handlers["cipher_delete_memory"] = HandleCipherDeleteMemory
	r.handlers["cipher_ask"] = HandleCipherAskCipher

	// DeepContext Code Understanding Tools (Assimilated from @wildcard-ai/deepcontext)
	r.handlers["deepcontext_analyze"] = HandleDeepContextAnalyzeCodebase
	r.handlers["deepcontext_get_context"] = HandleDeepContextGetContext
	r.handlers["deepcontext_find_patterns"] = HandleDeepContextFindPatterns
	r.handlers["deepcontext_summarize_architecture"] = HandleDeepContextSummarizeArchitecture

	// Windows MCP Tools (Assimilated from windows-mcp)
	r.handlers["windows_get_system_info"] = HandleWindowsMCPGetSystemInfo
	r.handlers["windows_list_services"] = HandleWindowsMCPListServices
	r.handlers["windows_get_service"] = HandleWindowsMCPGetService
	r.handlers["windows_list_processes"] = HandleWindowsMCPListProcesses
	r.handlers["windows_read_registry"] = HandleWindowsMCPReadRegistry
	r.handlers["windows_open_application"] = HandleWindowsMCPOpenApplication
	r.handlers["windows_get_clipboard"] = HandleWindowsMCPGetClipboard
	r.handlers["windows_set_clipboard"] = HandleWindowsMCPSetClipboard
	r.handlers["windows_list_drives"] = HandleWindowsMCPListDrives
	r.handlers["windows_get_event_log"] = HandleWindowsMCPGetEventLog

	// Prism Code Quality Tools (Assimilated from prism-mcp-server)
	r.handlers["prism_analyze_quality"] = HandlePrismAnalyzeQuality
	r.handlers["prism_suggest_refactor"] = HandlePrismSuggestRefactor
	r.handlers["prism_detect_smells"] = HandlePrismDetectSmells
	r.handlers["prism_transform_code"] = HandlePrismTransformCode

	// TaskMaster AI Task Management Tools (Assimilated from task-master-ai)
	r.handlers["taskmaster_create_task"] = HandleTaskMasterCreateTask
	r.handlers["taskmaster_get_task"] = HandleTaskMasterGetTask
	r.handlers["taskmaster_list_tasks"] = HandleTaskMasterListTasks
	r.handlers["taskmaster_update_status"] = HandleTaskMasterUpdateStatus
	r.handlers["taskmaster_add_subtask"] = HandleTaskMasterAddSubtask
	r.handlers["taskmaster_next_task"] = HandleTaskMasterNextTask
	r.handlers["taskmaster_generate_from_prd"] = HandleTaskMasterGenerateFromPRD
	r.handlers["taskmaster_expand_task"] = HandleTaskMasterExpandTask

	// ═══════════════════════════════════════════════════════════════
	// SKILL REGISTRY - Database-backed skill management with deduplication
	// ═══════════════════════════════════════════════════════════════

	// Skill Registry Tools
	r.handlers["skill_list"] = HandleSkillList
	r.handlers["skill_get"] = HandleSkillGet
	r.handlers["skill_store"] = HandleSkillStore
	r.handlers["skill_search"] = HandleSkillSearch
	r.handlers["skills_list"] = HandleSkillList
	r.handlers["skills_get"] = HandleSkillGet
	r.handlers["skills_store"] = HandleSkillStore
	r.handlers["skills_search"] = HandleSkillSearch

	// OpenMemory — local persistent memory store
	r.handlers["openmemory_add"] = HandleOpenMemoryAdd
	r.handlers["openmemory_search"] = HandleOpenMemorySearch
	r.handlers["openmemory_get"] = HandleOpenMemoryGet
	r.handlers["openmemory_delete"] = HandleOpenMemoryDelete
	r.handlers["openmemory_list"] = HandleOpenMemoryList

	// AutoMem — graph-vector memory for AI agents
	r.handlers["automem_add"] = HandleAutoMemAdd
	r.handlers["automem_search"] = HandleAutoMemSearch
	r.handlers["automem_get"] = HandleAutoMemGet
	r.handlers["automem_delete"] = HandleAutoMemDelete
	r.handlers["automem_list"] = HandleAutoMemList
	r.handlers["automem_associate"] = HandleAutoMemAssociate

	// lsmcp — LSP code manipulation and analysis
	r.handlers["project_overview"] = HandleLsmcpProjectOverview
	r.handlers["search_symbols"] = HandleLsmcpSearchSymbols
	r.handlers["get_diagnostics"] = HandleLsmcpGetDiagnostics
	r.handlers["find_references"] = HandleLsmcpFindReferences
	r.handlers["get_symbol_details"] = HandleLsmcpGetSymbolDetails

	// CodeAlive — semantic code search and context engine
	r.handlers["codealive_search"] = HandleCodeAliveSearch
	r.handlers["codealive_grep"] = HandleCodeAliveGrep
	r.handlers["codealive_ask"] = HandleCodeAliveAsk

	// Prometheus MCP — monitoring queries
	r.handlers["prom_query"] = HandlePromQuery
	r.handlers["prom_alerts"] = HandlePromAlerts
	r.handlers["prom_targets"] = HandlePromTargets
	r.handlers["prom_metadata"] = HandlePromMetadata

	// Smart-Thinking — graph-based reasoning
	r.handlers["smart_reason"] = HandleSmartReason
	r.handlers["smart_session"] = HandleSmartSession
	r.handlers["smart_evaluate"] = HandleSmartEvaluate
	r.handlers["smart_graph"] = HandleSmartGraph

	// Mimir — Neo4j-backed persistent memory
	r.handlers["mimir_store"] = HandleMimirStore
	r.handlers["mimir_search"] = HandleMimirSearch
	r.handlers["mimir_retrieve"] = HandleMimirRetrieve
	r.handlers["mimir_connect"] = HandleMimirConnect
	r.handlers["mimir_forget"] = HandleMimirForget

	// Sysmon — system monitoring
	r.handlers["sysmon_overview"] = HandleSysmonOverview
	r.handlers["sysmon_health"] = HandleSysmonHealth
	r.handlers["sysmon_top"] = HandleSysmonTop
	r.handlers["sysmon_disk"] = HandleSysmonDisk
	r.handlers["sysmon_network"] = HandleSysmonNetwork
	r.handlers["sysmon_find"] = HandleSysmonFind

	// Docker — container management
	r.handlers["docker_list_containers"] = HandleDockerListContainers
	r.handlers["docker_list_images"] = HandleDockerListImages
	r.handlers["docker_inspect"] = HandleDockerInspect
	r.handlers["docker_logs"] = HandleDockerLogs
	r.handlers["docker_stats"] = HandleDockerStats
	r.handlers["docker_exec"] = HandleDockerExec

	// Social — Twitter/X and Reddit
	r.handlers["twitter_search"] = HandleTwitterSearch
	r.handlers["twitter_user_timeline"] = HandleTwitterUserTimeline
	r.handlers["reddit_search"] = HandleRedditSearch
	r.handlers["reddit_get_posts"] = HandleRedditGetPosts

	// Git — repository operations
	r.handlers["git_status"] = HandleGitStatus
	r.handlers["git_log"] = HandleGitLog
	r.handlers["git_diff"] = HandleGitDiff
	r.handlers["git_branches"] = HandleGitBranches
	r.handlers["git_show"] = HandleGitShow
	r.handlers["git_blame"] = HandleGitBlame
	r.handlers["git_commit"] = HandleGitCommit
	r.handlers["git_checkout"] = HandleGitCheckout

	// Terraform — infrastructure management
	r.handlers["terraform_search_providers"] = HandleTerraformSearchProviders
	r.handlers["terraform_search_modules"] = HandleTerraformSearchModules
	r.handlers["terraform_get_provider"] = HandleTerraformGetProvider

	// Google News — news headlines and search
	r.handlers["google_news_headlines"] = HandleGoogleNewsHeadlines
	r.handlers["google_news_search"] = HandleGoogleNewsSearch

	// OpenRouter Deep Research
	r.handlers["deep_research"] = HandleDeepResearch
	r.handlers["deep_research_status"] = HandleDeepResearchStatus

	// Prompt Library — SQLite-backed prompt storage
	r.handlers["prompt_list"] = HandlePromptList
	r.handlers["prompt_get"] = HandlePromptGet
	r.handlers["prompt_search"] = HandlePromptSearch

	// Context Server — SQLite-backed context management
	r.handlers["context_store"] = HandleContextStore
	r.handlers["context_search"] = HandleContextSearch
	r.handlers["context_get"] = HandleContextGet
	r.handlers["context_delete"] = HandleContextDelete
	r.handlers["context_list_threads"] = HandleContextListThreads
	r.handlers["context_stats"] = HandleContextStats

	// WebPeel — web data extraction
	r.handlers["webpeel_fetch"] = HandleWebpeelFetch
	r.handlers["webpeel_search"] = HandleWebpeelSearch
	r.handlers["webpeel_extract"] = HandleWebpeelExtract

	// Omnisearch — universal search
	r.handlers["omnisearch_github"] = HandleOmnisearchGithub
	r.handlers["omnisearch_stackoverflow"] = HandleOmnisearchStackoverflow
	r.handlers["omnisearch_npm"] = HandleOmnisearchNpm
	r.handlers["omnisearch_pypi"] = HandleOmnisearchPypi
	r.handlers["omnisearch_web"] = HandleOmnisearchWeb

	// Grants — government grants discovery
	r.handlers["grants_search"] = HandleGrantsSearch
	r.handlers["grants_by_agency"] = HandleGrantsByAgency
	r.handlers["grants_by_category"] = HandleGrantsByCategory
	r.handlers["grants_trends"] = HandleGrantsTrends

	// Food Data Central — USDA nutrition database
	r.handlers["food_search"] = HandleFoodSearch
	r.handlers["food_get"] = HandleFoodGet
	r.handlers["food_list"] = HandleFoodList

	// Panther — security monitoring
	r.handlers["panther_query"] = HandlePantherQuery
	r.handlers["panther_list_detections"] = HandlePantherListDetections
	r.handlers["panther_get_findings"] = HandlePantherGetFindings
	r.handlers["panther_list_policies"] = HandlePantherListPolicies

	// Srclight — code indexing for AI agents
	r.handlers["srclight_index"] = HandleSrclightIndex
	r.handlers["srclight_search"] = HandleSrclightSearch
	r.handlers["srclight_status"] = HandleSrclightStatus
	r.handlers["srclight_list_languages"] = HandleSrclightListLanguages

	// Coolify — deployment & infrastructure management
	r.handlers["coolify_list_projects"] = HandleCoolifyListProjects
	r.handlers["coolify_create_project"] = HandleCoolifyCreateProject
	r.handlers["coolify_list_services"] = HandleCoolifyListServices
	r.handlers["coolify_deploy_service"] = HandleCoolifyDeployService
	r.handlers["coolify_get_logs"] = HandleCoolifyGetLogs
	r.handlers["coolify_list_databases"] = HandleCoolifyListDatabases

	// Harness Integrations
	r.handlers["launch_tabby"] = HandleTabby
	r.handlers["launch_warp"] = HandleWarp
	r.handlers["hermes_agent"] = HandleHermesAgent
	r.handlers["pi_mono"] = HandlePiMono

	// Bobbybookmarks Integration
	r.handlers["bobbybookmarks_sync"] = HandleBobbyBookmarksSync
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (ToolResponse, error) {
	r.mu.RLock()
	handler, ok := r.handlers[name]
	r.mu.RUnlock()
	if !ok {
		return ToolResponse{}, fmt.Errorf("tool handler not found for: %s", name)
	}
	return handler(ctx, args)
}

func (r *Registry) HasTool(name string) bool {
	r.mu.RLock()
	_, ok := r.handlers[name]
	r.mu.RUnlock()
	return ok
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		result = append(result, name)
	}
	return result
}
