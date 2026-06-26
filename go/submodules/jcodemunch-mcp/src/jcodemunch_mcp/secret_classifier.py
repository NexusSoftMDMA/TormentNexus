"""Structured credential / secret-file classifier.

Decides whether a file path should be excluded from indexing because it holds
credential material, returning a structured :class:`SecretFileDecision` so the
verdict is explainable (reason, group, matched pattern, confidence) for logging,
overrides, and tests. :func:`classify_secret_file` is the pure core;
``security.is_secret_file`` is the boolean wrapper most call sites use.

Design notes
------------
The classifier is filename- and path-shape only — it never reads file contents
(that is response-level redaction's job, in ``redact.py``). It is grounded in the
conventional layout of real credential files: dotfiles (``.env``, ``.netrc``),
provider credential JSON (service accounts, OAuth client secrets, Firebase admin
SDK), private-key containers (``*.pem``/``*.p12``/``*.jks``), path-specific
credential stores (``~/.aws/credentials``, ``~/.kube/config``), secret-store
directories (``secrets/``, ``vault/``), key-material directories (``keys/``,
``certs/`` — private keys only, NOT public certs), and a token-boundary broad
``secret`` basename heuristic that no longer trips on substrings like
``secretariat.csv``.

A source module that *handles* secrets (``secret_redaction.py``,
``secret_scanner.ts``) is code, not credential material, so source-code
extensions are exempt from the broad heuristic — the same carve-out applies to
documentation prose.

Precedence (first match wins): caller-supplied path globs -> path-specific
credentials -> exact credential names -> key-material directory -> credential
extension -> secret-store data -> broad secret basename -> not secret.
"""

from __future__ import annotations

import re
from fnmatch import fnmatchcase
from typing import Iterable, NamedTuple, Optional


# ── Result type ───────────────────────────────────────────────────────────────

class SecretFileDecision(NamedTuple):
    """Structured classification result.

    is_secret:       final boolean verdict (the only field most callers need).
    reason:          stable category slug (see the ``*_GROUP`` constants).
    group:           the pattern group that fired (``"none"`` when not secret).
    matched_pattern: the specific pattern/path that matched, if any.
    confidence:      ``"high"`` / ``"medium"`` / ``"none"`` — exact credential
                     contracts are high; heuristic directory/basename rules are
                     medium.
    normalized_path: the lowercased forward-slash path the decision was made on.
    """

    is_secret: bool
    reason: str
    group: str
    matched_pattern: Optional[str]
    confidence: str
    normalized_path: str


# ── Group identifiers (also valid `exclude_secret_patterns` override tokens) ──

GROUP_PATH_SPECIFIC = "path_specific_credentials"
GROUP_EXACT_NAME = "exact_credential_names"
GROUP_KEY_MATERIAL = "key_material_directories"
GROUP_CREDENTIAL_EXT = "credential_extensions"
GROUP_SECRET_STORE = "secret_store_data"
GROUP_BROAD_BASENAME = "broad_secret_basenames"

ALL_GROUPS = frozenset({
    GROUP_PATH_SPECIFIC, GROUP_EXACT_NAME, GROUP_KEY_MATERIAL,
    GROUP_CREDENTIAL_EXT, GROUP_SECRET_STORE, GROUP_BROAD_BASENAME,
})


# ── Pattern data ──────────────────────────────────────────────────────────────

# Exact credential-file basenames (fnmatch globs, matched case-insensitively).
EXACT_CREDENTIAL_NAME_PATTERNS = (
    ".env", ".env.*", "*.env", "*.env.*",
    ".htpasswd", ".netrc", ".npmrc", ".pypirc",
    "application_default_credentials.json",
    "client_secret*.json", "oauth2_client_secret*.json",
    "credentials.json",
    "service-account*.json", "service_account*.json",
    "*-firebase-adminsdk-*.json",
    "token.json",
    "*.agekey", "*.credentials", "*.secrets", "*.token",
)

# SSH private keys. Public companions (``*.pub``) are explicitly exempted.
SSH_PRIVATE_KEY_PATTERNS = (
    "id_rsa", "id_rsa.*",
    "id_ed25519", "id_ed25519.*",
    "id_dsa", "id_dsa.*",
    "id_ecdsa", "id_ecdsa.*",
)

# Private-key / keystore container extensions (basename globs).
CREDENTIAL_EXTENSION_PATTERNS = (
    "*.pem", "*.key", "*.p8", "*.p12", "*.pfx",
    "*.ppk", "*.jks", "*.keystore",
)

# Well-known credential stores that only a PATH (not a basename) identifies.
PATH_SPECIFIC_SECRET_PATTERNS = (
    ".aws/credentials",
    ".azure/accesstokens.json",
    ".cargo/credentials",
    ".cargo/credentials.toml",
    ".config/gcloud/application_default_credentials.json",
    ".docker/config.json",
    ".gem/credentials",
    ".kube/config",
    "composer/auth.json",
    "gcloud/application_default_credentials.json",
)

