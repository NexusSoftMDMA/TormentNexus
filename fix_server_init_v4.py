import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    new_lines.append(line)
    if 'server.sessionManager = session.NewSessionManager(100)' in line:
        new_lines.append('\t// Phase 4: Consensus & Budget-Aware Routing\n')
        new_lines.append('\tserver.quotaManager = providers.NewQuotaManager()\n')
        new_lines.append('\tserver.modelSelector = providers.NewModelSelector(server.quotaManager)\n')
        new_lines.append('\tserver.consensusEngine = orchestration.NewConsensusEngine(server.debateHistory, memoryVS)\n')

with open(file_path, 'w') as f:
    f.writelines(new_lines)
