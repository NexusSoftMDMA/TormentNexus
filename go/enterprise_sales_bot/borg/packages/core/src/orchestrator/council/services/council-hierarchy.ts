import { SupervisorCouncil } from './council.js';
import { loadConfig } from './config.js';
import { createSupervisors } from '../supervisors/index.js';
import type { SpecializedCouncilConfig, DevelopmentTask, CouncilDecision, TaskType } from './types.js';

export interface SpecializedCouncil {
  id: string;
  name: string;
  description: string;
  specialties: TaskType[];
  council: SupervisorCouncil;
}

class CouncilHierarchyService {
  private supremeCouncil!: SupervisorCouncil;
  private specializedCouncils: Map<string, SpecializedCouncil> = new Map();

  constructor() {
    // Supreme Council is initialized in its own module, 
    // but we can wrap it or reference it here.
    // For now, we'll re-initialize based on config to support hierarchy.
  }

  async initialize(supremeCouncil: SupervisorCouncil): Promise<void> {
    this.supremeCouncil = supremeCouncil;
    const config = loadConfig();

    if (config.council.specializedCouncils) {
      for (const subConfig of config.council.specializedCouncils) {
        this.registerSpecializedCouncil(subConfig);
      }
    }
    console.log(`[CouncilHierarchy] Initialized with ${this.specializedCouncils.size} specialized councils.`);
  }

  registerSpecializedCouncil(config: SpecializedCouncilConfig): void {
    const council = new SupervisorCouncil(config);
    
    // Add supervisors to the sub-council
    const supervisors = createSupervisors(config.supervisors);
    for (const supervisor of supervisors) {
      council.addSupervisor(supervisor);
    }

    this.specializedCouncils.set(config.id, {
      id: config.id,
      name: config.name,
      description: config.description,
      specialties: config.specialties,
      council,
    });
  }

  getSpecializedCouncil(id: string): SpecializedCouncil | undefined {
    return this.specializedCouncils.get(id);
  }

  /**
   * Routes a task to the most appropriate council.
   * If a specialized council matches the task type, it gets the task.
   * Otherwise, it goes to the Supreme Council.
   */
  async routeTask(task: DevelopmentTask, taskType?: TaskType): Promise<CouncilDecision> {
    if (taskType) {
      for (const sub of this.specializedCouncils.values()) {
        if (sub.specialties.includes(taskType)) {
          console.log(`[CouncilHierarchy] Routing task to specialized council: ${sub.name} (${sub.id})`);
          return sub.council.debate(task);
        }
      }
    }

    console.log('[CouncilHierarchy] Routing task to Supreme Council');
    return this.supremeCouncil.debate(task);
  }

  getAllSpecializedCouncils(): SpecializedCouncil[] {
    return Array.from(this.specializedCouncils.values());
  }
}

export const councilHierarchy = new CouncilHierarchyService();
