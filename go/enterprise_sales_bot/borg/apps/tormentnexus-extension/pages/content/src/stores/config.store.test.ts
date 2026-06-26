import { describe, it, expect, beforeEach } from 'vitest';
import { useConfigStore } from './config.store';

describe('useConfigStore', () => {
    beforeEach(() => {
        // Reset store before each test
        useConfigStore.getState().resetState();
    });

    it('initializes with default values', () => {
        const state = useConfigStore.getState();
        expect(state.featureFlags).toEqual({});
        expect(state.userSegment).toBe('new');
        expect(state.notificationConfig.enabled).toBe(true);
    });

    it('updates feature flags correctly', () => {
        useConfigStore.getState().updateFeatureFlags({
            'test-feature': { enabled: true, rollout: 100 }
        });

        const state = useConfigStore.getState();
        expect(state.featureFlags['test-feature']).toBeDefined();
        expect(state.featureFlags['test-feature'].enabled).toBe(true);
    });

    it('isFeatureEnabled handles 100% rollout', () => {
        useConfigStore.getState().updateFeatureFlags({
            'fully-rolled': { enabled: true, rollout: 100 },
            'disabled-feature': { enabled: false, rollout: 100 }
        });

        const state = useConfigStore.getState();
        expect(state.isFeatureEnabled('fully-rolled')).toBe(true);
        expect(state.isFeatureEnabled('disabled-feature')).toBe(false);
        expect(state.isFeatureEnabled('missing-feature')).toBe(false);
    });

    it('setUserProperties updates segment automatically based on usageDays', () => {
        useConfigStore.getState().setUserProperties({ usageDays: 8 });
        expect(useConfigStore.getState().userSegment).toBe('regular');

        useConfigStore.getState().setUserProperties({ usageDays: 35 });
        expect(useConfigStore.getState().userSegment).toBe('power');
    });

    it('canShowNotification respects global enable/disable', () => {
        const store = useConfigStore.getState();
        const mockNotification = {
            id: 'test-1',
            type: 'info' as const,
            title: 'Test',
            message: 'Test message'
        };

        // Should be able to show by default
        expect(store.canShowNotification(mockNotification)).toBe(true);

        // Disable globally
        store.updateNotificationConfig({ enabled: false });
        expect(useConfigStore.getState().canShowNotification(mockNotification)).toBe(false);
    });
});
