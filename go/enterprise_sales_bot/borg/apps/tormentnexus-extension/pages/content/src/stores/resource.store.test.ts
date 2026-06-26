import { describe, it, expect, beforeEach } from 'vitest';
import { useResourceStore } from './resource.store';

describe('useResourceStore', () => {
    beforeEach(() => {
        useResourceStore.setState({
            resources: [],
            templates: [],
            isLoading: false,
            selectedResourceUri: null,
            resourceContent: null,
            error: null
        });
    });

    it('initializes with empty state', () => {
        const state = useResourceStore.getState();
        expect(state.resources).toEqual([]);
        expect(state.templates).toEqual([]);
        expect(state.selectedResourceUri).toBeNull();
    });

    it('updates resources and templates', () => {
        const mockResources = [{ uri: 'test://1', name: 'Test 1' }];
        const mockTemplates = [{ uriTemplate: 'test://{id}', name: 'Test Template' }];

        useResourceStore.getState().setResources(mockResources, mockTemplates);

        const state = useResourceStore.getState();
        expect(state.resources).toHaveLength(1);
        expect(state.resources[0].name).toBe('Test 1');
        expect(state.templates).toHaveLength(1);
    });

    it('selects a resource', () => {
        useResourceStore.getState().selectResource('test://resource');
        expect(useResourceStore.getState().selectedResourceUri).toBe('test://resource');
    });

    it('deselecting a resource clears content', () => {
        const store = useResourceStore.getState();
        store.setResourceContent('test://resource', 'content data');
        expect(useResourceStore.getState().resourceContent).not.toBeNull();

        useResourceStore.getState().selectResource(null);
        expect(useResourceStore.getState().selectedResourceUri).toBeNull();
        // Setting selectResource(null) should also trigger clearResourceContent internally
        expect(useResourceStore.getState().resourceContent).toBeNull();
    });

    it('handles loading state', () => {
        useResourceStore.getState().setLoading(true);
        expect(useResourceStore.getState().isLoading).toBe(true);

        useResourceStore.getState().setLoading(false);
        expect(useResourceStore.getState().isLoading).toBe(false);
    });

    it('handles error state', () => {
        useResourceStore.getState().setError('Failed to fetch');
        expect(useResourceStore.getState().error).toBe('Failed to fetch');

        useResourceStore.getState().setError(null);
        expect(useResourceStore.getState().error).toBeNull();
    });
});
