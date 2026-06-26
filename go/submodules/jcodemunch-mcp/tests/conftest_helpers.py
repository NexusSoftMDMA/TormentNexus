"""Shared test helpers for session-aware routing features."""
from pathlib import Path


def _write(path: Path, content: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def create_mini_index(tmp_path: Path, filename: str = "test_module.py") -> tuple[str, str]:
    """Create a minimal indexed repo. Returns (repo_id, storage_path_str)."""
    from jcodemunch_mcp.tools.index_folder import index_folder

    _write(tmp_path / filename, (
        "def my_func(x: int, y: int) -> int:\n"
        "    '''Add two numbers.'''\n"
        "    return x + y\n\n"
        "class MyClass:\n"
        "    def method(self):\n"
        "        pass\n"
    ))
    sp = str(tmp_path / "idx")
    result = index_folder(path=str(tmp_path), use_ai_summaries=False, storage_path=sp)
    return result["repo"], sp


def create_exact_match_index(tmp_path: Path, filename: str = "ui_builder.py") -> tuple[str, str]:
    """Create an indexed repo with exact snake_case/camelCase method names.

    The methods intentionally contain call references so the SQLite loader must
    reconstruct v8 array-style ``data`` rows before search tools apply the
    ``language='python'`` filter.
    """
    from jcodemunch_mcp.tools.index_folder import index_folder

    _write(tmp_path / filename, (
        "class UiBuilder:\n"
        "    def helper(self):\n"
        "        return 1\n\n"
        "    def _build_left_pane_cache(self):\n"
        "        self.helper()\n"
        "        return {}\n\n"
        "    def renderRow(self):\n"
        "        self.helper()\n"
        "        return 1\n\n"
        "    def set_ui(self):\n"
        "        return None\n\n"
        "    def build_ui(self):\n"
        "        self.helper()\n"
        "        return self.set_ui()\n"
    ))
    sp = str(tmp_path / "idx")
    result = index_folder(path=str(tmp_path), use_ai_summaries=False, storage_path=sp)
    return result["repo"], sp


def get_index(repo: str, storage_path: str):
    """Load the CodeIndex for a test repo."""
    from jcodemunch_mcp.storage.sqlite_store import SQLiteIndexStore
    store = SQLiteIndexStore(base_path=storage_path)
    parts = repo.split("/", 1)
    return store.load_index(parts[0], parts[1])
