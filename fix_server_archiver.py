import sys
import os

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    if 'server.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot)' in line:
        new_lines.append('\t// VectorStore for the archiver\n')
        new_lines.append('\tvs, _ := memorystore.NewVectorStore(filepath.Join(cfg.ConfigDir, "memory.db"))\n')
        new_lines.append('\tserver.memoryArchiver = memorystore.NewMemoryArchiver(cfg.WorkspaceRoot, vs)\n')
    else:
        new_lines.append(line)

with open(file_path, 'w') as f:
    f.writelines(new_lines)
