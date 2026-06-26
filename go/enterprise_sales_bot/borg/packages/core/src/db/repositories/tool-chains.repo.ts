import { db } from "../index.js";
import { eq, desc } from "drizzle-orm";
import {
    toolChainsTable,
    toolChainStepsTable,
    toolAliasesTable,
} from "../mcp-tool-chain-schema.js";

// === Types ===
export type ToolChain = typeof toolChainsTable.$inferSelect;
export type NewToolChain = typeof toolChainsTable.$inferInsert;

export type ToolChainStep = typeof toolChainStepsTable.$inferSelect;
export type NewToolChainStep = typeof toolChainStepsTable.$inferInsert;

export type ToolAlias = typeof toolAliasesTable.$inferSelect;
export type NewToolAlias = typeof toolAliasesTable.$inferInsert;

export type PopulatedToolChain = ToolChain & {
    steps: ToolChainStep[];
};

export const toolChainsRepository = {
    // --- Chains ---
    async createChain(chain: NewToolChain, steps: NewToolChainStep[]): Promise<PopulatedToolChain> {
        return await db.transaction(async (tx) => {
            const [savedChain] = await tx.insert(toolChainsTable).values(chain).returning();
            
            let savedSteps: ToolChainStep[] = [];
            if (steps.length > 0) {
                const stepsWithChainId = steps.map(s => ({ ...s, chainId: savedChain.id }));
                savedSteps = await tx.insert(toolChainStepsTable).values(stepsWithChainId).returning();
            }
            
            return {
                ...savedChain,
                steps: savedSteps.sort((a, b) => a.stepOrder - b.stepOrder),
            };
        });
    },

    async getAllChains(): Promise<PopulatedToolChain[]> {
        const chains = await db.select().from(toolChainsTable).orderBy(desc(toolChainsTable.createdAt));
        const steps = await db.select().from(toolChainStepsTable).orderBy(toolChainStepsTable.stepOrder);
        
        return chains.map(chain => ({
            ...chain,
            steps: steps.filter(s => s.chainId === chain.id),
        }));
    },

    async getChainById(id: string): Promise<PopulatedToolChain | null> {
        const [chain] = await db.select().from(toolChainsTable).where(eq(toolChainsTable.id, id));
        if (!chain) return null;
        
        const steps = await db.select()
            .from(toolChainStepsTable)
            .where(eq(toolChainStepsTable.chainId, id))
            .orderBy(toolChainStepsTable.stepOrder);
            
        return {
            ...chain,
            steps,
        };
    },

    async deleteChain(id: string): Promise<boolean> {
        const res = await db.delete(toolChainsTable).where(eq(toolChainsTable.id, id)).returning();
        return res.length > 0;
    },

    // --- Aliases ---
    async upsertAlias(alias: NewToolAlias): Promise<ToolAlias> {
        const [saved] = await db.insert(toolAliasesTable).values(alias)
            .onConflictDoUpdate({
                target: toolAliasesTable.alias,
                set: {
                    targetTool: alias.targetTool,
                    description: alias.description,
                    defaultArguments: alias.defaultArguments,
                }
            })
            .returning();
        return saved;
    },

    async getAliases(): Promise<ToolAlias[]> {
        return db.select().from(toolAliasesTable).orderBy(desc(toolAliasesTable.createdAt));
    },

    async getAliasById(aliasName: string): Promise<ToolAlias | null> {
        const [alias] = await db.select().from(toolAliasesTable).where(eq(toolAliasesTable.alias, aliasName));
        return alias || null;
    },

    async deleteAlias(aliasName: string): Promise<boolean> {
        const res = await db.delete(toolAliasesTable).where(eq(toolAliasesTable.alias, aliasName)).returning();
        return res.length > 0;
    }
};
