import type { ConsensusMode, DevelopmentTask } from './types.js';

type TaskType = 
  | 'security-audit'
  | 'ui-design'
  | 'api-design'
  | 'performance'
  | 'refactoring'
  | 'bug-fix'
  | 'testing'
  | 'documentation'
  | 'architecture'
  | 'code-review'
  | 'general';

interface SupervisorProfile {
  name: string;
  provider: string;
  strengths: TaskType[];
  weaknesses?: TaskType[];
  specializations?: string[];
  preferredForLeadOn?: TaskType[];
}

interface TeamTemplate {
  name: string;
  description: string;
  taskTypes: TaskType[];
  supervisors: string[];
  leadSupervisor?: string;
  consensusMode?: ConsensusMode;
  minSupervisors?: number;
}

interface TeamSelectionResult {
  team: string[];
  leadSupervisor?: string;
  consensusMode: ConsensusMode;
  reasoning: string;
  taskType: TaskType;
  confidence: number;
}

const DEFAULT_PROFILES: SupervisorProfile[] = [
  {
    name: 'GPT-4',
    provider: 'openai',
    strengths: ['code-review', 'bug-fix', 'refactoring', 'general', 'documentation'],
    weaknesses: ['ui-design'],
    specializations: ['logic analysis', 'algorithm optimization'],
    preferredForLeadOn: ['code-review', 'refactoring'],
  },
  {
    name: 'Claude',
    provider: 'anthropic',
    strengths: ['architecture', 'security-audit', 'documentation', 'code-review', 'api-design'],
    weaknesses: [],
    specializations: ['system design', 'security analysis', 'technical writing'],
    preferredForLeadOn: ['architecture', 'security-audit', 'api-design'],
  },
  {
    name: 'Gemini',
    provider: 'google',
    strengths: ['ui-design', 'performance', 'testing', 'general'],
    weaknesses: ['security-audit'],
    specializations: ['creative solutions', 'multimodal understanding'],
    preferredForLeadOn: ['ui-design'],
  },
  {
    name: 'DeepSeek',
    provider: 'deepseek',
    strengths: ['bug-fix', 'refactoring', 'code-review', 'performance'],
    weaknesses: ['documentation', 'ui-design'],
    specializations: ['deep code analysis', 'optimization'],
    preferredForLeadOn: ['performance'],
  },
  {
    name: 'Grok',
    provider: 'grok',
    strengths: ['general', 'bug-fix', 'testing'],
    weaknesses: ['architecture'],
    specializations: ['quick analysis', 'pragmatic solutions'],
    preferredForLeadOn: [],
  },
  {
    name: 'Qwen',
    provider: 'qwen',
    strengths: ['code-review', 'refactoring', 'documentation', 'general'],
    weaknesses: ['security-audit'],
    specializations: ['multilingual code', 'comprehensive analysis'],
    preferredForLeadOn: ['documentation'],
  },
  {
    name: 'Kimi',
    provider: 'kimi',
    strengths: ['documentation', 'code-review', 'general'],
    weaknesses: ['architecture', 'security-audit'],
    specializations: ['long context analysis'],
    preferredForLeadOn: [],
  },
];

