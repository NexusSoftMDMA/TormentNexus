import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
skip_old_init = False
for line in lines:
    if 'server.sessionManager = session.NewSessionManager(100)' in line:
        new_lines.append(line)
        new_lines.append('\t// Phase 3 & 4: Fleet, Shared Memory, Consensus, Quotas\n')
        new_lines.append('\tmemoryVS, _ := memorystore.NewVectorStore(filepath.Join(cfg.ConfigDir, "memory.db"))\n')
        new_lines.append('\tserver.fleetManager = orchestration.NewFleetManagerPlus(memoryVS, server.eventBus, server.supervisorManager)\n')
        new_lines.append('\tserver.a2aBroker.SetSignalProcessor(server.fleetManager)\n')
        new_lines.append('\tserver.quotaManager = providers.NewQuotaManager()\n')
        new_lines.append('\tserver.modelSelector = providers.NewModelSelector(server.quotaManager)\n')
        new_lines.append('\tserver.consensusEngine = orchestration.NewConsensusEngine(server.debateHistory, memoryVS)\n')
        continue

    if 'server.memoryReactor = memorystore.NewMemoryReactor(cfg.WorkspaceRoot, memoryVS)' in line:
        continue # Already handled or will be handled
    if 'server.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot, archiverVS)' in line:
        continue

    new_lines.append(line)

# This script is still too risky. I'll just replace the whole New function body if I have to.
