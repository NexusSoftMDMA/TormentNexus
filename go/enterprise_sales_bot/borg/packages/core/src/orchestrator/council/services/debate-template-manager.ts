import { EventEmitter } from 'events';
import type { ConsensusMode, TaskType } from './types.js';

export interface DebateTemplate {
  id: string;
  name: string;
  description: string;
  category: TemplateCategory;
  config: TemplateConfig;
  prompts: TemplatePrompts;
  scoring: ScoringConfig;
  createdAt: Date;
  updatedAt: Date;
  isBuiltIn: boolean;
  usageCount: number;
}

export type TemplateCategory = 
  | 'code-review'
  | 'security-audit'
  | 'architecture-review'
  | 'performance-review'
  | 'api-design'
  | 'testing-strategy'
  | 'documentation'
  | 'refactoring'
  | 'custom';

export interface TemplateConfig {
  consensusMode: ConsensusMode;
  debateRounds: number;
  consensusThreshold: number;
  timeoutMs?: number;
  requiredSupervisors: string[];
  optionalSupervisors: string[];
  minSupervisors: number;
  maxSupervisors: number;
  taskTypes: TaskType[];
  autoSelectTeam: boolean;
}

export interface TemplatePrompts {
  systemPrompt: string;
  initialPrompt: string;
  debatePrompt: string;
  votingPrompt: string;
  customCriteria: string[];
}

export interface ScoringConfig {
  weights: {
    correctness: number;
    security: number;
    performance: number;
    maintainability: number;
    testability: number;
    documentation: number;
  };
  passingThreshold: number;
  criticalCriteria: string[];
}

export interface TemplateApplication {
  templateId: string;
  taskDescription: string;
  resolvedConfig: TemplateConfig;
  resolvedPrompts: {
    system: string;
    initial: string;
    debate: string;
    voting: string;
  };
}

export class DebateTemplateManagerService extends EventEmitter {
  private templates: Map<string, DebateTemplate> = new Map();

  constructor() {
    super();
    this.initializeBuiltInTemplates();
  }

