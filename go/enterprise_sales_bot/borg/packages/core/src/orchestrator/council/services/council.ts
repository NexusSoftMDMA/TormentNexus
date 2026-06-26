import type { Supervisor, CouncilConfig, CouncilDecision, DevelopmentTask, Message, ConsensusMode, Vote, TaskType } from './types.js';
import { metrics } from './metrics.js';
import { dynamicSupervisorSelection } from './dynamic-supervisor-selection.js';
import { debateHistory } from './debate-history.js';
import { supervisorAnalytics } from './supervisor-analytics.js';
import type { TaskPlan, SubTask } from './types.js';

interface ConsensusModeHandler {
  (votes: Vote[], config: CouncilConfig, leadVote?: Vote): { approved: boolean; reasoning: string };
}

export class SupervisorCouncil {
  private supervisors: Supervisor[] = [];
  private supervisorWeights: Map<string, number> = new Map();
  private config: CouncilConfig;
  private fallbackIndex = 0;

  private consensusHandlers: Record<ConsensusMode, ConsensusModeHandler> = {
    'simple-majority': this.handleSimpleMajority.bind(this),
    'supermajority': this.handleSupermajority.bind(this),
    'unanimous': this.handleUnanimous.bind(this),
    'weighted': this.handleWeighted.bind(this),
    'ceo-override': this.handleCeoOverride.bind(this),
    'ceo-veto': this.handleCeoVeto.bind(this),
    'hybrid-ceo-majority': this.handleHybridCeoMajority.bind(this),
    'ranked-choice': this.handleRankedChoice.bind(this),
  };

  constructor(config: CouncilConfig) {
    this.config = config;
  }

  addSupervisor(supervisor: Supervisor, weight: number = 1.0) {
    this.supervisors.push(supervisor);
    this.supervisorWeights.set(supervisor.name, weight);
  }

  setSupervisorWeight(name: string, weight: number): void {
    this.supervisorWeights.set(name, Math.max(0, Math.min(2, weight))); // Clamp between 0 and 2
  }

  getSupervisorWeight(name: string): number {
    return this.supervisorWeights.get(name) ?? 1.0;
  }

  setLeadSupervisor(name: string): void {
    this.config.leadSupervisor = name;
  }

  getLeadSupervisor(): string | undefined {
    return this.config.leadSupervisor;
  }

  setFallbackChain(supervisors: string[]): void {
    this.config.fallbackSupervisors = supervisors;
  }

  getFallbackChain(): string[] {
    return this.config.fallbackSupervisors ?? [];
  }

  setConsensusMode(mode: ConsensusMode): void {
    this.config.consensusMode = mode;
  }

  getConsensusMode(): ConsensusMode {
    return this.config.consensusMode ?? 'weighted';
  }

  async getAvailableSupervisors(): Promise<Supervisor[]> {
    const results = await Promise.all(
      this.supervisors.map(async (supervisor) => ({
        supervisor,
        available: await supervisor.isAvailable(),
      }))
    );
    return results.filter((r) => r.available).map((r) => r.supervisor);
  }

  /**
   * Get the next available supervisor from the fallback chain
   */
  async getNextFallbackSupervisor(): Promise<Supervisor | null> {
    const fallbackChain = this.config.fallbackSupervisors ?? [];
    
    while (this.fallbackIndex < fallbackChain.length) {
      const name = fallbackChain[this.fallbackIndex];
      const supervisor = this.supervisors.find(s => s.name === name);
      this.fallbackIndex++;
      
      if (supervisor && await supervisor.isAvailable()) {
        return supervisor;
      }
    }
    
    // Reset fallback index when exhausted
    this.fallbackIndex = 0;
    return null;
  }

  /**
   * Try to get a response with fallback chain support
   */
  async chatWithFallback(messages: Message[]): Promise<{ response: string; supervisor: string } | null> {
    // Try lead supervisor first
    if (this.config.leadSupervisor) {
      const lead = this.supervisors.find(s => s.name === this.config.leadSupervisor);
      if (lead && await lead.isAvailable()) {
        try {
          const response = await lead.chat(messages);
          return { response, supervisor: lead.name };
        } catch {
          // Fall through to fallback chain
        }
      }
    }

    // Try fallback chain
    let fallback = await this.getNextFallbackSupervisor();
    while (fallback) {
      try {
        const response = await fallback.chat(messages);
        return { response, supervisor: fallback.name };
      } catch {
        fallback = await this.getNextFallbackSupervisor();
      }
    }

    // Try any available supervisor
    const available = await this.getAvailableSupervisors();
    for (const supervisor of available) {
      try {
        const response = await supervisor.chat(messages);
        return { response, supervisor: supervisor.name };
      } catch {
        continue;
      }
    }

    return null;
  }

