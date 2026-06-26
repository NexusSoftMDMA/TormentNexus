/**
 * Prompt Templates Component
 *
 * Sidebar tab for managing reusable prompt templates.
 * Supports creating, editing, deleting, searching, and copying templates to clipboard.
 */

import React, { useState, useMemo } from 'react';
import { usePromptStore, type PromptTemplate } from '@src/stores';
import { Button, Input, Textarea, Typography, Icon } from '../ui';
import { useToastStore } from '@src/stores';

const PromptTemplates: React.FC = () => {
    const { templates, addTemplate, updateTemplate, deleteTemplate } = usePromptStore();
    const addToast = useToastStore(s => s.addToast);

    // Local state
    const [searchQuery, setSearchQuery] = useState('');
    const [isEditing, setIsEditing] = useState(false);
    const [editingTemplate, setEditingTemplate] = useState<PromptTemplate | null>(null);
    const [formTitle, setFormTitle] = useState('');
    const [formContent, setFormContent] = useState('');

    // Filtered templates based on search
    const filteredTemplates = useMemo(() => {
        if (!searchQuery.trim()) return templates;
        const query = searchQuery.toLowerCase();
        return templates.filter(
            t =>
                t.title.toLowerCase().includes(query) ||
                t.content.toLowerCase().includes(query) ||
                t.tags.some(tag => tag.toLowerCase().includes(query)),
        );
    }, [templates, searchQuery]);

    // Copy template content to clipboard
    const handleCopy = async (template: PromptTemplate) => {
        try {
            await navigator.clipboard.writeText(template.content);
            addToast({ title: 'Copied!', message: `"${template.title}" copied to clipboard.`, type: 'success', duration: 2000 });
        } catch {
            addToast({ title: 'Copy failed', message: 'Could not access clipboard.', type: 'error', duration: 3000 });
        }
    };

    // Start creating a new template
    const handleNew = () => {
        setEditingTemplate(null);
        setFormTitle('');
        setFormContent('');
        setIsEditing(true);
    };

    // Start editing an existing template
    const handleEdit = (template: PromptTemplate) => {
        setEditingTemplate(template);
        setFormTitle(template.title);
        setFormContent(template.content);
        setIsEditing(true);
    };

    // Save (create or update)
    const handleSave = () => {
        if (!formTitle.trim() || !formContent.trim()) {
            addToast({ title: 'Missing fields', message: 'Title and content are required.', type: 'warning', duration: 3000 });
            return;
        }
        if (editingTemplate) {
            updateTemplate(editingTemplate.id, { title: formTitle, content: formContent });
            addToast({ title: 'Updated', message: `"${formTitle}" saved.`, type: 'success', duration: 2000 });
        } else {
            addTemplate(formTitle, formContent);
            addToast({ title: 'Created', message: `"${formTitle}" added.`, type: 'success', duration: 2000 });
        }
        setIsEditing(false);
        setEditingTemplate(null);
        setFormTitle('');
        setFormContent('');
    };

    // Cancel editing
    const handleCancel = () => {
        setIsEditing(false);
        setEditingTemplate(null);
        setFormTitle('');
        setFormContent('');
    };

    // Delete with confirmation
    const handleDelete = (template: PromptTemplate) => {
        deleteTemplate(template.id);
        addToast({ title: 'Deleted', message: `"${template.title}" removed.`, type: 'info', duration: 2000 });
    };

    // --- EDITOR VIEW ---
    if (isEditing) {
        return (
            <div className="flex flex-col h-full bg-slate-50 dark:bg-slate-900 overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800">
                    <div className="flex items-center gap-2">
                        <Button variant="ghost" size="icon" onClick={handleCancel}>
                            <Icon name="arrow-left" size="sm" />
                        </Button>
                        <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-100">
                            {editingTemplate ? 'Edit Template' : 'New Template'}
                        </Typography>
                    </div>
                    <Button onClick={handleSave} size="sm" className="bg-primary-600 hover:bg-primary-700 text-white">
                        <Icon name="save" size="xs" className="mr-1" />
                        Save
                    </Button>
                </div>

                {/* Form */}
                <div className="flex-1 overflow-y-auto p-4 space-y-4">
                    <div>
                        <label className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-1 block">Title</label>
                        <Input
                            value={formTitle}
                            onChange={e => setFormTitle(e.target.value)}
                            placeholder="e.g., Code Review Prompt"
                            className="w-full"
                        />
                    </div>
                    <div>
                        <label className="text-xs font-medium text-slate-700 dark:text-slate-300 mb-1 block">Content</label>
                        <Textarea
                            value={formContent}
                            onChange={e => setFormContent(e.target.value)}
                            placeholder="Enter your prompt template text..."
                            className="w-full h-48"
                        />
                    </div>
                </div>
            </div>
        );
    }

    // --- LIST VIEW ---
    return (
        <div className="flex flex-col h-full">
            {/* Header */}
            <div className="flex items-center justify-between p-3 border-b border-slate-200 dark:border-slate-700">
                <Typography variant="h4" className="font-semibold text-slate-800 dark:text-slate-200">
                    Prompt Templates
                </Typography>
                <Button onClick={handleNew} size="sm" className="bg-primary-600 hover:bg-primary-700 text-white">
                    <Icon name="plus" size="xs" className="mr-1" />
                    New
                </Button>
            </div>

            {/* Search */}
            {templates.length > 0 && (
                <div className="p-3 border-b border-slate-200 dark:border-slate-700">
                    <Input
                        value={searchQuery}
                        onChange={e => setSearchQuery(e.target.value)}
                        placeholder="Search templates..."
                        className="w-full"
                    />
                </div>
            )}

            {/* Template List */}
            <div className="flex-1 overflow-y-auto p-3 space-y-2">
                {filteredTemplates.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-12 text-slate-400 dark:text-slate-500">
                        <Icon name="file-text" size="lg" className="mb-3 opacity-50" />
                        <Typography variant="body" className="text-center">
                            {templates.length === 0
                                ? 'No templates yet. Create your first prompt template!'
                                : 'No templates match your search.'}
                        </Typography>
                    </div>
                ) : (
                    filteredTemplates.map(template => (
                        <div
                            key={template.id}
                            className="bg-white dark:bg-slate-800 rounded-lg border border-slate-200 dark:border-slate-700 p-3 hover:shadow-sm transition-shadow"
                        >
                            {/* Template title */}
                            <div className="flex items-center justify-between mb-1">
                                <Typography variant="subtitle" className="font-medium text-slate-800 dark:text-slate-200 truncate flex-1">
                                    {template.title}
                                </Typography>
                            </div>

                            {/* Content preview (first 100 chars) */}
                            <Typography variant="caption" className="text-slate-500 dark:text-slate-400 line-clamp-2 mb-2">
                                {template.content.length > 100
                                    ? `${template.content.substring(0, 100)}...`
                                    : template.content}
                            </Typography>

                            {/* Action buttons */}
                            <div className="flex items-center gap-1 justify-end">
                                <Button variant="ghost" size="sm" onClick={() => handleCopy(template)} title="Copy to clipboard">
                                    <Icon name="box" size="xs" />
                                </Button>
                                <Button variant="ghost" size="sm" onClick={() => handleEdit(template)} title="Edit">
                                    <Icon name="settings" size="xs" />
                                </Button>
                                <Button variant="ghost" size="sm" onClick={() => handleDelete(template)} title="Delete"
                                    className="text-red-500 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300">
                                    <Icon name="x" size="xs" />
                                </Button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default PromptTemplates;
