'use client';

import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from './ui/card';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Switch } from './ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Textarea } from './ui/textarea';
import { Settings2, Save, RefreshCw, Target, MessageSquare, Bot } from 'lucide-react';
import { trpc } from '../utils/trpc';

export function DirectorConfigPanel() {
    const [localConfig, setLocalConfig] = useState<any>(null);
    const [isSaving, setIsSaving] = useState(false);

    const { data: config, refetch } = trpc.directorConfig.get.useQuery();
    const updateMutation = trpc.directorConfig.update.useMutation({
        onSuccess: () => {
            refetch();
            setIsSaving(false);
        },
        onError: () => setIsSaving(false)
    });

    useEffect(() => {
        if (config && !localConfig) {
            setLocalConfig(config);
        }
    }, [config]);

    const handleSave = () => {
        if (!localConfig) return;
        setIsSaving(true);
        updateMutation.mutate(localConfig);
    };

    const updateField = (field: string, value: any) => {
        setLocalConfig((prev: any) => ({ ...prev, [field]: value }));
    };

    if (!localConfig) {
        return (
            <Card>
                <CardContent className="py-8 text-center text-muted-foreground">
                    Loading configuration...
                </CardContent>
            </Card>
        );
    }

    return (
        <Card className="h-full">
            <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                    <div>
                        <CardTitle className="text-sm font-medium flex items-center gap-2">
                            <Settings2 className="w-4 h-4" />
                            Director Configuration
                        </CardTitle>
                        <CardDescription className="text-xs">
                            Customize Director behavior and persona
                        </CardDescription>
                    </div>
                    <div className="flex gap-2">
                        <Button size="sm" variant="outline" onClick={() => refetch()}>
                            <RefreshCw className="w-3 h-3" />
                        </Button>
                        <Button size="sm" onClick={handleSave} disabled={isSaving}>
                            <Save className="w-3 h-3 mr-1" />
                            {isSaving ? 'Saving...' : 'Save'}
                        </Button>
                    </div>
                </div>
            </CardHeader>
            <CardContent className="space-y-4 overflow-y-auto max-h-[500px]">
                {/* Focus Topic */}
                <div className="space-y-2">
                    <Label className="text-xs flex items-center gap-1">
                        <Target className="w-3 h-3" />
                        Focus Topic
                    </Label>
                    <Input
                        value={localConfig.defaultTopic || ''}
                        onChange={(e) => updateField('defaultTopic', e.target.value)}
                        placeholder="What should Director focus on?"
                        className="h-8 text-xs"
                    />
                </div>

                {/* Persona */}
                <div className="space-y-2">
                    <Label className="text-xs flex items-center gap-1">
                        <Bot className="w-3 h-3" />
                        Persona
                    </Label>
                    <Select
                        value={localConfig.persona || 'default'}
                        onValueChange={(v) => updateField('persona', v)}
                    >
                        <SelectTrigger className="h-8 text-xs">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="default">Default</SelectItem>
                            <SelectItem value="homie">Homie 🤙</SelectItem>
                            <SelectItem value="professional">Professional 💼</SelectItem>
                            <SelectItem value="chaos">Chaos 🎲</SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                {/* Chat Prefix */}
                <div className="space-y-2">
                    <Label className="text-xs flex items-center gap-1">
                        <MessageSquare className="w-3 h-3" />
                        Chat Prefix
                    </Label>
                    <Input
                        value={localConfig.chatPrefix || ''}
                        onChange={(e) => updateField('chatPrefix', e.target.value)}
                        placeholder="[Director]:"
                        className="h-8 text-xs"
                    />
                </div>

                {/* Council Prefix */}
                <div className="space-y-2">
                    <Label className="text-xs">Council Prefix</Label>
                    <Input
                        value={localConfig.councilPrefix || ''}
                        onChange={(e) => updateField('councilPrefix', e.target.value)}
                        placeholder="🏛️ [Council]:"
                        className="h-8 text-xs"
                    />
                </div>

                {/* Director Action Prefix */}
                <div className="space-y-2">
                    <Label className="text-xs">Director Action Prefix</Label>
                    <Input
                        value={localConfig.directorActionPrefix || ''}
                        onChange={(e) => updateField('directorActionPrefix', e.target.value)}
                        placeholder="🎬 **Director Action**:"
                        className="h-8 text-xs"
                    />
                </div>

                {/* Custom Instructions */}
                <div className="space-y-2">
                    <Label className="text-xs">Custom Instructions</Label>
                    <Textarea
                        value={localConfig.customInstructions || ''}
                        onChange={(e) => updateField('customInstructions', e.target.value)}
                        placeholder="Additional instructions for Director..."
                        className="text-xs min-h-[60px]"
                    />
                </div>

                {/* Toggles */}
                <div className="space-y-3 pt-2 border-t">
                    <div className="flex items-center justify-between">
                        <Label className="text-xs">Enable Council</Label>
                        <Switch
                            checked={localConfig.enableCouncil ?? true}
                            onCheckedChange={(v) => updateField('enableCouncil', v)}
                        />
                    </div>
                    <div className="flex items-center justify-between">
                        <Label className="text-xs">Enable Chat Paste</Label>
                        <Switch
                            checked={localConfig.enableChatPaste ?? true}
                            onCheckedChange={(v) => updateField('enableChatPaste', v)}
                        />
                    </div>
                    <div className="flex items-center justify-between">
                        <Label className="text-xs">Auto-Submit Chat</Label>
                        <Switch
                            checked={localConfig.autoSubmitChat ?? false}
                            onCheckedChange={(v) => updateField('autoSubmitChat', v)}
                        />
                    </div>
                    <div className="flex items-center justify-between">
                        <Label className="text-xs">Verbose Logging</Label>
                        <Switch
                            checked={localConfig.verboseLogging ?? false}
                            onCheckedChange={(v) => updateField('verboseLogging', v)}
                        />
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
