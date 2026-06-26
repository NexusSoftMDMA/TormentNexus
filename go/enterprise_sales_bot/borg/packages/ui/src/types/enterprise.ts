export type UserRole = 'admin' | 'developer' | 'operator' | 'viewer';

export interface RoleDefinition {
  role: UserRole;
  permissions: string[];
  description: string;
}

export interface UserAccess {
  userId: string;
  role: UserRole;
}