  async debate(task: DevelopmentTask): Promise<CouncilDecision> {
    const startTime = Date.now();
    
    dynamicSupervisorSelection.setAvailableSupervisors(this.supervisors.map(s => s.name));
    
    let supervisorsToUse = this.supervisors;
    let consensusModeToUse = this.config.consensusMode ?? 'weighted';
    let leadSupervisorToUse = this.config.leadSupervisor;
    let dynamicSelectionInfo: string | undefined;
    let dynamicSelectionData: { enabled: boolean; taskType?: TaskType; confidence?: number } | undefined;
    
    if (dynamicSupervisorSelection.isEnabled()) {
      const selection = dynamicSupervisorSelection.selectTeam(task);
      
      const selectedSupervisors = this.supervisors.filter(s => selection.team.includes(s.name));
      if (selectedSupervisors.length > 0) {
        supervisorsToUse = selectedSupervisors;
        if (!leadSupervisorToUse) {
          leadSupervisorToUse = selection.leadSupervisor;
        }
        dynamicSelectionInfo = `**Dynamic Selection:** ${selection.reasoning} (confidence: ${(selection.confidence * 100).toFixed(0)}%)`;
        dynamicSelectionData = {
          enabled: true,
          taskType: selection.taskType as TaskType,
          confidence: selection.confidence,
        };
      }
    }
    
    const available: Supervisor[] = [];
    for (const supervisor of supervisorsToUse) {
      if (await supervisor.isAvailable()) {
        available.push(supervisor);
      }
    }
    
    if (available.length === 0) {
      metrics.recordDebate(Date.now() - startTime, 0, true);
      return {
        approved: true,
        consensus: 1.0,
        weightedConsensus: 1.0,
        votes: [],
        reasoning: 'No supervisors available - auto-approving',
        dissent: [],
      };
    }

    const rounds = this.config.debateRounds || 2;
    const votes: Vote[] = [];

    const taskContext: Message = {
      role: 'user',
      content: this.formatTaskForDebate(task),
    };

    const initialOpinions = await Promise.all(
      available.map(async (supervisor) => {
        try {
          const response = await supervisor.chat([taskContext]);
          return `**${supervisor.name}**: ${response}`;
        } catch (error) {
          return `**${supervisor.name}**: [Unable to provide opinion]`;
        }
      })
    );

    let debateContext = taskContext.content + '\n\n**Initial Opinions:**\n' + initialOpinions.join('\n\n');
    
    for (let round = 2; round <= rounds; round++) {
      const roundOpinions = await Promise.all(
        available.map(async (supervisor) => {
          try {
            const message: Message = {
              role: 'user',
              content: debateContext + '\n\nConsidering the above opinions, provide your refined assessment.',
            };
            
            const response = await supervisor.chat([message]);
            return `**${supervisor.name}**: ${response}`;
          } catch {
            return null;
          }
        })
      );
      
      const validOpinions = roundOpinions.filter((o): o is string => o !== null);
      debateContext += '\n\n**Round ' + round + ' Opinions:**\n' + validOpinions.join('\n\n');
    }

    const voteResults = await Promise.all(
      available.map(async (supervisor) => {
        const voteStartTime = Date.now();
        try {
          const votePrompt: Message = {
            role: 'user',
            content: debateContext + 
              '\n\nBased on all discussions, provide your FINAL VOTE:\n' +
              '1. Vote: APPROVE or REJECT\n' +
              '2. Confidence: A number between 0.0 and 1.0 (how confident are you in this decision?)\n' +
              '3. Brief reasoning (2-3 sentences)\n\n' +
              'Format:\nVOTE: [APPROVE/REJECT]\nCONFIDENCE: [0.0-1.0]\nREASONING: [your reasoning]',
          };
          
          const response = await supervisor.chat([votePrompt]);
          const responseTimeMs = Date.now() - voteStartTime;
          const approved = this.parseVote(response);
          const confidence = this.parseConfidence(response);
          const weight = this.getSupervisorWeight(supervisor.name);
          
          return {
            supervisor: supervisor.name,
            approved,
            confidence,
            weight,
            comment: response,
            responseTimeMs,
          };
        } catch {
          return {
            supervisor: supervisor.name,
            approved: false,
            confidence: 0.5,
            weight: this.getSupervisorWeight(supervisor.name),
            comment: 'Failed to vote',
            responseTimeMs: Date.now() - voteStartTime,
          };
        }
      })
    );

    votes.push(...voteResults);

    // Calculate simple consensus (backward compatible)
    const approvals = votes.filter(v => v.approved).length;
    const consensus = votes.length > 0 ? approvals / votes.length : 0;

    // Calculate weighted consensus
    const weightedConsensus = this.calculateWeightedConsensus(votes);
    
    // Track strong dissent (rejections with high confidence)
    const dissent = this.extractDissent(votes);

    // Get lead supervisor's vote if applicable
    const leadVote = leadSupervisorToUse 
      ? votes.find(v => v.supervisor === leadSupervisorToUse)
      : undefined;

    const mode = consensusModeToUse;
    const handler = this.consensusHandlers[mode];
    const { approved, reasoning: modeReasoning } = handler(votes, this.config, leadVote);

    metrics.recordDebate(Date.now() - startTime, rounds, approved);

    // Record to analytics
    const durationMs = Date.now() - startTime;
    supervisorAnalytics.recordDebateOutcome(
      task.id,
      task.description.substring(0, 100),
      mode,
      approved ? 'approved' : 'rejected',
      available.map(s => s.name),
      rounds,
      durationMs
    );

    for (const vote of votes) {
      // Find the response time from the augmented vote results
      const result = voteResults.find(v => v.supervisor === vote.supervisor);
      const responseTimeMs = (result as any).responseTimeMs || 0;

      supervisorAnalytics.recordVote(
        vote.supervisor,
        task.id,
        vote.approved ? 'approve' : 'reject',
        vote.confidence,
        responseTimeMs,
        0, // tokensUsed not tracked yet
        approved ? 'approve' : 'reject'
      );
    }

    const decision: CouncilDecision = {
      approved,
      consensus,
      weightedConsensus,
      votes,
      reasoning: this.generateConsensusReasoning(votes, approved, weightedConsensus, dissent, mode, modeReasoning, dynamicSelectionInfo),
      dissent,
    };

    if (debateHistory.isEnabled()) {
      debateHistory.saveDebate(task, decision, {
        debateRounds: rounds,
        consensusMode: mode,
        leadSupervisor: leadSupervisorToUse,
        dynamicSelection: dynamicSelectionData,
        durationMs,
      });
    }

    return decision;
  }

