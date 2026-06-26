import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
    CallToolRequestSchema,
    ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { Installer } from './installer.js';
import { ProcessManager } from './process_manager.js';
import { InputManager } from './input_manager.js';
import { UiAutomationManager } from './ui_automation.js';

import { logger } from './logger.js';

class SupervisorServer {
    private server: Server;
    private processManager: ProcessManager;
    private inputManager: InputManager;
    private uiAutomationManager: UiAutomationManager;

    constructor() {
        this.processManager = new ProcessManager();
        this.inputManager = new InputManager();
        this.uiAutomationManager = new UiAutomationManager();
        this.server = new Server(
            {
                name: "tormentnexus-supervisor",
                version: "0.1.0",
            },
            {
                capabilities: {
                    tools: {},
                },
            }
        );

        this.setupHandlers();
    }

    private setupHandlers() {
        this.server.setRequestHandler(ListToolsRequestSchema, async () => {
            return {
                tools: [
                    {
                        name: "install_supervisor",
                        description: "Install TormentNexus Supervisor into Antigravity MCP Config",
                        inputSchema: {
                            type: "object",
                            properties: {
                                configPath: {
                                    type: "string",
                                    description: "Abs path to mcp.json"
                                }
                            }
                        }
                    },
                    {
                        name: "list_processes",
                        description: "List active system processes",
                        inputSchema: { type: "object", properties: {} }
                    },
                    {
                        name: "kill_process",
                        description: "Kill a process by PID",
                        inputSchema: {
                            type: "object",
                            properties: {
                                pid: { type: "number", description: "Process ID" }
                            },
                            required: ["pid"]
                        }
                    },
                    {
                        name: "simulate_input",
                        description: "Send keyboard input (PowerShell SendKeys)",
                        inputSchema: {
                            type: "object",
                            properties: {
                                keys: {
                                    type: "string",
                                    description: "Keys to send (e.g. 'ctrl+r', 'f5')"
                                },
                                windowTitle: {
                                    type: "string",
                                    description: "Exact window title to focus before sending keys (Recommended)"
                                }
                            },
                            required: ["keys"]
                        }
                    },
                    {
                        name: "detect_chat_surface",
                        description: "Inspect the active or matching window and classify the current chat surface heuristically",
                        inputSchema: {
                            type: "object",
                            properties: {
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target (e.g. chrome, firefox)"
                                },
                                surfaceOverride: {
                                    type: "string",
                                    description: "Optional explicit surface/profile id to force instead of heuristic detection"
                                }
                            }
                        }
                    },
                    {
                        name: "list_surface_profiles",
                        description: "List the known supervisor surface profiles and their default labels, submit chords, and input preferences",
                        inputSchema: {
                            type: "object",
                            properties: {}
                        }
                    },
                    {
                        name: "inspect_window_ui",
                        description: "List visible button-like controls and text inputs from the active or matching window",
                        inputSchema: {
                            type: "object",
                            properties: {
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target (e.g. chrome, firefox)"
                                }
                            }
                        }
                    },
                    {
                        name: "detect_chat_state",
                        description: "Heuristically detect whether the current chat is waiting on action buttons or ready for bump text",
                        inputSchema: {
                            type: "object",
                            properties: {
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target"
                                },
                                surfaceOverride: {
                                    type: "string",
                                    description: "Optional explicit surface/profile id to force"
                                }
                            }
                        }
                    },
                    {
                        name: "click_action_buttons",
                        description: "Find real button-like UI elements by label and click them without treating comboboxes as buttons",
                        inputSchema: {
                            type: "object",
                            properties: {
                                labels: {
                                    type: "array",
                                    items: { type: "string" },
                                    description: "Labels to click. Defaults to Run/Expand/Allow/Accept style actions."
                                },
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target"
                                },
                                surfaceOverride: {
                                    type: "string",
                                    description: "Optional explicit surface/profile id to force"
                                }
                            }
                        }
                    },
                    {
                        name: "set_chat_input",
                        description: "Find the active chat composer, replace its content, and type bump text",
                        inputSchema: {
                            type: "object",
                            properties: {
                                text: {
                                    type: "string",
                                    description: "Text to place in the detected chat input"
                                },
                                clearExisting: {
                                    type: "boolean",
                                    description: "Whether to replace existing composer text",
                                    default: true
                                },
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target"
                                },
                                surfaceOverride: {
                                    type: "string",
                                    description: "Optional explicit surface/profile id to force"
                                }
                            },
                            required: ["text"]
                        }
                    },
                    {
                        name: "submit_chat_input",
                        description: "Submit the current chat composer with a configurable key chord such as alt+enter",
                        inputSchema: {
                            type: "object",
                            properties: {
                                keyChord: {
                                    type: "string",
                                    description: "Submission key chord",
                                    default: "alt+enter"
                                },
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target"
                                },
                                surfaceOverride: {
                                    type: "string",
                                    description: "Optional explicit surface/profile id to force"
                                }
                            }
                        }
                    },
                    {
                        name: "advance_chat",
                        description: "Single-step autopilot helper: click pending action buttons, otherwise type and submit bump text",
                        inputSchema: {
                            type: "object",
                            properties: {
                                bumpText: {
                                    type: "string",
                                    description: "Optional bump text to type when the chat is ready for input"
                                },
                                actionLabels: {
                                    type: "array",
                                    items: { type: "string" },
                                    description: "Optional button labels to click"
                                },
                                windowTitle: {
                                    type: "string",
                                    description: "Optional partial window title to target"
                                },
                                processName: {
                                    type: "string",
                                    description: "Optional process name to target"
                                },
                                surfaceOverride: {
                                    type: "string",
                                    description: "Optional explicit surface/profile id to force"
                                }
                            }
                        }
                    },
                    {
                        name: "get_supervisor_settings",
                        description: "Read the persisted supervisor defaults for bump text, action labels, and timing",
                        inputSchema: {
                            type: "object",
                            properties: {}
                        }
                    },
                    {
                        name: "update_supervisor_settings",
                        description: "Persist supervisor defaults for bump text, action labels, and timing",
                        inputSchema: {
                            type: "object",
                            properties: {
                                bumpText: {
                                    type: "string",
                                    description: "Default bump text used by advance_chat"
                                },
                                actionLabels: {
                                    type: "array",
                                    items: { type: "string" },
                                    description: "Default action labels to match exactly"
                                },
                                focusDelayMs: {
                                    type: "number",
                                    description: "Focus settle delay before submission"
                                },
                                afterClickDelayMs: {
                                    type: "number",
                                    description: "Delay after clicking an action button"
                                },
                                inputSettleDelayMs: {
                                    type: "number",
                                    description: "Delay after focusing an input before typing"
                                }
                            }
                        }
                    }
                ]
            };
        });

        this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
            logger.info(`Executing tool: ${request.params.name}`, request.params.arguments);

            try {
                if (request.params.name === "install_supervisor") {
                    const configPath = request.params.arguments?.configPath as string | undefined;
                    const installer = new Installer(configPath);
                    const result = await installer.install();
                    logger.info("Install Result", { result });
                    return {
                        content: [{ type: "text", text: result }]
                    };
                }

                if (request.params.name === "list_processes") {
                    const processes = await this.processManager.listProcesses();
                    return {
                        content: [{ type: "text", text: JSON.stringify(processes, null, 2) }]
                    };
                }

                if (request.params.name === "kill_process") {
                    const pid = request.params.arguments?.pid as number;
                    const result = await this.processManager.killProcess(pid);
                    logger.warn("Process Killed", { pid, result });
                    return {
                        content: [{ type: "text", text: result }]
                    };
                }

                if (request.params.name === "simulate_input") {
                    const keys = request.params.arguments?.keys as string;
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const result = await this.inputManager.sendKeys(keys, windowTitle);
                    logger.info("Input Simulated", { keys, result });
                    return {
                        content: [{ type: "text", text: result }]
                    };
                }

                if (request.params.name === "detect_chat_surface") {
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const surfaceOverride = request.params.arguments?.surfaceOverride as string | undefined;
                    const result = await this.uiAutomationManager.detectChatSurface({ windowTitle, processName, surfaceOverride });
                    logger.info("Chat Surface Detected", { windowTitle, processName, detectedSurface: result.detectedSurface });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "list_surface_profiles") {
                    const result = this.uiAutomationManager.listSurfaceProfiles();
                    logger.info("Surface Profiles Listed", { count: result.length });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "inspect_window_ui") {
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const result = await this.uiAutomationManager.inspectWindow(windowTitle, processName);
                    logger.info("Window UI Inspected", { windowTitle, processName });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "detect_chat_state") {
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const surfaceOverride = request.params.arguments?.surfaceOverride as string | undefined;
                    const result = await this.uiAutomationManager.detectChatState(windowTitle, processName, undefined, { surfaceOverride });
                    logger.info("Chat State Detected", { windowTitle, processName, state: result.state });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "click_action_buttons") {
                    const labels = (request.params.arguments?.labels as string[] | undefined) ?? undefined;
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const surfaceOverride = request.params.arguments?.surfaceOverride as string | undefined;
                    const result = await this.uiAutomationManager.clickActionButtons(labels, windowTitle, processName, { surfaceOverride });
                    logger.info("Action Buttons Clicked", { labels, windowTitle, processName, clicked: result.clicked.map((item) => item.name) });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "set_chat_input") {
                    const text = request.params.arguments?.text as string;
                    const clearExisting = request.params.arguments?.clearExisting as boolean | undefined;
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const surfaceOverride = request.params.arguments?.surfaceOverride as string | undefined;
                    const result = await this.uiAutomationManager.setChatInput(text, { clearExisting, windowTitle, processName, surfaceOverride });
                    logger.info("Chat Input Set", { textLength: text.length, clearExisting, windowTitle, processName, method: result.method });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "submit_chat_input") {
                    const keyChord = request.params.arguments?.keyChord as string | undefined;
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const surfaceOverride = request.params.arguments?.surfaceOverride as string | undefined;
                    const result = await this.uiAutomationManager.submitChatInput(keyChord, windowTitle, processName, { surfaceOverride });
                    logger.info("Chat Input Submitted", { keyChord, windowTitle, processName });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "advance_chat") {
                    const bumpText = request.params.arguments?.bumpText as string | undefined;
                    const actionLabels = request.params.arguments?.actionLabels as string[] | undefined;
                    const windowTitle = request.params.arguments?.windowTitle as string | undefined;
                    const processName = request.params.arguments?.processName as string | undefined;
                    const surfaceOverride = request.params.arguments?.surfaceOverride as string | undefined;
                    const result = await this.uiAutomationManager.advanceChat({
                        bumpText,
                        actionLabels,
                        windowTitle,
                        processName,
                        surfaceOverride
                    });
                    logger.info("Advance Chat Completed", { detail: result.detail });
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "get_supervisor_settings") {
                    const result = await this.uiAutomationManager.getSettings();
                    logger.info("Supervisor Settings Read");
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                if (request.params.name === "update_supervisor_settings") {
                    const result = await this.uiAutomationManager.updateSettings({
                        bumpText: request.params.arguments?.bumpText as string | undefined,
                        actionLabels: request.params.arguments?.actionLabels as string[] | undefined,
                        focusDelayMs: request.params.arguments?.focusDelayMs as number | undefined,
                        afterClickDelayMs: request.params.arguments?.afterClickDelayMs as number | undefined,
                        inputSettleDelayMs: request.params.arguments?.inputSettleDelayMs as number | undefined
                    });
                    logger.info("Supervisor Settings Updated", result);
                    return {
                        content: [{ type: "text", text: JSON.stringify(result, null, 2) }]
                    };
                }

                throw new Error(`Tool ${request.params.name} not found`);
            } catch (err: any) {
                logger.error(`Tool Execution Failed: ${request.params.name}`, { error: err.message });
                throw err;
            }
        });
    }

    async start() {
        logger.info("TormentNexus Supervisor Starting...");
        const transport = new StdioServerTransport();
        await this.server.connect(transport);
        logger.info("TormentNexus Supervisor Connected to Stdio");
    }
}

const server = new SupervisorServer();
server.start().catch((err: any) => logger.error("Fatal Error", err));
