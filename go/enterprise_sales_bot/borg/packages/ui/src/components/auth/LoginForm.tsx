"use client";

import React, { useState } from 'react';
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { Loader2, Mail, Lock, ArrowRight } from "lucide-react";
import { cn } from "../../lib/utils";

interface LoginFormProps extends React.HTMLAttributes<HTMLDivElement> {
    onSuccess?: () => void;
}

export function LoginForm({ className, onSuccess, ...props }: LoginFormProps) {
    const [isLoading, setIsLoading] = useState<boolean>(false);

    async function onSubmit(event: React.SyntheticEvent) {
        event.preventDefault();
        setIsLoading(true);

        setTimeout(() => {
            setIsLoading(false);
            if (onSuccess) onSuccess();
        }, 1500);
    }

    return (
        <div className={cn("grid gap-6", className)} {...props}>
            <form onSubmit={onSubmit}>
                <div className="grid gap-4">
                    <div className="grid gap-2">
                        <Label className="sr-only" htmlFor="email">
                            Email
                        </Label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <Mail className="h-4 w-4 text-zinc-400" />
                            </div>
                            <Input
                                id="email"
                                placeholder="name@example.com"
                                type="email"
                                autoCapitalize="none"
                                autoComplete="email"
                                autoCorrect="off"
                                disabled={isLoading}
                                className="pl-10 bg-zinc-900/50 border-zinc-800 focus:ring-zinc-700 focus:border-zinc-700"
                            />
                        </div>
                    </div>
                    <div className="grid gap-2">
                        <Label className="sr-only" htmlFor="password">
                            Password
                        </Label>
                        <div className="relative">
                            <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                                <Lock className="h-4 w-4 text-zinc-400" />
                            </div>
                            <Input
                                id="password"
                                placeholder="Password"
                                type="password"
                                autoComplete="current-password"
                                disabled={isLoading}
                                className="pl-10 bg-zinc-900/50 border-zinc-800 focus:ring-zinc-700 focus:border-zinc-700"
                            />
                        </div>
                    </div>
                    <Button disabled={isLoading} className="bg-white text-black hover:bg-zinc-200">
                        {isLoading && (
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        )}
                        Sign In with Email <ArrowRight className="ml-2 h-4 w-4" />
                    </Button>
                </div>
            </form>
            <div className="relative">
                <div className="absolute inset-0 flex items-center">
                    <span className="w-full border-t border-zinc-800" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-[#0a0a0f] px-2 text-zinc-500">
                        Or continue with
                    </span>
                </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
                <Button variant="outline" type="button" disabled={isLoading} className="border-zinc-800 bg-zinc-900/50 hover:bg-zinc-900 hover:text-white">
                    Git (Mock)
                </Button>
                <Button variant="outline" type="button" disabled={isLoading} className="border-zinc-800 bg-zinc-900/50 hover:bg-zinc-900 hover:text-white">
                    Google (Mock)
                </Button>
            </div>
        </div>
    );
}
