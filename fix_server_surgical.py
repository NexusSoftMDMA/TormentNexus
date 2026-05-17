import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
seen_fleet = False
seen_consensus = False
for line in lines:
    if 'fleetManager      *orchestration.FleetManagerPlus' in line:
        if seen_fleet: continue
        seen_fleet = True
    if 'consensusEngine   *orchestration.ConsensusEngine' in line:
        if seen_consensus: continue
        seen_consensus = True

    if 'server.consensusEngine = orchestration.NewConsensusEngine(server.debateHistory)' in line:
        # Replace with correct call
        # But wait, I need memoryVS to be defined.
        pass

    new_lines.append(line)