  // ============ Consensus Mode Handlers ============

  private handleSimpleMajority(votes: Vote[], config: CouncilConfig): { approved: boolean; reasoning: string } {
    const approvals = votes.filter(v => v.approved).length;
    const consensus = votes.length > 0 ? approvals / votes.length : 0;
    const threshold = config.consensusThreshold ?? 0.5;
    const approved = consensus >= threshold;
    return {
      approved,
      reasoning: `Simple majority: ${approvals}/${votes.length} (${(consensus * 100).toFixed(0)}%) approved (threshold: ${(threshold * 100).toFixed(0)}%)`,
    };
  }

  private handleSupermajority(votes: Vote[]): { approved: boolean; reasoning: string } {
    const approvals = votes.filter(v => v.approved).length;
    const threshold = votes.length * 0.667;
    const approved = approvals >= threshold;
    return {
      approved,
      reasoning: `Supermajority: ${approvals}/${votes.length} approved (need >=${Math.ceil(threshold)}, 66.7%)`,
    };
  }

  private handleUnanimous(votes: Vote[]): { approved: boolean; reasoning: string } {
    const approvals = votes.filter(v => v.approved).length;
    const approved = approvals === votes.length;
    return {
      approved,
      reasoning: `Unanimous: ${approvals}/${votes.length} approved (need ${votes.length}/${votes.length})`,
    };
  }