# Whole-segment directory names that conventionally hold credential MATERIAL as
# data/config files (k8s Secrets, Vault, Terraform). Matched as exact segments,
# never substrings — ``secrets-manager`` is a service, not a store.
SECRET_STORE_DIRECTORY_NAMES = frozenset({
    "secret", "secrets", "credential", "credentials", "creds", "vault",
})

# Data/config extensions that are credential material when under a secret store.
DATA_CONFIG_EXTENSIONS = frozenset({
    ".yaml", ".yml", ".json", ".ini", ".cfg", ".conf", ".config",
    ".properties", ".toml", ".xml", ".env", ".tfvars", ".tfstate",
})

# Compound (multi-dot) suffixes that os.path.splitext can't see in one shot.
COMPOUND_DATA_CONFIG_SUFFIXES = (
    ".tfstate.backup", ".tfvars.json", ".auto.tfvars",
)

# Key-material directories: a private-key/keystore file here is credential
# material. Public certs (.crt/.cer/.der) are NOT skipped just for living here.
KEY_MATERIAL_DIRECTORY_NAMES = frozenset({
    "keys", "private-keys", "private_keys",
    "certs", "certificates", "ssl", "tls", "pki",
})
PRIVATE_KEY_MATERIAL_EXTENSIONS = frozenset({
    ".pem", ".key", ".p8", ".p12", ".pfx", ".ppk", ".jks", ".keystore",
})

# Broad heuristic: a `secret`/`secrets` token at a name boundary (so
# ``prod-secrets.yaml`` hits but ``secretariat.csv`` / ``prodsecret.yaml`` do
# not), excluding source/doc/template files.
SECRET_WORD_PATTERN = re.compile(r"(^|[._-])secrets?($|[._-])")

# Programming-language extensions: a *secret* source module is code that handles
# secrets, not a credential file. Deliberately NOT data/config markup.
SOURCE_EXTENSIONS = frozenset({
    ".py", ".pyw", ".pyi",
    ".js", ".jsx", ".mjs", ".cjs", ".ts", ".tsx",
    ".go", ".rs", ".rb", ".php", ".java", ".kt", ".kts", ".scala",
    ".c", ".h", ".cc", ".cpp", ".cxx", ".hpp", ".hh", ".hxx",
    ".cs", ".swift", ".dart", ".m", ".mm",
    ".sh", ".bash", ".zsh", ".fish", ".ps1",
    ".lua", ".pl", ".pm", ".r", ".ex", ".exs", ".erl", ".clj", ".cljs",
    ".fs", ".fsx", ".vb", ".groovy", ".sql",
})
DOCUMENTATION_EXTENSIONS = frozenset({
    ".md", ".markdown", ".mdx", ".rst", ".txt",
    ".adoc", ".asciidoc", ".asc", ".html", ".htm", ".ipynb",
})

# Template/example/sample markers: a fixture is not a live credential.
TEMPLATE_MARKERS = frozenset({"example", "sample", "template", "tmpl", "dist"})


# ── Helpers ───────────────────────────────────────────────────────────────────

def _normalize_path(file_path: str) -> str:
    """Lowercase, forward-slash, drop ``.`` and empty segments."""
    lowered = file_path.replace("\\", "/").lower()
    parts = [p for p in lowered.split("/") if p and p != "."]
    return "/".join(parts)


def _final_suffix(basename: str) -> str:
    dot = basename.rfind(".")
    return basename[dot:] if dot > 0 else ""


def _has_compound_suffix(basename: str, suffixes: Iterable[str]) -> bool:
    return any(basename.endswith(s) for s in suffixes)


def _is_template_name(basename: str) -> bool:
    parts = [p for p in re.split(r"[._-]+", basename) if p]
    return any(p in TEMPLATE_MARKERS for p in parts)


def _is_public_ssh_key(basename: str) -> bool:
    return basename.endswith(".pub")


def _path_matches(normalized_path: str, pattern: str) -> bool:
    """Glob match, or exact / trailing-segment match for plain patterns."""
    if any(c in pattern for c in "*?[]"):
        return fnmatchcase(normalized_path, pattern)
    return normalized_path == pattern or normalized_path.endswith("/" + pattern)


def _hit(reason: str, group: str, pattern: Optional[str], confidence: str,
         normalized_path: str) -> SecretFileDecision:
    return SecretFileDecision(True, reason, group, pattern, confidence, normalized_path)


# ── Core classifier ───────────────────────────────────────────────────────────

