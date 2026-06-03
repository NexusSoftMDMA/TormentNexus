"use client";

import { Suspense, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { Card, CardHeader, CardTitle, CardContent } from "@tormentnexus/ui";
import { Loader2, CheckCircle, XCircle } from "lucide-react";
import { trpc } from '@/utils/trpc';
import { toast } from 'sonner';

function OAuthCallbackContent() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const error = searchParams.get('error');

    const [status, setStatus] = useState<'PROCESSING' | 'SUCCESS' | 'ERROR'>('PROCESSING');

    // Assuming we have an oauth router to generic handle exchange
    // Or we might need to know WHICH provider from state/localStorage
    const exchangeMutation = trpc.oauth.exchange.useMutation({
        onSuccess: (data) => {
            setStatus('SUCCESS');
            toast.success("Authentication successful");
            setTimeout(() => {
                // Close popup if popup, or redirect
                if (window.opener) {
                    window.opener.postMessage({ type: 'OAUTH_SUCCESS', data }, '*');
                    window.close();
                } else {
                    router.push('/dashboard/mcp/settings');
                }
            }, 1000);
        },
        onError: (err) => {
            setStatus('ERROR');
            toast.error(`Auth failed: ${err.message}`);
        }
    });

    useEffect(() => {
        if (error) {
            setStatus('ERROR');
            return;
        }

        if (code && state) {
            exchangeMutation.mutate({ code, state });
        } else {
            setStatus('ERROR');
        }
    }, [code, state, error]);

    return (
        <div className="flex items-center justify-center min-h-screen bg-black">
            <Card className="w-[400px] bg-zinc-900 border-zinc-800">
                <CardHeader className="text-center">
                    <CardTitle className="text-white">
                        {status === 'PROCESSING' && "Authenticating..."}
                        {status === 'SUCCESS' && "Connected!"}
                        {status === 'ERROR' && "Authentication Failed"}
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex justify-center p-8">
                    {status === 'PROCESSING' && <Loader2 className="h-12 w-12 text-blue-500 animate-spin" />}
                    {status === 'SUCCESS' && <CheckCircle className="h-12 w-12 text-green-500" />}
                    {status === 'ERROR' && (
                        <div className="text-center">
                            <XCircle className="h-12 w-12 text-red-500 mx-auto mb-2" />
                            <p className="text-zinc-500 text-sm">
                                {error || exchangeMutation.error?.message || "Invalid response parameters"}
                            </p>
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

function OAuthCallbackFallback() {
    return (
        <div className="flex items-center justify-center min-h-screen bg-black">
            <Card className="w-[400px] bg-zinc-900 border-zinc-800">
                <CardHeader className="text-center">
                    <CardTitle className="text-white">Authenticating...</CardTitle>
                </CardHeader>
                <CardContent className="flex justify-center p-8">
                    <Loader2 className="h-12 w-12 text-blue-500 animate-spin" />
                </CardContent>
            </Card>
        </div>
    );
}

export default function OAuthCallbackPage() {
    return (
        <Suspense fallback={<OAuthCallbackFallback />}>
            <OAuthCallbackContent />
        </Suspense>
    );
}
