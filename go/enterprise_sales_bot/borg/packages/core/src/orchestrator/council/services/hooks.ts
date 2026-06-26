import type { Session, DevelopmentTask, CouncilDecision, Guidance } from './types.js';

export type HookPhase = 
  | 'pre-debate'
  | 'post-debate'
  | 'pre-guidance'
  | 'post-guidance'
  | 'on-error';

export interface HookContext {
  phase: HookPhase;
  session: Session;
  task?: DevelopmentTask;
  decision?: CouncilDecision;
  guidance?: Guidance;
  error?: Error;
}

export type HookHandler = (context: HookContext) => Promise<HookResult>;

export interface HookResult {
  continue: boolean;
  modifiedGuidance?: Guidance;
  modifiedDecision?: CouncilDecision;
  reason?: string;
}

interface RegisteredHook {
  id: string;
  phase: HookPhase;
  handler: HookHandler;
  priority: number;
}

class AutoContinueHooks {
  private hooks: RegisteredHook[] = [];
  private nextId = 1;

  register(phase: HookPhase, handler: HookHandler, priority: number = 0): string {
    const id = `hook_${this.nextId++}`;
    this.hooks.push({ id, phase, handler, priority });
    this.hooks.sort((a, b) => b.priority - a.priority);
    return id;
  }

  unregister(id: string): boolean {
    const idx = this.hooks.findIndex(h => h.id === id);
    if (idx === -1) return false;
    this.hooks.splice(idx, 1);
    return true;
  }

  async execute(context: HookContext): Promise<HookResult> {
    const phaseHooks = this.hooks.filter(h => h.phase === context.phase);
    
    let result: HookResult = { continue: true };
    
    for (const hook of phaseHooks) {
      try {
        const hookResult = await hook.handler(context);
        
        if (!hookResult.continue) {
          return hookResult;
        }
        
        if (hookResult.modifiedGuidance) {
          result.modifiedGuidance = hookResult.modifiedGuidance;
        }
        if (hookResult.modifiedDecision) {
          result.modifiedDecision = hookResult.modifiedDecision;
        }
      } catch (err) {
        console.error(`[Hooks] Error in hook ${hook.id}:`, err);
      }
    }
    
    return result;
  }

  getRegisteredHooks(): Array<{ id: string; phase: HookPhase; priority: number }> {
    return this.hooks.map(h => ({ id: h.id, phase: h.phase, priority: h.priority }));
  }

  clear(): void {
    this.hooks = [];
  }
}

export const autoContinueHooks = new AutoContinueHooks();

export function registerPreDebateHook(handler: HookHandler, priority?: number): string {
  return autoContinueHooks.register('pre-debate', handler, priority);
}

export function registerPostDebateHook(handler: HookHandler, priority?: number): string {
  return autoContinueHooks.register('post-debate', handler, priority);
}

export function registerPreGuidanceHook(handler: HookHandler, priority?: number): string {
  return autoContinueHooks.register('pre-guidance', handler, priority);
}

export function registerPostGuidanceHook(handler: HookHandler, priority?: number): string {
  return autoContinueHooks.register('post-guidance', handler, priority);
}

export function registerErrorHook(handler: HookHandler, priority?: number): string {
  return autoContinueHooks.register('on-error', handler, priority);
}
