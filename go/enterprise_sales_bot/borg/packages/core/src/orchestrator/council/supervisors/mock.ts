import type { Supervisor, SupervisorConfig, Message } from './types.js';

export class MockSupervisor implements Supervisor {
  name: string;
  provider: string;

  private delay: number;
  private shouldApprove: boolean;

  constructor(
    config: SupervisorConfig,
    options: { delay?: number; shouldApprove?: boolean } = {}
  ) {
    this.name = config.name;
    this.provider = config.provider;
    this.delay = options.delay ?? 500;
    this.shouldApprove = options.shouldApprove ?? true;
  }

  async chat(messages: Message[]): Promise<string> {
    await new Promise((resolve) => setTimeout(resolve, this.delay));

    const lastMessage = messages[messages.length - 1];
    const content = lastMessage?.content || '';

    if (content.includes('FINAL VOTE') || content.includes('APPROVE or REJECT')) {
      return this.generateVoteResponse();
    }

    return this.generateReviewResponse(content);
  }

  async isAvailable(): Promise<boolean> {
    return true;
  }

  private generateVoteResponse(): string {
    if (this.shouldApprove) {
      return `VOTE: APPROVE

REASONING: As a mock supervisor, I've reviewed the proposed changes and found them to be well-structured and following best practices. The implementation appears solid and ready for integration.`;
    }

    return `VOTE: REJECT

REASONING: As a mock supervisor, I've identified some concerns with the proposed changes that should be addressed before proceeding. Consider reviewing the implementation for potential improvements.`;
  }

  private generateReviewResponse(context: string): string {
    const isTask = context.includes('Development Task') || context.includes('Task ID');
    
    if (isTask) {
      return `**Mock Review by ${this.name}**

I've analyzed the development task and here are my observations:

1. **Code Quality**: The proposed changes appear to follow standard coding conventions.

2. **Potential Issues**: No critical issues identified in this mock review.

3. **Suggestions**: 
   - Consider adding unit tests for new functionality
   - Ensure error handling is comprehensive

4. **Overall Assessment**: The task looks reasonable and well-scoped.

This is a simulated response for testing purposes.`;
    }

    return `**${this.name} Response**

I've reviewed the provided context and have the following thoughts:

The approach seems reasonable. Key considerations:
- Implementation follows common patterns
- No obvious security concerns
- Performance should be acceptable

This is a mock response for testing without API keys.`;
  }
}
