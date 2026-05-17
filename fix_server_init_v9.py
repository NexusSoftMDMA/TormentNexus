import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    new_lines.append(line)
    if 'memoryVS, _ := memorystore.NewVectorStore(filepath.Join(cfg.ConfigDir, "memory.db"))' in line:
        new_lines.append('\tserver.memoryReactor = memorystore.NewMemoryReactor(cfg.WorkspaceRoot, memoryVS)\n')
        new_lines.append('\tserver.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot, memoryVS)\n')
        new_lines.append('\tserver.consensusEngine = orchestration.NewConsensusEngine(server.debateHistory, memoryVS)\n')

with open(file_path, 'w') as f:
    f.writelines(new_lines)
