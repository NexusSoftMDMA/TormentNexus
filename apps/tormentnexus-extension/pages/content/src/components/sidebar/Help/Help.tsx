import React, { useState, useMemo } from 'react';
import { Typography, Icon, Button } from '../ui';
import { Card, CardContent, CardHeader, CardTitle } from '@src/components/ui/card';
import { cn } from '@src/lib/utils';
import { useProfileStore } from '@src/stores';
import { useActivityStore } from '@src/stores';
import { useToastStore } from '@src/stores';

type Section =
  | 'overview'
  | 'setup'
  | 'features'
  | 'advanced'
  | 'development'
  | 'security'
  | 'diagnostics'
  | 'troubleshooting'
  | 'faq'
  | 'support';

const Help: React.FC = () => {
  const [activeSection, setActiveSection] = useState<Section>('overview');
  const [searchQuery, setSearchQuery] = useState('');

  const renderNavButton = (section: Section, label: string, iconName: any) => (
    <Button
      variant={activeSection === section ? 'secondary' : 'ghost'}
      size="sm"
      className={cn(
        'w-full justify-start mb-1 h-auto py-2 px-3',
        activeSection === section
          ? 'bg-slate-200 dark:bg-slate-700 font-semibold text-slate-900 dark:text-slate-100'
          : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800',
      )}
      onClick={() => setActiveSection(section)}>
      <Icon
        name={iconName}
        size="sm"
        className={cn('mr-2 flex-shrink-0', activeSection === section ? 'text-primary-600 dark:text-primary-400' : '')}
      />
      <span className="truncate">{label}</span>
    </Button>
  );

  const sections: { id: Section; title: string; icon: any; content: React.ReactNode }[] = [
    {
      id: 'overview',
      title: 'Overview',
      icon: 'info',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>What is TormentNexus Extension?</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Typography variant="body" className="text-sm">
                TormentNexus Extension is a Chrome extension that bridges the Model Context Protocol (MCP) with web-based AI
                platforms like ChatGPT, Claude, Perplexity, and others.
              </Typography>
              <Typography variant="body" className="text-sm">
                It allows you to use your local tools and data directly within these AI interfaces, enhancing their
                capabilities with file system access, command execution, and more.
              </Typography>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Key Benefits</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="list-disc pl-5 text-sm space-y-1 text-slate-700 dark:text-slate-300">
                <li>Connect local MCP servers to web AI</li>
                <li>Securely execute tools locally</li>
                <li>Seamlessly insert results into chat</li>
                <li>Support for multiple AI providers</li>
              </ul>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'setup',
      title: 'Setup',
      icon: 'settings',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Connecting a Proxy</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Typography variant="body" className="text-sm">
                To connect to local MCP servers, you need to run the TormentNexus Extension Proxy. This proxy bridges the
                browser (extension) to your local MCP servers.
              </Typography>
              <div className="bg-slate-100 dark:bg-slate-900 p-2 rounded-md text-xs font-mono overflow-x-auto border border-slate-200 dark:border-slate-700">
                npx -y @srbhptl39/tormentnexus-extension-proxy@latest --config ./config.json
              </div>
              <Typography variant="caption" className="block mt-2">
                Create a <code>config.json</code> file defining your MCP servers (e.g., filesystem, postgres) and point
                the proxy to it.
              </Typography>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Connection Types</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="space-y-3">
                <div className="p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="px-1.5 py-0.5 text-[10px] font-bold bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300 rounded uppercase">
                      Default
                    </span>
                    <Typography variant="subtitle" className="font-semibold text-sm">
                      SSE (Server-Sent Events)
                    </Typography>
                  </div>
                  <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400 mb-1">
                    Standard HTTP streaming. Best for stability and most use cases.
                  </Typography>
                  <code className="text-[10px] bg-slate-200 dark:bg-slate-900 px-1 py-0.5 rounded">
                    http://localhost:3006/sse
                  </code>
                </div>

                <div className="p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="px-1.5 py-0.5 text-[10px] font-bold bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300 rounded uppercase">
                      Fast
                    </span>
                    <Typography variant="subtitle" className="font-semibold text-sm">
                      WebSocket
                    </Typography>
                  </div>
                  <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400 mb-1">
                    Full-duplex communication. Lower latency for high-frequency updates.
                  </Typography>
                  <code className="text-[10px] bg-slate-200 dark:bg-slate-900 px-1 py-0.5 rounded">
                    ws://localhost:3006/message
                  </code>
                </div>

                <div className="p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <Typography variant="subtitle" className="font-semibold text-sm mb-1">
                    Streamable HTTP
                  </Typography>
                  <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400 mb-1">
                    Alternative HTTP transport. Use if SSE is blocked by network policies.
                  </Typography>
                  <code className="text-[10px] bg-slate-200 dark:bg-slate-900 px-1 py-0.5 rounded">
                    http://localhost:3006/mcp
                  </code>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'features',
      title: 'Features',
      icon: 'lightning',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Core Features</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <div className="grid grid-cols-1 gap-3">
                <div className="border border-slate-100 dark:border-slate-700 p-3 rounded-lg">
                  <Typography variant="subtitle" className="font-semibold text-sm flex items-center gap-2">
                    <Icon name="search" size="xs" className="text-primary-500" /> Tool Detection
                  </Typography>
                  <Typography variant="body" className="text-xs mt-1 text-slate-600 dark:text-slate-400">
                    The extension automatically detects when the AI wants to call a tool. It presents a "Call Tool" card
                    in the chat.
                  </Typography>
                </div>

                <div className="border border-slate-100 dark:border-slate-700 p-3 rounded-lg">
                  <Typography variant="subtitle" className="font-semibold text-sm flex items-center gap-2">
                    <Icon name="play" size="xs" className="text-green-500" /> Auto-Execute
                  </Typography>
                  <Typography variant="body" className="text-xs mt-1 text-slate-600 dark:text-slate-400">
                    If enabled in Settings, the extension will automatically run tools without requiring you to click
                    "Run". Use with caution.
                  </Typography>
                </div>

                <div className="border border-slate-100 dark:border-slate-700 p-3 rounded-lg">
                  <Typography variant="subtitle" className="font-semibold text-sm flex items-center gap-2">
                    <Icon name="arrow-up-right" size="xs" className="text-blue-500" /> Auto-Submit
                  </Typography>
                  <Typography variant="body" className="text-xs mt-1 text-slate-600 dark:text-slate-400">
                    After a tool runs and the result is pasted into the input box, this feature automatically sends the
                    message to the AI.
                  </Typography>
                </div>

                <div className="border border-slate-100 dark:border-slate-700 p-3 rounded-lg">
                  <Typography variant="subtitle" className="font-semibold text-sm flex items-center gap-2">
                    <Icon name="menu" size="xs" className="text-purple-500" /> Push Content Mode
                  </Typography>
                  <Typography variant="body" className="text-xs mt-1 text-slate-600 dark:text-slate-400">
                    Adjusts the page layout so the sidebar pushes content aside instead of overlaying it. Useful for
                    smaller screens.
                  </Typography>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'advanced',
      title: 'Advanced Config',
      icon: 'tool',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Proxy Configuration</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Typography variant="body" className="text-sm">
                Your <code>config.json</code> defines how the proxy launches MCP servers. You can pass environment
                variables here.
              </Typography>
              <div className="bg-slate-100 dark:bg-slate-900 p-3 rounded-md text-xs font-mono overflow-x-auto border border-slate-200 dark:border-slate-700">
                {`{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/allowed/folder"],
      "env": {
        "DEBUG": "true"
      }
    }
  }
}`}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Environment Variables</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400">
                The proxy respects system environment variables. You can also set them inline in your command:
              </Typography>
              <div className="bg-slate-100 dark:bg-slate-900 p-2 rounded-md text-xs font-mono border border-slate-200 dark:border-slate-700">
                PORT=3007 npx ...
              </div>
              <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400">
                This changes the port the proxy listens on. Remember to update the Extension Settings to match!
              </Typography>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'development',
      title: 'Tool Development',
      icon: 'box',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Creating Custom Tools</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Typography variant="body" className="text-sm">
                You can create your own MCP server to expose custom tools (e.g., internal API access, database queries).
              </Typography>
              <div className="p-3 bg-blue-50 dark:bg-blue-900/10 rounded-lg border border-blue-100 dark:border-blue-900/30 my-2">
                <Typography variant="body" className="text-sm">
                  Refer to the{' '}
                  <a
                    href="https://modelcontextprotocol.io/docs/server"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 dark:text-blue-400 hover:underline font-medium">
                    Official MCP Documentation
                  </a>{' '}
                  to learn how to build an MCP server in Python or TypeScript.
                </Typography>
              </div>
              <Typography variant="body" className="text-sm">
                Once built, simply add it to your <code>config.json</code> and restart the proxy.
              </Typography>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'security',
      title: 'Security',
      icon: 'check',
      content: (
        <div className="space-y-4">
          <Card className="border-l-4 border-l-green-500">
            <CardHeader>
              <CardTitle>Security Best Practices</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Typography variant="subtitle" className="font-semibold text-sm text-green-700 dark:text-green-400">
                  Local Execution
                </Typography>
                <Typography variant="body" className="text-xs">
                  Tools run on your local machine. Be careful with tools that modify files or execute system commands.
                  Always review the tool call before clicking "Run" unless you trust the source completely.
                </Typography>
              </div>
              <div className="border-t border-slate-100 dark:border-slate-800 pt-3">
                <Typography variant="subtitle" className="font-semibold text-sm text-green-700 dark:text-green-400">
                  Data Privacy
                </Typography>
                <Typography variant="body" className="text-xs">
                  Your data (files, database content) remains local. It is only sent to the AI provider (OpenAI,
                  Anthropic, etc.) when a tool result is explicitly inserted into the chat.
                </Typography>
              </div>
              <div className="border-t border-slate-100 dark:border-slate-800 pt-3">
                <Typography variant="subtitle" className="font-semibold text-sm text-green-700 dark:text-green-400">
                  API Keys
                </Typography>
                <Typography variant="body" className="text-xs">
                  Never hardcode API keys in your <code>config.json</code> if you plan to share it. Use environment
                  variables instead.
                </Typography>
              </div>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'diagnostics',
      title: 'Diagnostics',
      icon: 'activity',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>System Diagnostics</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Typography variant="body" className="text-sm">
                Use this information when reporting issues or debugging problems.
              </Typography>

              <div className="space-y-2">
                <div className="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">Extension Version</span>
                  <span className="text-xs font-mono text-slate-800 dark:text-slate-200">0.6.1</span>
                </div>
                <div className="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">User Agent</span>
                  <span
                    className="text-xs font-mono text-slate-800 dark:text-slate-200 truncate max-w-[200px]"
                    title={navigator.userAgent}>
                    {navigator.userAgent}
                  </span>
                </div>
                <div className="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">Activity Logs</span>
                  <span className="text-xs font-mono text-slate-800 dark:text-slate-200">
                    {useActivityStore.getState().logs.length} entries
                  </span>
                </div>
                <div className="flex justify-between items-center p-2 bg-slate-50 dark:bg-slate-800 rounded border border-slate-100 dark:border-slate-700">
                  <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">Profiles</span>
                  <span className="text-xs font-mono text-slate-800 dark:text-slate-200">
                    {useProfileStore.getState().profiles.length} configured
                  </span>
                </div>
              </div>

              <Button
                variant="outline"
                className="w-full mt-2"
                onClick={() => {
                  const info = {
                    version: '0.6.1',
                    userAgent: navigator.userAgent,
                    timestamp: new Date().toISOString(),
                    profiles: useProfileStore.getState().profiles.map(p => ({ name: p.name, type: p.connectionType })),
                    logsSummary: {
                      total: useActivityStore.getState().logs.length,
                      errors: useActivityStore.getState().logs.filter(l => l.status === 'error').length,
                    },
                  };
                  navigator.clipboard.writeText(JSON.stringify(info, null, 2));
                  useToastStore.getState().addToast({
                    title: 'Copied',
                    message: 'Diagnostic info copied to clipboard',
                    type: 'success',
                  });
                }}>
                <Icon name="copy" size="sm" className="mr-2" />
                Copy Diagnostic Info
              </Button>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'troubleshooting',
      title: 'Troubleshooting',
      icon: 'alert-triangle',
      content: (
        <div className="space-y-4">
          <Card className="border-l-4 border-l-amber-500">
            <CardHeader>
              <CardTitle>Common Issues</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Typography variant="subtitle" className="font-semibold text-sm text-amber-700 dark:text-amber-400">
                  Connection Refused / 404
                </Typography>
                <Typography variant="body" className="text-xs">
                  Ensure your proxy server is running. Check if the port (default 3006) matches the URI in Server
                  Status.
                </Typography>
              </div>
              <div className="border-t border-slate-100 dark:border-slate-800 pt-3">
                <Typography variant="subtitle" className="font-semibold text-sm text-amber-700 dark:text-amber-400">
                  Tools Not Showing
                </Typography>
                <Typography variant="body" className="text-xs">
                  Click the "Refresh" button in the Available Tools tab. Ensure your <code>config.json</code> is correct
                  and the MCP server is healthy.
                </Typography>
              </div>
              <div className="border-t border-slate-100 dark:border-slate-800 pt-3">
                <Typography variant="subtitle" className="font-semibold text-sm text-amber-700 dark:text-amber-400">
                  Extension Context Invalidated
                </Typography>
                <Typography variant="body" className="text-xs">
                  This happens if the extension is updated or reloaded. Refresh the web page to reconnect.
                </Typography>
              </div>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'faq',
      title: 'FAQ',
      icon: 'help-circle',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Frequently Asked Questions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Typography variant="subtitle" className="font-semibold text-sm mb-1">
                  Is my data secure?
                </Typography>
                <div className="p-2 bg-slate-50 dark:bg-slate-800 rounded">
                  <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400">
                    Yes. The extension communicates directly with your local proxy. Your data (files, etc.) stays local
                    unless you explicitly send it to the AI.
                  </Typography>
                </div>
              </div>
              <div>
                <Typography variant="subtitle" className="font-semibold text-sm mb-1">
                  Which AI models work best?
                </Typography>
                <div className="p-2 bg-slate-50 dark:bg-slate-800 rounded">
                  <Typography variant="body" className="text-xs text-slate-600 dark:text-slate-400">
                    Models with strong function calling capabilities (like GPT-4, Claude 3.5 Sonnet) work best.
                  </Typography>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      ),
    },
    {
      id: 'support',
      title: 'Support',
      icon: 'life-buoy',
      content: (
        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Contact & Community</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              <Typography variant="body" className="text-sm text-slate-600 dark:text-slate-400">
                Need more help? Found a bug? Join the community.
              </Typography>
              <div className="flex flex-col gap-3 mt-4">
                <a
                  href="https://github.com/srbhptl39/TormentNexus-Extension/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="no-underline">
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full justify-start h-10 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <Icon name="life-buoy" size="sm" className="mr-3 text-primary-500" />
                    Report an Issue on GitHub
                  </Button>
                </a>
                <a
                  href="https://github.com/srbhptl39/TormentNexus-Extension"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="no-underline">
                  <Button
                    variant="outline"
                    size="sm"
                    className="w-full justify-start h-10 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <Icon name="arrow-up-right" size="sm" className="mr-3 text-primary-500" />
                    Project Repository
                  </Button>
                </a>
              </div>
            </CardContent>
          </Card>
        </div>
      ),
    },
  ];

  const filteredSections = useMemo(() => {
    if (!searchQuery) return sections;
    const lowerQuery = searchQuery.toLowerCase();

    // First, verify if a section title matches
    const titleMatches = sections.filter(section => section.title.toLowerCase().includes(lowerQuery));

    // If we have title matches, return them
    if (titleMatches.length > 0) return titleMatches;

    // Otherwise try to match content (this is a simplified approach)
    // In a real app we might index the content text properly
    return sections.filter(section => {
      // Basic check if common keywords related to the section match
      if (section.id === 'setup' && ['config', 'proxy', 'install', 'connect'].some(k => lowerQuery.includes(k)))
        return true;
      if (
        section.id === 'troubleshooting' &&
        ['error', 'fail', 'bug', 'issue', 'problem'].some(k => lowerQuery.includes(k))
      )
        return true;
      if (section.id === 'security' && ['privacy', 'data', 'safe'].some(k => lowerQuery.includes(k))) return true;
      return false;
    });
  }, [searchQuery]);

  // If search yields only one result, automatically select it
  React.useEffect(() => {
    if (searchQuery && filteredSections.length === 1 && filteredSections[0].id !== activeSection) {
      setActiveSection(filteredSections[0].id);
    }
  }, [filteredSections, searchQuery]);

  return (
    <div className="flex flex-col h-full space-y-4 p-4">
      <div className="flex flex-col space-y-2 flex-shrink-0">
        <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100">
          Help & Documentation
        </Typography>
        <div className="relative">
          <input
            type="text"
            placeholder="Search help..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="w-full px-3 py-2 pl-9 text-xs border border-slate-300 dark:border-slate-600 rounded bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-400 dark:placeholder-slate-500 transition-shadow"
          />
          <div className="absolute left-2.5 top-2.5">
            <Icon name="search" size="xs" className="text-slate-400 dark:text-slate-500" />
          </div>
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-2.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300">
              <Icon name="x" size="xs" />
            </button>
          )}
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden gap-4 min-h-0">
        {/* Navigation Sidebar */}
        <div className="w-1/3 flex flex-col space-y-1 overflow-y-auto pr-2 border-r border-slate-200 dark:border-slate-700 scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
          {sections.map(section => {
            const isVisible = filteredSections.find(s => s.id === section.id);
            if (!isVisible && searchQuery) return null;

            return renderNavButton(section.id, section.title, section.icon);
          })}

          {searchQuery && filteredSections.length === 0 && (
            <div className="p-2 text-xs text-center text-slate-500 dark:text-slate-400 italic">No matches found</div>
          )}
        </div>

        {/* Content Area */}
        <div className="w-2/3 overflow-y-auto pl-2 pb-2 scrollbar-thin scrollbar-thumb-slate-300 dark:scrollbar-thumb-slate-600">
          {filteredSections.map(
            section =>
              activeSection === section.id && (
                <div key={section.id} className="animate-in fade-in slide-in-from-right-2 duration-300">
                  {section.content}
                </div>
              ),
          )}
          {filteredSections.length > 0 && !filteredSections.find(s => s.id === activeSection) && (
            <div className="flex flex-col items-center justify-center h-full text-slate-400 dark:text-slate-500">
              <Icon name="search" size="lg" className="mb-2 opacity-50" />
              <Typography variant="body" className="text-sm">
                Select a topic from the results
              </Typography>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Help;