const DEFAULT_TEMPLATES: TeamTemplate[] = [
  {
    name: 'security-audit-team',
    description: 'Specialized team for security reviews and vulnerability analysis',
    taskTypes: ['security-audit'],
    supervisors: ['Claude', 'GPT-4', 'DeepSeek'],
    leadSupervisor: 'Claude',
    consensusMode: 'supermajority',
    minSupervisors: 2,
  },
  {
    name: 'ui-design-team',
    description: 'Team focused on UI/UX design decisions',
    taskTypes: ['ui-design'],
    supervisors: ['Gemini', 'GPT-4', 'Claude'],
    leadSupervisor: 'Gemini',
    consensusMode: 'weighted',
    minSupervisors: 2,
  },
  {
    name: 'architecture-team',
    description: 'Team for high-level architecture and system design',
    taskTypes: ['architecture', 'api-design'],
    supervisors: ['Claude', 'GPT-4', 'DeepSeek'],
    leadSupervisor: 'Claude',
    consensusMode: 'hybrid-ceo-majority',
    minSupervisors: 2,
  },
  {
    name: 'performance-team',
    description: 'Team specialized in performance optimization',
    taskTypes: ['performance'],
    supervisors: ['DeepSeek', 'GPT-4', 'Gemini'],
    leadSupervisor: 'DeepSeek',
    consensusMode: 'weighted',
    minSupervisors: 2,
  },
  {
    name: 'code-quality-team',
    description: 'Team for code reviews and refactoring',
    taskTypes: ['code-review', 'refactoring'],
    supervisors: ['GPT-4', 'Claude', 'DeepSeek', 'Qwen'],
    leadSupervisor: 'GPT-4',
    consensusMode: 'weighted',
    minSupervisors: 2,
  },
  {
    name: 'bug-fix-team',
    description: 'Team for debugging and fixing issues',
    taskTypes: ['bug-fix'],
    supervisors: ['GPT-4', 'DeepSeek', 'Grok'],
    leadSupervisor: 'GPT-4',
    consensusMode: 'simple-majority',
    minSupervisors: 2,
  },
  {
    name: 'testing-team',
    description: 'Team for test design and coverage analysis',
    taskTypes: ['testing'],
    supervisors: ['Gemini', 'GPT-4', 'Grok'],
    leadSupervisor: 'Gemini',
    consensusMode: 'weighted',
    minSupervisors: 2,
  },
  {
    name: 'documentation-team',
    description: 'Team for documentation and technical writing',
    taskTypes: ['documentation'],
    supervisors: ['Claude', 'Qwen', 'Kimi'],
    leadSupervisor: 'Claude',
    consensusMode: 'simple-majority',
    minSupervisors: 2,
  },
  {
    name: 'general-team',
    description: 'Balanced team for general development tasks',
    taskTypes: ['general'],
    supervisors: ['GPT-4', 'Claude', 'Gemini'],
    leadSupervisor: 'GPT-4',
    consensusMode: 'weighted',
    minSupervisors: 2,
  },
];

const TASK_KEYWORDS: Record<TaskType, string[]> = {
  'security-audit': ['security', 'vulnerability', 'audit', 'cve', 'injection', 'xss', 'csrf', 'auth', 'encryption', 'permission', 'access control'],
  'ui-design': ['ui', 'ux', 'design', 'layout', 'component', 'style', 'css', 'responsive', 'animation', 'visual', 'frontend'],
  'api-design': ['api', 'endpoint', 'rest', 'graphql', 'schema', 'request', 'response', 'route', 'contract'],
  'performance': ['performance', 'optimize', 'slow', 'memory', 'cpu', 'latency', 'throughput', 'cache', 'profil'],
  'refactoring': ['refactor', 'clean', 'restructure', 'simplify', 'extract', 'rename', 'modular', 'dry', 'solid'],
  'bug-fix': ['bug', 'fix', 'error', 'issue', 'crash', 'broken', 'fail', 'exception', 'debug'],
  'testing': ['test', 'spec', 'coverage', 'unit', 'integration', 'e2e', 'mock', 'assert', 'expect'],
  'documentation': ['doc', 'readme', 'comment', 'jsdoc', 'explain', 'describe', 'guide', 'tutorial'],
  'architecture': ['architecture', 'design pattern', 'structure', 'system', 'scalab', 'microservice', 'monolith', 'dependency'],
  'code-review': ['review', 'pr', 'pull request', 'feedback', 'approve', 'suggest', 'improve'],
  'general': [],
};

export class DynamicSupervisorSelection {
  private profiles: Map<string, SupervisorProfile> = new Map();
  private templates: TeamTemplate[] = [];
  private availableSupervisors: Set<string> = new Set();
  private enabled = true;

