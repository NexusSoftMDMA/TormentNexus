
import { CouncilService } from '../services/CouncilService.js';

async function verifyCouncil() {
    console.log("🏛️ Verifying Council Service...");
    const council = new CouncilService();

    // 1. Start Session
    const session = council.startSession("Should we use Rust or Go for the new high-performance microservice?");
    console.log(`✅ Session Started: ${session.id} (Topic: ${session.topic})`);

    // 2. Submit Opinions
    council.submitOpinion(session.id, "The Architect", "Rust offers better memory safety guarantees and performance.");
    council.submitOpinion(session.id, "The Pragmatist", "Go has faster build times and simpler concurrency for our team.");
    council.submitOpinion(session.id, "The Critic", "Both have steep learning curves depending on the team's background.");

    // Check opinions
    const updatedSession = council.getSession(session.id)!;
    if (updatedSession.opinions.length !== 3) {
        throw new Error(`Expected 3 opinions, got ${updatedSession.opinions.length}`);
    }
    console.log(`✅ Opinions verified (${updatedSession.opinions.length})`);

    // 3. Advance Round
    council.advanceRound(session.id);
    if (updatedSession.round !== 2) {
        throw new Error(`Expected round 2, got ${updatedSession.round}`);
    }
    console.log("✅ Round Advanced");

    // 4. Vote
    council.castVote(session.id, "The Architect", "Rust", "Safety first.");
    council.castVote(session.id, "The Pragmatist", "Go", "Speed to market.");
    council.castVote(session.id, "The Critic", "Rust", "Long-term maintainability.");

    // Check votes
    if (updatedSession.votes.length !== 3) {
        throw new Error(`Expected 3 votes, got ${updatedSession.votes.length}`);
    }
    console.log(`✅ Votes verified (${updatedSession.votes.length})`);

    // 5. Conclude
    council.concludeSession(session.id);
    if (updatedSession.status !== 'concluded') {
        throw new Error("Session failed to conclude");
    }
    console.log("✅ Session Concluded");

    console.log("🎉 Council Service Verification Passed!");
}

verifyCouncil().catch(e => {
    console.error("❌ Verification Failed:", e);
    process.exit(1);
});
