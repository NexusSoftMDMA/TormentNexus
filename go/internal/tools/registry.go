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
)

type ToolHandler func(ctx context.Context, args map[string]interface{}) (ToolResponse, error)

type Registry struct {
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
}

func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (ToolResponse, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return ToolResponse{}, fmt.Errorf("tool handler not found for: %s", name)
	}
	return handler(ctx, args)
}

func (r *Registry) HasTool(name string) bool {
	_, ok := r.handlers[name]
	return ok
}
