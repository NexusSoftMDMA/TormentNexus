import { describe, it, expect, beforeEach } from 'vitest';
import { debateTemplateManager, DebateTemplateManagerService } from '../debate-template-manager.js';

describe('DebateTemplateManagerService', () => {
  let manager: DebateTemplateManagerService;

  beforeEach(() => {
    manager = new DebateTemplateManagerService();
  });

  describe('built-in templates', () => {
    it('should initialize with built-in templates', () => {
      const templates = manager.getAllTemplates();
      expect(templates.length).toBeGreaterThan(0);
    });

    it('should have code review template', () => {
      const template = manager.getTemplateByName('Code Review');
      expect(template).toBeDefined();
      expect(template?.category).toBe('code-review');
    });

    it('should have security audit template', () => {
      const template = manager.getTemplateByName('Security Audit');
      expect(template).toBeDefined();
      expect(template?.category).toBe('security-audit');
      expect(template?.config.consensusThreshold).toBe(0.8);
    });

    it('should have architecture review template', () => {
      const template = manager.getTemplateByName('Architecture Review');
      expect(template).toBeDefined();
      expect(template?.config.consensusMode).toBe('supermajority');
    });

    it('should not allow modifying built-in templates', () => {
      const builtIn = manager.getBuiltInTemplates()[0];
      const updated = manager.updateTemplate(builtIn.id, { name: 'Modified' });
      expect(updated).toBeUndefined();
    });

    it('should not allow deleting built-in templates', () => {
      const builtIn = manager.getBuiltInTemplates()[0];
      const deleted = manager.deleteTemplate(builtIn.id);
      expect(deleted).toBe(false);
    });
  });

  describe('custom templates', () => {
    it('should create a custom template', () => {
      const template = manager.createTemplate({
        name: 'My Custom Review',
        description: 'Custom review template',
        category: 'custom',
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: ['code-review'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are a reviewer.',
          initialPrompt: 'Review: {task}',
          debatePrompt: 'Discuss.',
          votingPrompt: 'Vote.',
          customCriteria: ['Quality check'],
        },
        scoring: {
          weights: {
            correctness: 0.2,
            security: 0.2,
            performance: 0.2,
            maintainability: 0.2,
            testability: 0.1,
            documentation: 0.1,
          },
          passingThreshold: 0.7,
          criticalCriteria: [],
        },
      });

      expect(template.id).toMatch(/^template_/);
      expect(template.isBuiltIn).toBe(false);
      expect(template.usageCount).toBe(0);
    });

    it('should update custom template', () => {
      const template = manager.createTemplate({
        name: 'Original',
        description: '',
        category: 'custom',
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: [],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: '',
          initialPrompt: '{task}',
          debatePrompt: '',
          votingPrompt: '',
          customCriteria: [],
        },
        scoring: {
          weights: { correctness: 1, security: 0, performance: 0, maintainability: 0, testability: 0, documentation: 0 },
          passingThreshold: 0.5,
          criticalCriteria: [],
        },
      });

      const updated = manager.updateTemplate(template.id, { name: 'Updated' });
      expect(updated?.name).toBe('Updated');
    });

    it('should delete custom template', () => {
      const template = manager.createTemplate({
        name: 'ToDelete',
        description: '',
        category: 'custom',
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: [],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: '',
          initialPrompt: '{task}',
          debatePrompt: '',
          votingPrompt: '',
          customCriteria: [],
        },
        scoring: {
          weights: { correctness: 1, security: 0, performance: 0, maintainability: 0, testability: 0, documentation: 0 },
          passingThreshold: 0.5,
          criticalCriteria: [],
        },
      });

      const deleted = manager.deleteTemplate(template.id);
      expect(deleted).toBe(true);
      expect(manager.getTemplate(template.id)).toBeUndefined();
    });
  });

  describe('template filtering', () => {
    it('should get templates by category', () => {
      const securityTemplates = manager.getTemplatesByCategory('security-audit');
      expect(securityTemplates.length).toBeGreaterThan(0);
      expect(securityTemplates.every((t: { category: string }) => t.category === 'security-audit')).toBe(true);
    });

    it('should get built-in templates only', () => {
      const builtIn = manager.getBuiltInTemplates();
      expect(builtIn.every((t: { isBuiltIn: boolean }) => t.isBuiltIn)).toBe(true);
    });

    it('should get custom templates only', () => {
      manager.createTemplate({
        name: 'Custom',
        description: '',
        category: 'custom',
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: [],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: '',
          initialPrompt: '{task}',
          debatePrompt: '',
          votingPrompt: '',
          customCriteria: [],
        },
        scoring: {
          weights: { correctness: 1, security: 0, performance: 0, maintainability: 0, testability: 0, documentation: 0 },
          passingThreshold: 0.5,
          criticalCriteria: [],
        },
      });

      const custom = manager.getCustomTemplates();
      expect(custom.length).toBe(1);
      expect(custom[0].isBuiltIn).toBe(false);
    });
  });

  describe('template application', () => {
    it('should apply template to task', () => {
      const template = manager.getTemplateByName('Code Review')!;
      const application = manager.applyTemplate(template.id, 'Review this PR for user authentication');

      expect(application).toBeDefined();
      expect(application?.templateId).toBe(template.id);
      expect(application?.resolvedPrompts.initial).toContain('Review this PR for user authentication');
    });

    it('should increment usage count on application', () => {
      const template = manager.getTemplateByName('Code Review')!;
      const initialCount = template.usageCount;

      manager.applyTemplate(template.id, 'Task 1');
      manager.applyTemplate(template.id, 'Task 2');

      const updated = manager.getTemplate(template.id)!;
      expect(updated.usageCount).toBe(initialCount + 2);
    });

    it('should include criteria in voting prompt', () => {
      const template = manager.getTemplateByName('Security Audit')!;
      const application = manager.applyTemplate(template.id, 'Check for vulnerabilities');

      expect(application?.resolvedPrompts.voting).toContain('No SQL injection');
    });
  });

  describe('template matching', () => {
    it('should find best template for security task', () => {
      const template = manager.findBestTemplate('Check for security vulnerabilities in auth module');
      expect(template?.category).toBe('security-audit');
    });

    it('should find best template for architecture task', () => {
      const template = manager.findBestTemplate('Review the new microservices architecture design');
      expect(template?.category).toBe('architecture-review');
    });

    it('should find best template for performance task', () => {
      const template = manager.findBestTemplate('Optimize database query performance');
      expect(template?.category).toBe('performance-review');
    });

    it('should find best template for API task', () => {
      const template = manager.findBestTemplate('Review the new REST API endpoints');
      expect(template?.category).toBe('api-design');
    });

    it('should suggest multiple templates', () => {
      const suggestions = manager.suggestTemplates('Review code changes', 3);
      expect(suggestions.length).toBeLessThanOrEqual(3);
    });
  });

  describe('template cloning', () => {
    it('should clone a template', () => {
      const original = manager.getTemplateByName('Code Review')!;
      const cloned = manager.cloneTemplate(original.id, 'My Code Review');

      expect(cloned).toBeDefined();
      expect(cloned?.name).toBe('My Code Review');
      expect(cloned?.isBuiltIn).toBe(false);
      expect(cloned?.config.consensusMode).toBe(original.config.consensusMode);
    });
  });

  describe('statistics', () => {
    it('should return template stats', () => {
      const stats = manager.getTemplateStats();

      expect(stats.total).toBeGreaterThan(0);
      expect(stats.builtIn).toBeGreaterThan(0);
      expect(stats.byCategory).toBeDefined();
    });
  });

  describe('export/import', () => {
    it('should export a template', () => {
      const template = manager.getTemplateByName('Code Review')!;
      const exported = manager.exportTemplate(template.id);

      expect(exported).toBeDefined();
      expect(exported?.name).toBe('Code Review');
    });

    it('should export all custom templates', () => {
      manager.createTemplate({
        name: 'Custom 1',
        description: '',
        category: 'custom',
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: [],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: '',
          initialPrompt: '{task}',
          debatePrompt: '',
          votingPrompt: '',
          customCriteria: [],
        },
        scoring: {
          weights: { correctness: 1, security: 0, performance: 0, maintainability: 0, testability: 0, documentation: 0 },
          passingThreshold: 0.5,
          criticalCriteria: [],
        },
      });

      const exported = manager.exportAllCustomTemplates();
      expect(exported.length).toBe(1);
    });

    it('should import a template', () => {
      const original = manager.getTemplateByName('Code Review')!;
      const imported = manager.importTemplate({
        name: 'Imported Review',
        description: original.description,
        category: original.category,
        config: original.config,
        prompts: original.prompts,
        scoring: original.scoring,
        isBuiltIn: true,
      });

      expect(imported.name).toBe('Imported Review');
      expect(imported.isBuiltIn).toBe(false);
    });
  });

  describe('reset', () => {
    it('should reset to built-in templates only', () => {
      manager.createTemplate({
        name: 'Custom',
        description: '',
        category: 'custom',
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: [],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: '',
          initialPrompt: '{task}',
          debatePrompt: '',
          votingPrompt: '',
          customCriteria: [],
        },
        scoring: {
          weights: { correctness: 1, security: 0, performance: 0, maintainability: 0, testability: 0, documentation: 0 },
          passingThreshold: 0.5,
          criticalCriteria: [],
        },
      });

      manager.resetToBuiltIn();

      expect(manager.getCustomTemplates().length).toBe(0);
      expect(manager.getBuiltInTemplates().length).toBeGreaterThan(0);
    });
  });
});