  constructor() {
    DEFAULT_PROFILES.forEach(p => this.profiles.set(p.name, p));
    this.templates = [...DEFAULT_TEMPLATES];
  }

  setEnabled(enabled: boolean): void {
    this.enabled = enabled;
  }

  isEnabled(): boolean {
    return this.enabled;
  }

  setAvailableSupervisors(names: string[]): void {
    this.availableSupervisors = new Set(names);
  }

  addSupervisorProfile(profile: SupervisorProfile): void {
    this.profiles.set(profile.name, profile);
  }

  removeSupervisorProfile(name: string): void {
    this.profiles.delete(name);
  }

  getSupervisorProfile(name: string): SupervisorProfile | undefined {
    return this.profiles.get(name);
  }

  getAllProfiles(): SupervisorProfile[] {
    return Array.from(this.profiles.values());
  }

  addTeamTemplate(template: TeamTemplate): void {
    const existingIndex = this.templates.findIndex(t => t.name === template.name);
    if (existingIndex >= 0) {
      this.templates[existingIndex] = template;
    } else {
      this.templates.push(template);
    }
  }

  removeTeamTemplate(name: string): void {
    this.templates = this.templates.filter(t => t.name !== name);
  }

  getTeamTemplate(name: string): TeamTemplate | undefined {
    return this.templates.find(t => t.name === name);
  }

  getAllTemplates(): TeamTemplate[] {
    return [...this.templates];
  }

  detectTaskType(task: DevelopmentTask): { taskType: TaskType; confidence: number; reasoning: string } {
    const text = `${task.description} ${task.context}`.toLowerCase();
    const scores: Record<TaskType, number> = {
      'security-audit': 0,
      'ui-design': 0,
      'api-design': 0,
      'performance': 0,
      'refactoring': 0,
      'bug-fix': 0,
      'testing': 0,
      'documentation': 0,
      'architecture': 0,
      'code-review': 0,
      'general': 0.1,
    };

    const matchedKeywords: Record<TaskType, string[]> = {} as Record<TaskType, string[]>;
    for (const taskType of Object.keys(TASK_KEYWORDS) as TaskType[]) {
      matchedKeywords[taskType] = [];
    }

    for (const [taskType, keywords] of Object.entries(TASK_KEYWORDS)) {
      for (const keyword of keywords) {
        if (text.includes(keyword)) {
          scores[taskType as TaskType] += 1;
          matchedKeywords[taskType as TaskType].push(keyword);
        }
      }
    }

    const fileExtensions = task.files.map(f => f.split('.').pop()?.toLowerCase() || '');
    if (fileExtensions.some(ext => ['css', 'scss', 'less', 'tsx', 'jsx'].includes(ext))) {
      scores['ui-design'] += 0.5;
    }
    if (fileExtensions.some(ext => ['test.ts', 'spec.ts', 'test.js', 'spec.js'].some(s => ext.includes(s)))) {
      scores['testing'] += 1;
    }
    if (fileExtensions.some(ext => ['md', 'mdx', 'rst'].includes(ext))) {
      scores['documentation'] += 1;
    }

    let maxScore = 0;
    let detectedType: TaskType = 'general';
    for (const [taskType, score] of Object.entries(scores)) {
      if (score > maxScore) {
        maxScore = score;
        detectedType = taskType as TaskType;
      }
    }

    const totalKeywords = Object.values(TASK_KEYWORDS).flat().length;
    const confidence = Math.min(0.95, maxScore / 5);

    const matched = matchedKeywords[detectedType];
    const reasoning = matched.length > 0
      ? `Detected ${detectedType} based on keywords: ${matched.slice(0, 5).join(', ')}`
      : `Defaulting to ${detectedType} (no strong signals detected)`;

    return { taskType: detectedType, confidence, reasoning };
  }