  private initializeBuiltInTemplates(): void {
    const builtInTemplates: Omit<DebateTemplate, 'id' | 'createdAt' | 'updatedAt' | 'usageCount'>[] = [
      {
        name: 'Code Review',
        description: 'Standard code review focusing on quality, maintainability, and best practices',
        category: 'code-review',
        isBuiltIn: true,
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 5,
          taskTypes: ['code-review', 'refactoring', 'bug-fix'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are a senior software engineer conducting a thorough code review.',
          initialPrompt: 'Review the following code changes for quality, correctness, and adherence to best practices:\n\n{task}',
          debatePrompt: 'Consider the previous opinions and refine your assessment. Focus on areas of disagreement.',
          votingPrompt: 'Based on all discussions, provide your final vote on whether this code should be approved.',
          customCriteria: [
            'Code follows project conventions',
            'No obvious bugs or logic errors',
            'Proper error handling',
            'Adequate test coverage',
          ],
        },
        scoring: {
          weights: {
            correctness: 0.25,
            security: 0.15,
            performance: 0.15,
            maintainability: 0.2,
            testability: 0.15,
            documentation: 0.1,
          },
          passingThreshold: 0.7,
          criticalCriteria: ['No security vulnerabilities', 'No breaking changes'],
        },
      },
      {
        name: 'Security Audit',
        description: 'In-depth security review focusing on vulnerabilities and secure coding practices',
        category: 'security-audit',
        isBuiltIn: true,
        config: {
          consensusMode: 'ceo-veto',
          debateRounds: 3,
          consensusThreshold: 0.8,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 3,
          maxSupervisors: 5,
          taskTypes: ['security-audit'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are a security expert conducting a comprehensive security audit.',
          initialPrompt: 'Analyze the following code for security vulnerabilities, potential attack vectors, and compliance with secure coding standards:\n\n{task}',
          debatePrompt: 'Review other security assessments and identify any overlooked vulnerabilities or false positives.',
          votingPrompt: 'Based on the security analysis, determine if this code is safe for production.',
          customCriteria: [
            'No SQL injection vulnerabilities',
            'No XSS vulnerabilities',
            'Proper authentication/authorization',
            'Secure data handling',
            'No hardcoded secrets',
            'Input validation present',
          ],
        },
        scoring: {
          weights: {
            correctness: 0.1,
            security: 0.5,
            performance: 0.1,
            maintainability: 0.1,
            testability: 0.1,
            documentation: 0.1,
          },
          passingThreshold: 0.9,
          criticalCriteria: ['No critical vulnerabilities', 'No exposed secrets', 'Proper input validation'],
        },
      },
      {
        name: 'Architecture Review',
        description: 'High-level architecture and design review focusing on scalability and patterns',
        category: 'architecture-review',
        isBuiltIn: true,
        config: {
          consensusMode: 'supermajority',
          debateRounds: 3,
          consensusThreshold: 0.75,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 3,
          maxSupervisors: 5,
          taskTypes: ['architecture'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are a software architect reviewing system design and architecture decisions.',
          initialPrompt: 'Evaluate the following architecture/design decision for scalability, maintainability, and alignment with best practices:\n\n{task}',
          debatePrompt: 'Consider alternative architectural approaches and weigh trade-offs discussed by other reviewers.',
          votingPrompt: 'Based on architectural analysis, should this design be approved?',
          customCriteria: [
            'Follows SOLID principles',
            'Appropriate separation of concerns',
            'Scalable design',
            'Clear interfaces and contracts',
            'Proper error boundaries',
          ],
        },
        scoring: {
          weights: {
            correctness: 0.15,
            security: 0.15,
            performance: 0.2,
            maintainability: 0.25,
            testability: 0.15,
            documentation: 0.1,
          },
          passingThreshold: 0.75,
          criticalCriteria: ['No architectural anti-patterns', 'Scalability considered'],
        },
      },
      {
        name: 'Performance Review',
        description: 'Performance-focused review analyzing efficiency and resource usage',
        category: 'performance-review',
        isBuiltIn: true,
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 4,
          taskTypes: ['performance'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are a performance engineer analyzing code for efficiency and resource optimization.',
          initialPrompt: 'Analyze the following code for performance issues, inefficiencies, and optimization opportunities:\n\n{task}',
          debatePrompt: 'Discuss performance trade-offs and validate optimization suggestions from other reviewers.',
          votingPrompt: 'Based on performance analysis, does this code meet performance requirements?',
          customCriteria: [
            'No N+1 query problems',
            'Efficient algorithms used',
            'Proper caching strategy',
            'No memory leaks',
            'Async operations where appropriate',
          ],
        },
        scoring: {
          weights: {
            correctness: 0.15,
            security: 0.1,
            performance: 0.4,
            maintainability: 0.15,
            testability: 0.1,
            documentation: 0.1,
          },
          passingThreshold: 0.7,
          criticalCriteria: ['No O(n²) or worse algorithms for large datasets', 'No blocking operations in hot paths'],
        },
      },
      {
        name: 'API Design Review',
        description: 'API design review focusing on usability, consistency, and RESTful principles',
        category: 'api-design',
        isBuiltIn: true,
        config: {
          consensusMode: 'weighted',
          debateRounds: 2,
          consensusThreshold: 0.7,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 4,
          taskTypes: ['api-design'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are an API design expert reviewing endpoint design and contracts.',
          initialPrompt: 'Review the following API design for usability, consistency, and adherence to REST/GraphQL best practices:\n\n{task}',
          debatePrompt: 'Discuss API design choices and suggest improvements based on industry standards.',
          votingPrompt: 'Based on API design analysis, should this API design be approved?',
          customCriteria: [
            'Consistent naming conventions',
            'Proper HTTP methods used',
            'Appropriate status codes',
            'Clear error responses',
            'Proper versioning strategy',
            'Complete documentation',
          ],
        },
        scoring: {
          weights: {
            correctness: 0.2,
            security: 0.15,
            performance: 0.15,
            maintainability: 0.2,
            testability: 0.15,
            documentation: 0.15,
          },
          passingThreshold: 0.7,
          criticalCriteria: ['No breaking changes to existing endpoints', 'Proper authentication on protected routes'],
        },
      },
      {
        name: 'Quick Review',
        description: 'Fast single-round review for minor changes',
        category: 'code-review',
        isBuiltIn: true,
        config: {
          consensusMode: 'simple-majority',
          debateRounds: 1,
          consensusThreshold: 0.5,
          requiredSupervisors: [],
          optionalSupervisors: [],
          minSupervisors: 2,
          maxSupervisors: 3,
          taskTypes: ['code-review', 'bug-fix', 'documentation'],
          autoSelectTeam: true,
        },
        prompts: {
          systemPrompt: 'You are a developer conducting a quick review of a small change.',
          initialPrompt: 'Quickly review the following minor change:\n\n{task}',
          debatePrompt: 'Briefly note any concerns.',
          votingPrompt: 'Approve or reject this minor change.',
          customCriteria: ['No obvious bugs', 'Follows conventions'],
        },
        scoring: {
          weights: {
            correctness: 0.3,
            security: 0.2,
            performance: 0.1,
            maintainability: 0.2,
            testability: 0.1,
            documentation: 0.1,
          },
          passingThreshold: 0.6,
          criticalCriteria: ['No security issues'],
        },
      },
    ];

    for (const template of builtInTemplates) {
      const id = `builtin_${template.category}_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
      const now = new Date();
      this.templates.set(id, {
        ...template,
        id,
        createdAt: now,
        updatedAt: now,
        usageCount: 0,
      });
    }
  }

  // ============ Template CRUD ============

  createTemplate(template: Omit<DebateTemplate, 'id' | 'createdAt' | 'updatedAt' | 'usageCount' | 'isBuiltIn'>): DebateTemplate {
    const id = `template_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    const now = new Date();

    const newTemplate: DebateTemplate = {
      ...template,
      id,
      createdAt: now,
      updatedAt: now,
      isBuiltIn: false,
      usageCount: 0,
    };

    this.templates.set(id, newTemplate);
    this.emit('template:created', newTemplate);

    return newTemplate;
  }

  getTemplate(id: string): DebateTemplate | undefined {
    return this.templates.get(id);
  }

  getTemplateByName(name: string): DebateTemplate | undefined {
    for (const template of this.templates.values()) {
      if (template.name.toLowerCase() === name.toLowerCase()) {
        return template;
      }
    }
    return undefined;
  }

  getAllTemplates(): DebateTemplate[] {
    return Array.from(this.templates.values());
  }

  getTemplatesByCategory(category: TemplateCategory): DebateTemplate[] {
    return this.getAllTemplates().filter(t => t.category === category);
  }

  getBuiltInTemplates(): DebateTemplate[] {
    return this.getAllTemplates().filter(t => t.isBuiltIn);
  }

  getCustomTemplates(): DebateTemplate[] {
    return this.getAllTemplates().filter(t => !t.isBuiltIn);
  }

  updateTemplate(
    id: string,
    updates: Partial<Omit<DebateTemplate, 'id' | 'createdAt' | 'isBuiltIn' | 'usageCount'>>
  ): DebateTemplate | undefined {
    const template = this.templates.get(id);
    if (!template) return undefined;

    if (template.isBuiltIn) {
      this.emit('template:error', { id, error: 'Cannot modify built-in templates' });
      return undefined;
    }

    const updated: DebateTemplate = {
      ...template,
      ...updates,
      id: template.id,
      createdAt: template.createdAt,
      isBuiltIn: template.isBuiltIn,
      usageCount: template.usageCount,
      updatedAt: new Date(),
      config: updates.config ? { ...template.config, ...updates.config } : template.config,
      prompts: updates.prompts ? { ...template.prompts, ...updates.prompts } : template.prompts,
      scoring: updates.scoring ? { ...template.scoring, ...updates.scoring } : template.scoring,
    };

    this.templates.set(id, updated);
    this.emit('template:updated', updated);

    return updated;
  }

  deleteTemplate(id: string): boolean {
    const template = this.templates.get(id);
    if (!template) return false;

    if (template.isBuiltIn) {
      this.emit('template:error', { id, error: 'Cannot delete built-in templates' });
      return false;
    }

    this.templates.delete(id);
    this.emit('template:deleted', { id, name: template.name });

    return true;
  }

  // ============ Template Application ============

  applyTemplate(templateId: string, taskDescription: string): TemplateApplication | undefined {
    const template = this.templates.get(templateId);
    if (!template) return undefined;

    template.usageCount++;
    template.updatedAt = new Date();
    this.templates.set(templateId, template);

    const resolvedPrompts = {
      system: template.prompts.systemPrompt,
      initial: template.prompts.initialPrompt.replace('{task}', taskDescription),
      debate: template.prompts.debatePrompt,
      voting: this.buildVotingPrompt(template),
    };

    this.emit('template:applied', { templateId, taskDescription });

    return {
      templateId,
      taskDescription,
      resolvedConfig: { ...template.config },
      resolvedPrompts,
    };
  }

  private buildVotingPrompt(template: DebateTemplate): string {
    let prompt = template.prompts.votingPrompt;

    if (template.prompts.customCriteria.length > 0) {
      prompt += '\n\nEvaluation Criteria:\n';
      for (const criterion of template.prompts.customCriteria) {
        prompt += `- ${criterion}\n`;
      }
    }

    if (template.scoring.criticalCriteria.length > 0) {
      prompt += '\n\nCRITICAL (must pass):\n';
      for (const criterion of template.scoring.criticalCriteria) {
        prompt += `- ${criterion}\n`;
      }
    }

    return prompt;
  }

  // ============ Template Matching ============

  findBestTemplate(taskDescription: string, taskType?: TaskType): DebateTemplate | undefined {
    const keywords = taskDescription.toLowerCase();
    let bestMatch: DebateTemplate | undefined;
    let highestScore = 0;

    for (const template of this.templates.values()) {
      let score = 0;

      if (taskType && template.config.taskTypes.includes(taskType)) {
        score += 10;
      }

      if (keywords.includes('security') && template.category === 'security-audit') {
        score += 5;
      }
      if (keywords.includes('architecture') && template.category === 'architecture-review') {
        score += 5;
      }
      if (keywords.includes('performance') && template.category === 'performance-review') {
        score += 5;
      }
      if (keywords.includes('api') && template.category === 'api-design') {
        score += 5;
      }
      if ((keywords.includes('review') || keywords.includes('refactor')) && template.category === 'code-review') {
        score += 3;
      }

      score += template.usageCount * 0.1;

      if (score > highestScore) {
        highestScore = score;
        bestMatch = template;
      }
    }

    return bestMatch;
  }

  suggestTemplates(taskDescription: string, limit: number = 3): DebateTemplate[] {
    const keywords = taskDescription.toLowerCase();
    const scored: Array<{ template: DebateTemplate; score: number }> = [];

    for (const template of this.templates.values()) {
      let score = 0;

      const categoryKeywords: Record<TemplateCategory, string[]> = {
        'code-review': ['review', 'code', 'pr', 'pull request', 'change'],
        'security-audit': ['security', 'vulnerability', 'audit', 'penetration', 'attack'],
        'architecture-review': ['architecture', 'design', 'system', 'scalability', 'pattern'],
        'performance-review': ['performance', 'speed', 'optimization', 'slow', 'memory'],
        'api-design': ['api', 'endpoint', 'rest', 'graphql', 'contract'],
        'testing-strategy': ['test', 'testing', 'coverage', 'unit', 'integration'],
        'documentation': ['doc', 'documentation', 'readme', 'comment'],
        'refactoring': ['refactor', 'cleanup', 'tech debt', 'rewrite'],
        'custom': [],
      };

      const relevantKeywords = categoryKeywords[template.category] || [];
      for (const kw of relevantKeywords) {
        if (keywords.includes(kw)) {
          score += 2;
        }
      }

      score += template.usageCount * 0.05;

      scored.push({ template, score });
    }

    return scored
      .sort((a, b) => b.score - a.score)
      .slice(0, limit)
      .map(s => s.template);
  }

  // ============ Template Cloning ============

  cloneTemplate(id: string, newName: string): DebateTemplate | undefined {
    const source = this.templates.get(id);
    if (!source) return undefined;

    return this.createTemplate({
      name: newName,
      description: `Clone of ${source.name}`,
      category: source.category,
      config: { ...source.config },
      prompts: { ...source.prompts },
      scoring: { ...source.scoring },
    });
  }

  // ============ Statistics ============

  getTemplateStats(): {
    total: number;
    builtIn: number;
    custom: number;
    byCategory: Record<TemplateCategory, number>;
    mostUsed: Array<{ id: string; name: string; usageCount: number }>;
  } {
    const templates = this.getAllTemplates();
    const byCategory: Record<TemplateCategory, number> = {
      'code-review': 0,
      'security-audit': 0,
      'architecture-review': 0,
      'performance-review': 0,
      'api-design': 0,
      'testing-strategy': 0,
      'documentation': 0,
      'refactoring': 0,
      'custom': 0,
    };

    for (const template of templates) {
      byCategory[template.category]++;
    }

    const mostUsed = [...templates]
      .sort((a, b) => b.usageCount - a.usageCount)
      .slice(0, 5)
      .map(t => ({ id: t.id, name: t.name, usageCount: t.usageCount }));

    return {
      total: templates.length,
      builtIn: templates.filter(t => t.isBuiltIn).length,
      custom: templates.filter(t => !t.isBuiltIn).length,
      byCategory,
      mostUsed,
    };
  }

  // ============ Export/Import ============

  exportTemplate(id: string): DebateTemplate | undefined {
    return this.templates.get(id);
  }

  exportAllCustomTemplates(): DebateTemplate[] {
    return this.getCustomTemplates();
  }

  importTemplate(template: Omit<DebateTemplate, 'id' | 'createdAt' | 'updatedAt' | 'usageCount'>): DebateTemplate {
    return this.createTemplate({
      name: template.name,
      description: template.description,
      category: template.category,
      config: template.config,
      prompts: template.prompts,
      scoring: template.scoring,
    });
  }

  // ============ Cleanup ============

  resetToBuiltIn(): void {
    for (const [id, template] of this.templates.entries()) {
      if (!template.isBuiltIn) {
        this.templates.delete(id);
      }
    }
    this.emit('templates:reset');
  }
}

export const debateTemplateManager = new DebateTemplateManagerService();