  private handleWeighted(votes: Vote[], config: CouncilConfig): { approved: boolean; reasoning: string } {
    const weightedConsensus = this.calculateWeightedConsensus(votes);
    const threshold = config.consensusThreshold ?? 0.5;
    const approved = weightedConsensus >= threshold;
    return {
      approved,
      reasoning: `Weighted consensus: ${(weightedConsensus * 100).toFixed(1)}% (threshold: ${(threshold * 100).toFixed(1)}%)`,
    };
  }

  private handleCeoOverride(votes: Vote[], config: CouncilConfig, leadVote?: Vote): { approved: boolean; reasoning: string } {
    if (!leadVote) {
      // Fall back to weighted if no lead supervisor vote
      return this.handleWeighted(votes, config);
    }
    
    return {
      approved: leadVote.approved,
      reasoning: `CEO Override: ${config.leadSupervisor} ${leadVote.approved ? 'APPROVED' : 'REJECTED'} (confidence: ${leadVote.confidence.toFixed(2)})`,
    };
  }

  private handleCeoVeto(votes: Vote[], config: CouncilConfig, leadVote?: Vote): { approved: boolean; reasoning: string } {
    // First calculate majority
    const approvals = votes.filter(v => v.approved).length;
    const majorityApproved = approvals > votes.length / 2;
    
    // CEO can only veto (reject), not force approve
    if (leadVote && !leadVote.approved && leadVote.confidence >= 0.7) {
      return {
        approved: false,
        reasoning: `CEO Veto: ${config.leadSupervisor} VETOED with high confidence (${leadVote.confidence.toFixed(2)}). Majority was ${majorityApproved ? 'in favor' : 'against'}.`,
      };
    }
    
    return {
      approved: majorityApproved,
      reasoning: `CEO Veto (not used): Majority ${majorityApproved ? 'approved' : 'rejected'} (${approvals}/${votes.length}). ${config.leadSupervisor || 'Lead'} did not veto.`,
    };
  }

  private handleHybridCeoMajority(votes: Vote[], config: CouncilConfig, leadVote?: Vote): { approved: boolean; reasoning: string } {
    const approvals = votes.filter(v => v.approved).length;
    const rejections = votes.length - approvals;
    
    // Clear majority
    if (approvals > rejections + 1) {
      return {
        approved: true,
        reasoning: `Hybrid CEO-Majority: Clear majority approved (${approvals}/${votes.length})`,
      };
    }
    
    if (rejections > approvals + 1) {
      return {
        approved: false,
        reasoning: `Hybrid CEO-Majority: Clear majority rejected (${rejections}/${votes.length} against)`,
      };
    }
    
    // Tie or close - CEO decides
    if (leadVote) {
      return {
        approved: leadVote.approved,
        reasoning: `Hybrid CEO-Majority: Tie/close vote (${approvals}-${rejections}), ${config.leadSupervisor} breaks tie: ${leadVote.approved ? 'APPROVED' : 'REJECTED'}`,
      };
    }
    
    // No CEO vote, default to slight majority or reject
    return {
      approved: approvals >= rejections,
      reasoning: `Hybrid CEO-Majority: Tie/close vote (${approvals}-${rejections}), no CEO to break tie, defaulting to ${approvals >= rejections ? 'approve' : 'reject'}`,
    };
  }

  private handleRankedChoice(votes: Vote[]): { approved: boolean; reasoning: string } {
    // Ranked choice using confidence as "rank" - higher confidence = stronger preference
    // Calculate weighted score for approve vs reject
    let approveScore = 0;
    let rejectScore = 0;
    
    for (const vote of votes) {
      const score = vote.weight * vote.confidence;
      if (vote.approved) {
        approveScore += score;
      } else {
        rejectScore += score;
      }
    }
    
    const approved = approveScore >= rejectScore;
    return {
      approved,
      reasoning: `Ranked Choice: Approve score ${approveScore.toFixed(2)} vs Reject score ${rejectScore.toFixed(2)}`,
    };
  }

  // ============ Helper Methods ============

  private calculateWeightedConsensus(votes: Vote[]): number {
    if (votes.length === 0) return 0;

    let weightedApprovals = 0;
    let totalWeight = 0;

    for (const vote of votes) {
      const effectiveWeight = vote.weight * vote.confidence;
      totalWeight += vote.weight;
      if (vote.approved) {
        weightedApprovals += effectiveWeight;
      }
    }

    return totalWeight > 0 ? weightedApprovals / totalWeight : 0;
  }

