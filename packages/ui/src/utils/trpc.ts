import { createTRPCReact } from '@trpc/react-query';

// Intentionally erase the client helper type surface.
// The server router remains authoritative, while the dashboard's app router is
// large enough to trip tRPC/TypeScript recursion and key-collision diagnostics.
export const trpc: any = createTRPCReact<any>();
