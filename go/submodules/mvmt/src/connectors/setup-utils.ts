import { MvmtConfig, ProxyConfig } from '../config/schema.js';

export function sameProxyName(left: string, right: string): boolean {
  return left.toLowerCase() === right.toLowerCase();
}

export function upsertProxyConfig(config: MvmtConfig, proxyConfig: ProxyConfig): MvmtConfig {
  const proxy = config.proxy.filter((entry) => !sameProxyName(entry.name, proxyConfig.name));
  proxy.push(proxyConfig);
  return { ...config, proxy };
}
