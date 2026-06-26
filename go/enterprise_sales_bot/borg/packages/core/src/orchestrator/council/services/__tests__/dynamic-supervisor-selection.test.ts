import { describe, test, expect, beforeEach } from 'vitest';
import { DynamicSupervisorSelection } from '../dynamic-supervisor-selection.js';

describe('DynamicSupervisorSelection', () => {
  let selection: DynamicSupervisorSelection;

  beforeEach(() => {
    selection = new DynamicSupervisorSelection();
    selection.setAvailableSupervisors(['GPT-4', 'Claude', 'Gemini', 'DeepSeek', 'Grok', 'Qwen', 'Kimi']);
  });

  describe('isEnabled/setEnabled', () => {
    test('defaults to enabled', () => {
      expect(selection.isEnabled()).toBe(true);
    });

    test('can be disabled', () => {
      selection.setEnabled(false);
      expect(selection.isEnabled()).toBe(false);
    });
  });

  describe('supervisor profiles', () => {
    test('has default profiles loaded', () => {
      const profiles = selection.getAllProfiles();
      expect(profiles.length).toBeGreaterThan(0);
      expect(profiles.some(p => p.name === 'GPT-4')).toBe(true);
      expect(profiles.some(p => p.name === 'Claude')).toBe(true);
    });

    test('can get specific profile', () => {
      const profile = selection.getSupervisorProfile('Claude');
      expect(profile).toBeDefined();
      expect(profile?.provider).toBe('anthropic');
      expect(profile?.strengths).toContain('architecture');
    });

    test('can add new profile', () => {
      selection.addSupervisorProfile({
        name: 'TestModel',
        provider: 'test',
        strengths: ['general', 'testing'],
      });
      const profile = selection.getSupervisorProfile('TestModel');
      expect(profile).toBeDefined();
      expect(profile?.provider).toBe('test');
    });

    test('can remove profile', () => {
      selection.removeSupervisorProfile('Kimi');
      expect(selection.getSupervisorProfile('Kimi')).toBeUndefined();
    });
  });

  describe('team templates', () => {
    test('has default templates loaded', () => {
      const templates = selection.getAllTemplates();
      expect(templates.length).toBeGreaterThan(0);
      expect(templates.some(t => t.name === 'security-audit-team')).toBe(true);
    });

    test('can get specific template', () => {
      const template = selection.getTeamTemplate('security-audit-team');
      expect(template).toBeDefined();
      expect(template?.leadSupervisor).toBe('Claude');
      expect(template?.consensusMode).toBe('supermajority');
    });

    test('can add new template', () => {
      selection.addTeamTemplate({
        name: 'custom-team',
        description: 'Custom test team',
        taskTypes: ['general'],
        supervisors: ['GPT-4', 'Claude'],
        leadSupervisor: 'GPT-4',
        consensusMode: 'weighted',
      });
      const template = selection.getTeamTemplate('custom-team');
      expect(template).toBeDefined();
      expect(template?.supervisors).toEqual(['GPT-4', 'Claude']);
    });

    test('can update existing template', () => {
      selection.addTeamTemplate({
        name: 'security-audit-team',
        description: 'Updated team',
        taskTypes: ['security-audit'],
        supervisors: ['Claude', 'GPT-4'],
        leadSupervisor: 'Claude',
        consensusMode: 'unanimous',
      });
      const template = selection.getTeamTemplate('security-audit-team');
      expect(template?.consensusMode).toBe('unanimous');
    });

    test('can remove template', () => {
      selection.removeTeamTemplate('documentation-team');
      expect(selection.getTeamTemplate('documentation-team')).toBeUndefined();
    });
  });

  describe('detectTaskType', () => {
    test('detects security task', () => {
      const result = selection.detectTaskType({
        id: 'test-1',
        description: 'Review code for SQL injection vulnerabilities and XSS attacks',
        context: '',
        files: [],
      });
      expect(result.taskType).toBe('security-audit');
      expect(result.confidence).toBeGreaterThan(0);
    });

    test('detects UI design task', () => {
      const result = selection.detectTaskType({
        id: 'test-2',
        description: 'Update the button component styling and layout',
        context: '',
        files: ['Button.tsx', 'styles.css'],
      });
      expect(result.taskType).toBe('ui-design');
    });

    test('detects bug fix task', () => {
      const result = selection.detectTaskType({
        id: 'test-3',
        description: 'Fix the crash that occurs when user clicks submit',
        context: 'Error: undefined is not a function',
        files: [],
      });
      expect(result.taskType).toBe('bug-fix');
    });

    test('detects documentation task from file extension', () => {
      const result = selection.detectTaskType({
        id: 'test-4',
        description: 'Update the project information',
        context: '',
        files: ['README.md', 'CHANGELOG.md'],
      });
      expect(result.taskType).toBe('documentation');
    });

    test('detects testing task', () => {
      const result = selection.detectTaskType({
        id: 'test-5',
        description: 'Add unit tests for the new authentication module',
        context: '',
        files: ['auth.test.ts'],
      });
      expect(result.taskType).toBe('testing');
    });

    test('defaults to general for ambiguous tasks', () => {
      const result = selection.detectTaskType({
        id: 'test-6',
        description: 'Make some changes to the codebase',
        context: '',
        files: [],
      });
      expect(result.taskType).toBe('general');
    });
  });

  describe('selectTeam', () => {
    test('selects security team for security tasks', () => {
      const result = selection.selectTeam({
        id: 'test-1',
        description: 'Audit the authentication system for vulnerabilities',
        context: '',
        files: [],
      });
      expect(result.taskType).toBe('security-audit');
      expect(result.team).toContain('Claude');
      expect(result.consensusMode).toBe('supermajority');
    });

    test('selects architecture team for architecture tasks', () => {
      const result = selection.selectTeam({
        id: 'test-2',
        description: 'Design the microservice architecture for the new system',
        context: '',
        files: [],
      });
      expect(result.taskType).toBe('architecture');
      expect(result.leadSupervisor).toBe('Claude');
    });

    test('returns all supervisors when disabled', () => {
      selection.setEnabled(false);
      const result = selection.selectTeam({
        id: 'test-3',
        description: 'Fix security vulnerability',
        context: '',
        files: [],
      });
      expect(result.taskType).toBe('general');
      expect(result.team.length).toBe(7);
    });

    test('handles unavailable supervisors gracefully', () => {
      selection.setAvailableSupervisors(['GPT-4', 'Grok']);
      const result = selection.selectTeam({
        id: 'test-4',
        description: 'Review the UI components',
        context: '',
        files: [],
      });
      expect(result.team.every(s => ['GPT-4', 'Grok'].includes(s))).toBe(true);
    });

    test('falls back to available supervisors when template team unavailable', () => {
      selection.setAvailableSupervisors(['Qwen']);
      const result = selection.selectTeam({
        id: 'test-5',
        description: 'Audit for security issues',
        context: '',
        files: [],
      });
      expect(result.team).toContain('Qwen');
    });
  });

  describe('getStats', () => {
    test('returns correct statistics', () => {
      const stats = selection.getStats();
      expect(stats.enabled).toBe(true);
      expect(stats.profileCount).toBeGreaterThan(0);
      expect(stats.templateCount).toBeGreaterThan(0);
      expect(stats.availableSupervisors).toBe(7);
    });
  });
});
