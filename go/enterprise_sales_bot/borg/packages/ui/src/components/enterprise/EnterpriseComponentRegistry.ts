import React from 'react';

export interface EnterpriseComponents {
  OidcConfig: React.ComponentType<any> | null;
  RbacManager: React.ComponentType<any> | null;
  AuditLogViewer: React.ComponentType<any> | null;
}

export const enterpriseRegistry: EnterpriseComponents = {
  OidcConfig: null,
  RbacManager: null,
  AuditLogViewer: null,
};
