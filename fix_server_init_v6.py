import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    if 'server.sessionManager = session.NewSessionManager(100)' in line:
        new_lines.append(line)
        new_lines.append('\t// Phase 3 & 4 logic\n')
        new_lines.append('\tmemoryDB := filepath.Join(cfg.ConfigDir, "memory.db")\n')
        new_lines.append('\tmemoryVS, _ := memorystore.NewVectorStore(memoryDB)\n')
        new_lines.append('\tserver.fleetManager = orchestration.NewFleetManagerPlus(memoryVS, server.eventBus, server.supervisorManager)\n')
        new_lines.append('\tserver.a2aBroker.SetSignalProcessor(server.fleetManager)\n')
        new_lines.append('\tserver.consensusEngine = orchestration.NewConsensusEngine(server.debateHistory, memoryVS)\n')
        new_lines.append('\tserver.quotaManager = providers.NewQuotaManager()\n')
        new_lines.append('\tserver.modelSelector = providers.NewModelSelector(server.quotaManager)\n')
        continue

    if 'server.memoryReactor = memorystore.NewMemoryReactor(cfg.WorkspaceRoot)' in line:
        new_lines.append('\tserver.memoryReactor = memorystore.NewMemoryReactor(cfg.WorkspaceRoot, memoryVS)\n')
        continue

    if 'server.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot)' in line:
        new_lines.append('\tserver.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot, memoryVS)\n')
        continue

    new_lines.append(line)

with open(file_path, 'w') as f:
    f.writelines(new_lines)