  selectTeam(task: DevelopmentTask): TeamSelectionResult {
    if (!this.enabled) {
      return {
        team: Array.from(this.availableSupervisors),
        consensusMode: 'weighted',
        reasoning: 'Dynamic selection disabled, using all available supervisors',
        taskType: 'general',
        confidence: 1.0,
      };
    }

    const { taskType, confidence: detectionConfidence, reasoning: detectionReasoning } = this.detectTaskType(task);

    const matchingTemplate = this.templates.find(t => t.taskTypes.includes(taskType));

    if (matchingTemplate) {
      const availableFromTemplate = matchingTemplate.supervisors.filter(s => 
        this.availableSupervisors.has(s)
      );

      if (availableFromTemplate.length >= (matchingTemplate.minSupervisors || 2)) {
        const leadAvailable = matchingTemplate.leadSupervisor && 
          this.availableSupervisors.has(matchingTemplate.leadSupervisor);

        return {
          team: availableFromTemplate,
          leadSupervisor: leadAvailable ? matchingTemplate.leadSupervisor : availableFromTemplate[0],
          consensusMode: matchingTemplate.consensusMode || 'weighted',
          reasoning: `Using ${matchingTemplate.name}: ${matchingTemplate.description}. ${detectionReasoning}`,
          taskType,
          confidence: detectionConfidence,
        };
      }
    }

    const scoredSupervisors = this.scoreSupervisorsForTask(taskType);
    const selectedTeam = scoredSupervisors
      .filter(s => this.availableSupervisors.has(s.name))
      .slice(0, 3)
      .map(s => s.name);

    if (selectedTeam.length === 0) {
      return {
        team: Array.from(this.availableSupervisors),
        consensusMode: 'weighted',
        reasoning: `No specialized supervisors available for ${taskType}, using all available`,
        taskType,
        confidence: detectionConfidence * 0.5,
      };
    }

    const leadSupervisor = this.selectLeadForTask(taskType, selectedTeam);

    return {
      team: selectedTeam,
      leadSupervisor,
      consensusMode: this.selectConsensusMode(taskType),
      reasoning: `Dynamically selected team for ${taskType}. ${detectionReasoning}`,
      taskType,
      confidence: detectionConfidence,
    };
  }

  private scoreSupervisorsForTask(taskType: TaskType): Array<{ name: string; score: number }> {
    const scores: Array<{ name: string; score: number }> = [];

    for (const profile of Array.from(this.profiles.values())) {
      let score = 0;

      if (profile.strengths.includes(taskType)) {
        score += 2;
      }
      if (profile.weaknesses?.includes(taskType)) {
        score -= 1;
      }
      if (profile.preferredForLeadOn?.includes(taskType)) {
        score += 0.5;
      }

      scores.push({ name: profile.name, score });
    }

    return scores.sort((a, b) => b.score - a.score);
  }

  private selectLeadForTask(taskType: TaskType, team: string[]): string | undefined {
    for (const name of team) {
      const profile = this.profiles.get(name);
      if (profile?.preferredForLeadOn?.includes(taskType)) {
        return name;
      }
    }

    for (const name of team) {
      const profile = this.profiles.get(name);
      if (profile?.strengths.includes(taskType)) {
        return name;
      }
    }

    return team[0];
  }

  private selectConsensusMode(taskType: TaskType): ConsensusMode {
    switch (taskType) {
      case 'security-audit':
        return 'supermajority';
      case 'architecture':
      case 'api-design':
        return 'hybrid-ceo-majority';
      case 'bug-fix':
      case 'documentation':
        return 'simple-majority';
      default:
        return 'weighted';
    }
  }

  getStats(): {
    profileCount: number;
    templateCount: number;
    availableSupervisors: number;
    enabled: boolean;
  } {
    return {
      profileCount: this.profiles.size,
      templateCount: this.templates.length,
      availableSupervisors: this.availableSupervisors.size,
      enabled: this.enabled,
    };
  }
}

export const dynamicSupervisorSelection = new DynamicSupervisorSelection();