def classify_secret_file(
    file_path: str,
    *,
    disabled_groups: Iterable[str] = (),
    extra_secret_path_patterns: Iterable[str] = (),
    allow_patterns: Iterable[str] = (),
) -> SecretFileDecision:
    """Classify a path by filename and directory contracts only.

    Args:
        file_path: Relative or absolute path (any separator style).
        disabled_groups: Group slugs to skip (see the ``GROUP_*`` constants),
            for repos that opt a category out without losing the others.
        extra_secret_path_patterns: Caller-supplied path globs honored before the
            built-in groups (e.g. a project-specific credential location).
        allow_patterns: Globs that, when they match the basename or path, force a
            not-secret verdict regardless of any group — the per-pattern opt-out
            behind the ``exclude_secret_patterns`` config key.

    Returns:
        A :class:`SecretFileDecision`. ``is_secret`` is the headline; the rest
        explains why.
    """
    normalized = _normalize_path(file_path)
    basename = normalized.rsplit("/", 1)[-1]
    segments = [s for s in normalized.split("/") if s]
    parent_segments = segments[:-1]
    suffix = _final_suffix(basename)
    disabled = set(disabled_groups)

    # Caller opt-out: an allow pattern matching the basename or path wins outright.
    for pattern in allow_patterns:
        p = pattern.replace("\\", "/").lower()
        if fnmatchcase(basename, p) or _path_matches(normalized, p):
            return SecretFileDecision(False, "allowlisted", "none", p, "none", normalized)

    # 0. Caller-supplied path globs — highest precedence among secret rules.
    for pattern in extra_secret_path_patterns:
        norm_pattern = _normalize_path(pattern) if "*" not in pattern and "?" not in pattern else pattern.replace("\\", "/").lower()
        if _path_matches(normalized, norm_pattern):
            return _hit("path_specific_credential", "extra_secret_path_patterns",
                        norm_pattern, "high", normalized)

    # 1. Path-specific credential stores (a basename alone can't identify these).
    if GROUP_PATH_SPECIFIC not in disabled:
        for pattern in PATH_SPECIFIC_SECRET_PATTERNS:
            if _path_matches(normalized, pattern):
                return _hit("path_specific_credential", GROUP_PATH_SPECIFIC,
                            pattern, "high", normalized)

    # 2. Exact credential-file basenames (incl. SSH private keys).
    if GROUP_EXACT_NAME not in disabled and not _is_template_name(basename):
        for pattern in (*EXACT_CREDENTIAL_NAME_PATTERNS, *SSH_PRIVATE_KEY_PATTERNS):
            if fnmatchcase(basename, pattern) and not _is_public_ssh_key(basename):
                return _hit("exact_credential_name", GROUP_EXACT_NAME,
                            pattern, "high", normalized)

    # 3. Key-material directory + private-key extension (before the generic
    #    credential-extension group so the more specific reason wins).
    if GROUP_KEY_MATERIAL not in disabled:
        if suffix in PRIVATE_KEY_MATERIAL_EXTENSIONS and any(
            seg in KEY_MATERIAL_DIRECTORY_NAMES for seg in parent_segments
        ):
            return _hit("key_material_directory", GROUP_KEY_MATERIAL,
                        "key-material directory + private-key suffix", "high", normalized)

    # 4. Private-key / keystore container extensions, anywhere.
    if GROUP_CREDENTIAL_EXT not in disabled and not _is_template_name(basename):
        for pattern in CREDENTIAL_EXTENSION_PATTERNS:
            if fnmatchcase(basename, pattern):
                return _hit("credential_extension", GROUP_CREDENTIAL_EXT,
                            pattern, "high", normalized)

    # 5. Data/config files inside a whole-segment secret-store directory.
    if GROUP_SECRET_STORE not in disabled:
        in_store = any(seg in SECRET_STORE_DIRECTORY_NAMES for seg in parent_segments)
        store_suffix = (
            suffix in DATA_CONFIG_EXTENSIONS
            or _has_compound_suffix(basename, COMPOUND_DATA_CONFIG_SUFFIXES)
        )
        if in_store and store_suffix:
            return _hit("secret_store_data", GROUP_SECRET_STORE,
                        "secret-store directory + data/config suffix", "medium", normalized)

    # 6. Broad token-boundary `secret` basename (non-source, non-doc, non-template).
    if GROUP_BROAD_BASENAME not in disabled:
        if (
            SECRET_WORD_PATTERN.search(basename) is not None
            and suffix not in SOURCE_EXTENSIONS
            and suffix not in DOCUMENTATION_EXTENSIONS
            and not _is_template_name(basename)
        ):
            return _hit("broad_secret_basename", GROUP_BROAD_BASENAME,
                        "token-boundary secret basename", "medium", normalized)

    return SecretFileDecision(False, "not_secret", "none", None, "none", normalized)
