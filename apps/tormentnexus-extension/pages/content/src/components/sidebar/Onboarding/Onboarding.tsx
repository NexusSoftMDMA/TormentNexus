import type React from 'react';
import { useState, useEffect } from 'react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@src/components/ui/card';
import { Typography, Button, Icon } from '../ui';
import { cn } from '@src/lib/utils';
import { getExtensionStorageValue, setExtensionStorageValue } from '@src/stores/extension-storage';

interface OnboardingStep {
  title: string;
  description: string;
  icon: any;
  image?: string; // Placeholder for potential image
}

const steps: OnboardingStep[] = [
  {
    title: 'Welcome to TormentNexus Extension',
    description:
      'Empower your AI with real-world tools. This sidebar is your control center for connecting local data, files, and APIs to ChatGPT, Claude, and more.',
    icon: 'lightning',
  },
  {
    title: 'Connect Your Proxy',
    description:
      'First, ensure your local MCP proxy is running (check the "Help" tab for setup instructions). Use the Server Status panel above to connect.',
    icon: 'server',
  },
  {
    title: 'Explore Tools',
    description:
      'Once connected, your tools will appear in the "Available Tools" tab. You can favorite, sort, and search them to keep your workflow efficient.',
    icon: 'tool',
  },
  {
    title: 'Automate & Monitor',
    description:
      'Use "Settings" to configure auto-execution for trusted tools. Track everything in the "Activity" log and "Dashboard" for full observability.',
    icon: 'activity',
  },
];

const Onboarding: React.FC = () => {
  const [currentStep, setCurrentStep] = useState(0);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;

    void (async () => {
      const completed = await getExtensionStorageValue('mcpOnboardingCompleted');
      if (!completed) {
        timeoutId = setTimeout(() => setIsVisible(true), 1000);
      }
    })();

    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    };
  }, []);

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep(prev => prev + 1);
    } else {
      handleComplete();
    }
  };

  const handleComplete = () => {
    setIsVisible(false);
    void setExtensionStorageValue('mcpOnboardingCompleted', 'true');
    // Optionally open Help tab?
  };

  if (!isVisible) return null;

  const step = steps[currentStep];

  return (
    <div className="absolute inset-0 z-[9000] flex items-center justify-center bg-black/60 backdrop-blur-[2px] p-6 animate-in fade-in duration-300">
      <Card className="w-full max-w-sm border-slate-200 dark:border-slate-700 shadow-2xl relative overflow-hidden">
        {/* Progress Bar */}
        <div className="absolute top-0 left-0 right-0 h-1 bg-slate-100 dark:bg-slate-800">
          <div
            className="h-full bg-blue-600 transition-all duration-300 ease-out"
            style={{ width: `${((currentStep + 1) / steps.length) * 100}%` }}
          />
        </div>

        <CardHeader className="text-center pt-8 pb-2">
          <div className="mx-auto w-12 h-12 bg-blue-50 dark:bg-blue-900/30 rounded-full flex items-center justify-center mb-4 text-blue-600 dark:text-blue-400">
            <Icon name={step.icon} size="lg" />
          </div>
          <CardTitle className="text-xl">{step.title}</CardTitle>
        </CardHeader>

        <CardContent className="text-center pb-6">
          <Typography variant="body" className="text-slate-600 dark:text-slate-300">
            {step.description}
          </Typography>
        </CardContent>

        <CardFooter className="flex justify-between border-t border-slate-100 dark:border-slate-800 p-4 bg-slate-50 dark:bg-slate-900/50">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleComplete}
            className="text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200">
            Skip
          </Button>

          <div className="flex items-center gap-1">
            <div className="flex gap-1 mr-4">
              {steps.map((_, i) => (
                <div
                  key={i}
                  className={cn(
                    'w-1.5 h-1.5 rounded-full transition-colors',
                    i === currentStep ? 'bg-blue-600' : 'bg-slate-300 dark:bg-slate-700',
                  )}
                />
              ))}
            </div>
            <Button onClick={handleNext} size="sm" className="bg-blue-600 hover:bg-blue-700 text-white">
              {currentStep === steps.length - 1 ? 'Get Started' : 'Next'}
            </Button>
          </div>
        </CardFooter>
      </Card>
    </div>
  );
};

export default Onboarding;
