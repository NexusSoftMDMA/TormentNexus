import { eq, desc } from "drizzle-orm";
import { randomUUID } from "node:crypto";
import { db } from "../index.js";
import { policiesTable } from "../mcp-admin-schema.js";
import { CreatePolicySchema, UpdatePolicySchema, PolicySchema } from "../../types/tormentnexus/policies.zod.js";
import { z } from "zod";

type Policy = z.infer<typeof PolicySchema>;
type PolicyRow = typeof policiesTable.$inferSelect;
type PolicyInsert = typeof policiesTable.$inferInsert;

export class PoliciesRepository {
    async findAll(): Promise<Policy[]> {
        const policies = await db.select().from(policiesTable).orderBy(desc(policiesTable.createdAt));
        return policies.map((policy) => this.mapToDomain(policy));
    }

    async findByUuid(uuid: string): Promise<Policy | undefined> {
        const [policy] = await db.select().from(policiesTable).where(eq(policiesTable.uuid, uuid));
        return policy ? this.mapToDomain(policy) : undefined;
    }

    async create(input: z.infer<typeof CreatePolicySchema>): Promise<Policy> {
        const payload: PolicyInsert = {
            uuid: randomUUID(),
            name: input.name,
            description: input.description ?? null,
            rules: input.rules,
        };

        const [policy] = await db.insert(policiesTable).values(payload).returning();
        return this.mapToDomain(policy);
    }

    async update(input: z.infer<typeof UpdatePolicySchema>): Promise<Policy> {
        const [policy] = await db.update(policiesTable)
            .set({
                ...(input.name && { name: input.name }),
                ...(input.description !== undefined && { description: input.description }),
                ...(input.rules && { rules: input.rules }),
                updatedAt: new Date(),
            })
            .where(eq(policiesTable.uuid, input.uuid))
            .returning();

        if (!policy) throw new Error("Policy not found");
        return this.mapToDomain(policy);
    }

    async delete(uuid: string): Promise<void> {
        await db.delete(policiesTable).where(eq(policiesTable.uuid, uuid));
    }

    private mapToDomain(dbPolicy: PolicyRow): Policy {
        return {
            uuid: dbPolicy.uuid,
            name: dbPolicy.name,
            description: dbPolicy.description,
            rules: PolicySchema.shape.rules.parse(dbPolicy.rules),
            createdAt: new Date(dbPolicy.createdAt),
            updatedAt: new Date(dbPolicy.updatedAt),
        };
    }
}

export const policiesRepository = new PoliciesRepository();
