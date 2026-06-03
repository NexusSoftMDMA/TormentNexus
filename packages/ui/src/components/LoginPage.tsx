import React from 'react';
import { LoginForm } from "./auth/LoginForm";
import { Brain } from "lucide-react";

export function LoginPage() {
    return (
        <div className="container relative h-[100vh] flex-col items-center justify-center grid lg:max-w-none lg:grid-cols-2 lg:px-0">
            {/* Left Column: Branding / Info */}
            <div className="relative hidden h-full flex-col bg-zinc-900 p-10 text-white dark:border-r border-zinc-800 lg:flex">
                <div className="absolute inset-0 bg-zinc-900" />
                <div className="absolute inset-0 bg-[url('https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?q=80&w=2564&auto=format&fit=crop')] bg-cover bg-center opacity-20 mix-blend-overlay" />

                <div className="relative z-20 flex items-center text-lg font-medium">
                    <Brain className="mr-2 h-6 w-6" />
                    TORMENTNEXUS SYSTEM
                </div>
                <div className="relative z-20 mt-auto">
                    <blockquote className="space-y-2">
                        <p className="text-lg">
                            &ldquo;Resistance is futile. You will be authenticated.&rdquo;
                        </p>
                        <footer className="text-sm text-zinc-400">The Collective</footer>
                    </blockquote>
                </div>
            </div>

            {/* Right Column: Form */}
            <div className="lg:p-8 relative">
                <div className="absolute inset-0 bg-gradient-to-br from-zinc-900/0 via-violet-500/5 to-zinc-900/0 pointer-events-none" />
                <div className="mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]">
                    <div className="flex flex-col space-y-2 text-center">
                        <h1 className="text-2xl font-semibold tracking-tight text-white">
                            Initialize Session
                        </h1>
                        <p className="text-sm text-zinc-400">
                            Enter your credentials to access the neural link.
                        </p>
                    </div>
                    <LoginForm />
                    <p className="px-8 text-center text-sm text-zinc-500">
                        By clicking continue, you agree to our{" "}
                        <a href="/terms" className="underline underline-offset-4 hover:text-primary">
                            Terms of Service
                        </a>{" "}
                        and{" "}
                        <a href="/privacy" className="underline underline-offset-4 hover:text-primary">
                            Privacy Policy
                        </a>
                        .
                    </p>
                </div>
            </div>
        </div>
    );
}
