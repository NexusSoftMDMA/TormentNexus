import { z } from 'zod';
import { t, publicProcedure } from '../../lib/trpc-core.js';
import { rotationRoomService } from '../../orchestrator/council/services/rotation-room.js';

const roleSchema = z.enum(['planner', 'implementer', 'tester']);

const participantSchema = z.object({
  id: z.string().min(1),
  name: z.string().min(1),
  provider: z.string().min(1),
  model: z.string().min(1),
});

const supervisorSchema = z.object({
  name: z.string().min(1).optional(),
  provider: z.string().min(1),
  model: z.string().min(1),
  completionCriteria: z.string().min(1),
  evaluationMode: z.enum(['manual', 'after_turn', 'after_cycle']).optional(),
  completionAction: z.enum(['pause', 'complete']).optional(),
  enabled: z.boolean().optional(),
  maxTranscriptMessages: z.number().int().min(1).max(100).optional(),
  prompt: z.string().min(1).optional(),
  temperature: z.number().min(0).max(2).optional(),
});

export const rotationRouter = t.router({
  list: publicProcedure.query(() => {
    return rotationRoomService.listRooms();
  }),

  get: publicProcedure.input(z.object({
    roomId: z.string().min(1),
  })).query(({ input }) => {
    return rotationRoomService.getRoom(input.roomId);
  }),

  create: publicProcedure.input(z.object({
    title: z.string().min(1),
    objective: z.string().min(1),
    sharedContext: z.string().optional(),
    mode: z.enum(['plan', 'execute']).optional(),
    roleOrder: z.array(roleSchema).min(1).optional(),
    participants: z.array(participantSchema).min(2),
    supervisor: supervisorSchema.optional(),
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.createRoom(input);
  }),

  addParticipant: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    participant: participantSchema,
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.addParticipant(input);
  }),

  postMessage: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    participantId: z.string().min(1),
    content: z.string().min(1),
    kind: z.enum(['system', 'discussion', 'handoff', 'artifact']).optional(),
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.postMessage(input);
  }),

  setAgreement: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    participantId: z.string().min(1),
    agrees: z.boolean(),
    note: z.string().optional(),
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.setAgreement(input);
  }),

  advanceTurn: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    summary: z.string().optional(),
    checkpoint: z.string().optional(),
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.advanceTurn(input);
  }),

  configureSupervisor: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    supervisor: supervisorSchema,
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.configureSupervisor(input);
  }),

  runSupervisorCheck: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    trigger: z.enum(['manual', 'turn_advanced', 'cycle_completed']).optional(),
  })).mutation(({ input }) => {
    // @ts-expect-error -- type mismatch from inferred schema types
    return rotationRoomService.runSupervisorCheck(input);
  }),

  updateSharedContext: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    sharedContext: z.string(),
  })).mutation(({ input }) => {
    return rotationRoomService.updateSharedContext(input.roomId, input.sharedContext);
  }),

  pause: publicProcedure.input(z.object({
    roomId: z.string().min(1),
  })).mutation(({ input }) => {
    return rotationRoomService.pauseRoom(input.roomId);
  }),

  resume: publicProcedure.input(z.object({
    roomId: z.string().min(1),
  })).mutation(({ input }) => {
    return rotationRoomService.resumeRoom(input.roomId);
  }),

  startExecution: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    checkpoint: z.string().optional(),
  })).mutation(({ input }) => {
    return rotationRoomService.startExecution(input.roomId, input.checkpoint);
  }),

  complete: publicProcedure.input(z.object({
    roomId: z.string().min(1),
    summary: z.string().optional(),
  })).mutation(({ input }) => {
    return rotationRoomService.completeRoom(input.roomId, input.summary);
  }),
});
