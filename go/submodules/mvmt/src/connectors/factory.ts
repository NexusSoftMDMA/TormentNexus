import { ProxyConfig } from '../config/schema.js';
import { Connector } from './types.js';
import { HttpProxyConnector } from './proxy-http.js';
import { StdioProxyConnector } from './proxy-stdio.js';

export function createProxyConnector(proxyConfig: ProxyConfig): Connector | undefined {
  if (proxyConfig.transport === 'http' && proxyConfig.url) {
    return new HttpProxyConnector({
      name: proxyConfig.name,
      url: proxyConfig.url,
      env: proxyConfig.env,
      writeAccess: proxyConfig.writeAccess,
    });
  }

  if (proxyConfig.transport === 'stdio' && proxyConfig.command) {
    return new StdioProxyConnector({
      name: proxyConfig.name,
      command: proxyConfig.command,
      args: proxyConfig.args || [],
      env: proxyConfig.env,
      writeAccess: proxyConfig.writeAccess,
    });
  }

  return undefined;
}
