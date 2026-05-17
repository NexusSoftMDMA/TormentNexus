import sys

file_path = 'go/internal/httpapi/server.go'
with open(file_path, 'r') as f:
    lines = f.readlines()

new_lines = []
for line in lines:
    new_lines.append(line)
    if 'memoryArchiver    *memorystore.MemoryArchiver' in line:
        new_lines.append('\tfleetManager      *orchestration.FleetManagerPlus\n')
        new_lines.append('\tconsensusEngine   *orchestration.ConsensusEngine\n')
        new_lines.append('\tquotaManager      *providers.QuotaManager\n')
        new_lines.append('\tmodelSelector     *providers.ModelSelector\n')

with open(file_path, 'w') as f:
    f.writelines(new_lines)