  private extractDissent(votes: Vote[]): string[] {
    const dissent: string[] = [];
    
    for (const vote of votes) {
      if (!vote.approved && vote.confidence > 0.7) {
        const shortComment = vote.comment.length > 300 
          ? vote.comment.substring(0, 300) + '...' 
          : vote.comment;
        dissent.push(`${vote.supervisor} (confidence: ${vote.confidence.toFixed(2)}): ${shortComment}`);
      }
    }
    
    return dissent;
  }

  private parseConfidence(response: string): number {
    const confidenceMatch = response.match(/CONFIDENCE:\s*([\d.]+)/i);
    if (confidenceMatch) {
      const value = parseFloat(confidenceMatch[1]);
      if (!isNaN(value)) {
        return value > 1 ? Math.min(1, value / 100) : Math.max(0, Math.min(1, value));
      }
    }

    const altMatch = response.match(/confidence[:\s]+(\d+(?:\.\d+)?)/i);
    if (altMatch) {
      const value = parseFloat(altMatch[1]);
      if (!isNaN(value)) {
        return value > 1 ? Math.min(1, value / 100) : Math.max(0, Math.min(1, value));
      }
    }

    return 0.7;
  }

  private formatTaskForDebate(task: DevelopmentTask): string {
    return `
# Development Task Review

**Task ID**: ${task.id}
**Description**: ${task.description}

**Context**: 
${task.context}

**Files Affected**: 
${task.files.join('\n')}

**Your Role**: 
As a supervisor, review this development task and provide your expert opinion on:
1. Code quality and best practices
2. Potential issues or risks
3. Suggestions for improvement
4. Whether this task should be approved to proceed

Be thorough but concise in your analysis.
`.trim();
  }

  private parseVote(response: string): boolean {
    const normalized = response.toUpperCase();
    
    if (normalized.includes('VOTE: APPROVE') || normalized.includes('VOTE:APPROVE')) {
      return true;
    }
    if (normalized.includes('VOTE: REJECT') || normalized.includes('VOTE:REJECT')) {
      return false;
    }
    
    const approveMatch = /\b(APPROVE|APPROVED|ACCEPT|ACCEPTED|LGTM)\b/.test(normalized);
    const rejectMatch = /\b(REJECT|REJECTED|DENY|DENIED)\b/.test(normalized);
    
    if (approveMatch && !rejectMatch) return true;
    if (rejectMatch && !approveMatch) return false;
    
    return false;
  }

  private generateConsensusReasoning(
    votes: Vote[], 
    approved: boolean,
    weightedConsensus: number,
    dissent: string[],
    mode: ConsensusMode,
    modeReasoning: string,
    dynamicSelectionInfo?: string
  ): string {
    const approvals = votes.filter(v => v.approved).length;
    const avgConfidence = votes.length > 0 
      ? votes.reduce((sum, v) => sum + v.confidence, 0) / votes.length 
      : 0;
    
    let reasoning = '';
    
    if (dynamicSelectionInfo) {
      reasoning += dynamicSelectionInfo + '\n\n';
    }
    
    reasoning += `After ${votes.length} supervisor votes using **${mode}** mode, `;
    
    if (approved) {
      reasoning += `the council has reached consensus to APPROVE this task.`;
    } else {
      reasoning += `the council has decided to REJECT this task.`;
    }

    reasoning += `\n\n**Consensus Mode Decision:**\n${modeReasoning}`;

    reasoning += `\n\n**Voting Summary:**`;
    reasoning += `\n- Simple consensus: ${approvals}/${votes.length} approved (${(approvals/votes.length*100).toFixed(0)}%)`;
    reasoning += `\n- Weighted consensus: ${(weightedConsensus * 100).toFixed(1)}%`;
    reasoning += `\n- Average confidence: ${(avgConfidence * 100).toFixed(1)}%`;
    
    if (this.config.leadSupervisor) {
      const leadVote = votes.find(v => v.supervisor === this.config.leadSupervisor);
      if (leadVote) {
        reasoning += `\n- Lead supervisor (${this.config.leadSupervisor}): ${leadVote.approved ? 'APPROVED' : 'REJECTED'} (confidence: ${leadVote.confidence.toFixed(2)})`;
      }
    }
    
    if (dissent.length > 0) {
      reasoning += `\n\n**Strong Dissenting Opinions (${dissent.length}):**`;
      for (const d of dissent) {
        reasoning += `\n- ${d}`;
      }
    }

    reasoning += '\n\n**Individual Votes:**';
    
    for (const vote of votes) {
      const status = vote.approved ? '✅' : '❌';
      const isLead = vote.supervisor === this.config.leadSupervisor ? ' [LEAD]' : '';
      const comment = vote.comment.length > 150 ? vote.comment.substring(0, 150) + '...' : vote.comment;
      reasoning += `\n- ${status} ${vote.supervisor}${isLead} (weight: ${vote.weight.toFixed(1)}, confidence: ${vote.confidence.toFixed(2)}): ${comment}`;
    }
    
    return reasoning;
  }

