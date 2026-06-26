import { db } from "../index.js";
import { workflowsTable } from "../tormentnexus-schema.js";
import { eq, desc } from "drizzle-orm";
import crypto from "crypto";

export class VisualWorkflowsRepository {
    async createWorkflow(data: { name: string; description?: string; nodes: any[]; edges: any[]; userId?: string }) {
        const id = crypto.randomUUID();
        const record = {
            id,
            name: data.name,
            description: data.description || null,
            nodes_json: data.nodes,
            edges_json: data.edges,
            user_id: data.userId || 'system',
        };
        await db.insert(workflowsTable).values(record);
        return record;
    }

    async updateWorkflow(id: string, data: { name?: string; description?: string; nodes?: any[]; edges?: any[] }) {
        const updates: any = { updated_at: new Date() };
        if (data.name !== undefined) updates.name = data.name;
        if (data.description !== undefined) updates.description = data.description;
        if (data.nodes !== undefined) updates.nodes_json = data.nodes;
        if (data.edges !== undefined) updates.edges_json = data.edges;

        await db.update(workflowsTable).set(updates).where(eq(workflowsTable.id, id));
    }

    async getWorkflow(id: string) {
        const results = await db.select().from(workflowsTable).where(eq(workflowsTable.id, id)).limit(1);
        return results[0] || null;
    }

    async listWorkflows(userId?: string) {
        const query = db.select().from(workflowsTable);
        if (userId) {
            query.where(eq(workflowsTable.user_id, userId));
        }
        return await query.orderBy(desc(workflowsTable.updated_at));
    }

    async deleteWorkflow(id: string) {
        await db.delete(workflowsTable).where(eq(workflowsTable.id, id));
    }
}

export const visualWorkflowsRepo = new VisualWorkflowsRepository();
