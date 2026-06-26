// Integration catalog utility functions
export type StartupStatusSummary = Record<string, any>;
export const getBridgeClientEmptyStateMessage = (_overview?: any): string => 'No bridge clients connected.';
export const getBridgeClientStatDetail = (_client: any): string => '';
export const getConnectedBridgeClientRows = (_status: any): any[] => [];
export const getExternalClientRows = (_status: any): any[] => [];
export const getInstallSurfaceRows = (_data: any): any[] => [];
export const getIntegrationOverview = (_status: any, _browser?: any, _sync?: any, _cli?: any): Record<string, any> => ({});
export const getStatusBadgeClasses = (_status: string): string => '';
