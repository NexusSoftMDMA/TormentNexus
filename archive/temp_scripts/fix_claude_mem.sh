cd archive/claude-mem || true
git config --file .gitmodules --remove-section submodule.target-repo 2>/dev/null || true
git rm --cached target-repo 2>/dev/null || true
cd ../../packages/claude-mem || true
git config --file .gitmodules --remove-section submodule.target-repo 2>/dev/null || true
git rm --cached target-repo 2>/dev/null || true
