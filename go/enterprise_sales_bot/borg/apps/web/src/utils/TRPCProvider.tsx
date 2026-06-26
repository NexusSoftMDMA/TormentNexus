"use client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { httpBatchLink, splitLink, unstable_httpSubscriptionLink } from "@trpc/client";
import { resolveTrpcHttpUrl } from "@tormentnexus/ui";
import React, { useState } from "react";
import { trpc } from "./trpc";

export function TRPCProvider({ children }: { children: React.ReactNode }) {
    const [queryClient] = useState(() => new QueryClient());
    const [trpcClient] = useState(() =>
        trpc.createClient({
            links: [
                splitLink({
                    condition: (op) => 
                        op.type === 'subscription' || 
                        op.path.toLowerCase().includes('subscribe') ||
                        op.path.startsWith('healer.'),
                    true: unstable_httpSubscriptionLink({
                        url: resolveTrpcHttpUrl(process.env.NEXT_PUBLIC_TRPC_URL),
                    }),
                    false: httpBatchLink({
                        url: resolveTrpcHttpUrl(process.env.NEXT_PUBLIC_TRPC_URL),
                    }),
                }),
            ],
        })
    );

    return (
        <trpc.Provider client={trpcClient} queryClient={queryClient}>
            <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
        </trpc.Provider>
    );
}