  getSupervisors(): Supervisor[] {
    return [...this.supervisors];
  }

  clearSupervisors(): void {
    this.supervisors = [];
    this.supervisorWeights.clear();
  }

  setDebateRounds(rounds: number): void {
    this.config.debateRounds = rounds;
  }

  setConsensusThreshold(threshold: number): void {
    this.config.consensusThreshold = threshold;
  }

  setWeightedVoting(enabled: boolean): void {
    this.config.weightedVoting = enabled;
    const currentMode = this.config.consensusMode ?? 'weighted';
    if (!enabled && (currentMode === 'weighted' || !this.config.consensusMode)) {
      this.config.consensusMode = 'simple-majority';
    } else if (enabled && currentMode === 'simple-majority') {
      this.config.consensusMode = 'weighted';
    }
  }

  async planTask(task: DevelopmentTask): Promise<TaskPlan> {
    const available = await this.getAvailableSupervisors();
    if (available.length === 0) throw new Error('No supervisors available for planning');

    // Use lead supervisor or best available for planning
    const planner = this.config.leadSupervisor
      ? available.find(s => s.name === this.config.leadSupervisor) || available[0]
      : available[0];

    const planningPrompt: Message = {
      role: 'user',
      content: `
# Task Decomposition

**Goal**: Break down the following high-level task into smaller, executable subtasks.

**Task Description**: ${task.description}
**Context**: ${task.context || 'No additional context'}

**Instructions**:
1. Analyze the task requirements.
2. Identify independent components or sequential steps.
3. Assign each subtask to the most suitable AI CLI tool if applicable:
   - 'opencode': Best for multi-file editing and workspace management.
   - 'gemini': Best for research, complex reasoning, and multimodal tasks.
   - 'copilot': Best for quick code suggestions, shell commands, and explaining existing logic.
4. Generate a JSON list of subtasks. Each subtask must have:
   - id: string (unique)
   - title: string
   - description: string (detailed instructions)
   - dependencies: string[] (ids of tasks that must finish first)
   - preferredCLI: 'opencode' | 'gemini' | 'copilot' | null (optional)

**Output Format**:
You MUST return ONLY a JSON object with this structure:
{
  "reasoning": "Brief explanation of the plan...",
  "subtasks": [
    { "id": "1", "title": "...", "description": "...", "dependencies": [], "preferredCLI": "opencode" }
  ]
}
`
    };

    try {
      const response = await planner.chat([planningPrompt]);
      // Extract JSON from response (handle markdown code blocks)
      const jsonMatch = response.match(/\{[\s\S]*\}/);
      if (!jsonMatch) throw new Error('No JSON found in planning response');

      const plan = JSON.parse(jsonMatch[0]);
      return {
        originalTaskId: task.id,
        subtasks: plan.subtasks.map((t: any) => ({ ...t, status: 'pending' })),
        reasoning: plan.reasoning
      };
    } catch (error) {
      console.error('Planning failed:', error);
      // Fallback: Return original task as single subtask
      return {
        originalTaskId: task.id,
        reasoning: 'Automatic decomposition failed, treating as single task.',
        subtasks: [{
          id: `${task.id}-1`,
          title: 'Execute Task',
          description: task.description,
          dependencies: [],
          status: 'pending'
        }]
      };
    }
  }

  getConfig(): CouncilConfig {
    return { ...this.config };
  }
}

export const council = new SupervisorCouncil({
  supervisors: [],
  debateRounds: 2,
  consensusThreshold: 0.7,
  enabled: true,
  weightedVoting: true,
  consensusMode: 'weighted',
});
