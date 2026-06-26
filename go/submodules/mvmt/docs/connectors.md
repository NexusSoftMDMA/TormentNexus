# Mount Providers

The active runtime is mount-based.

A mount maps a virtual path to a provider:

```yaml
mounts:
  - name: workspace
    type: local_folder
    path: /workspace
    root: /Users/you/code/mvmt
```

Clients call the same five tools regardless of provider:

```text
search
list
read
write
remove
```

Today, the only shipped provider is `local_folder`.

## Local Folder Provider

`local_folder` exposes a selected directory through a virtual path.

Responsibilities:

- resolve virtual paths under the configured root;
- reject path traversal;
- reject symlink escapes;
- hide excluded paths;
- block writes and removes for protected paths;
- enforce mount-level `writeAccess`;
- return text-only results to the tool layer.

Host filesystem paths are not returned to clients.

## Provider Interface Direction

The current implementation is local-folder specific, but the product direction is a provider interface behind the same tool surface.

Future providers should fit this shape:

```ts
interface StorageProvider {
  list(path: string): Promise<ListEntry[]>;
  read(path: string): Promise<TextFile>;
  write(path: string, content: string): Promise<TextFile>;
  remove(path: string): Promise<void>;
  walkTextFiles(): AsyncIterable<TextFile>;
}
```

Future candidates:

- another mvmt instance mounted remotely;
- S3 or object storage;
- Dropbox or Drive;
- a database-backed document store.

Do not add a new user-facing tool set for each provider. Prefer adapting the provider to the existing `search` / `list` / `read` / `write` / `remove` surface.

## Legacy Connector Code

The repository still contains legacy connector and proxy modules for compatibility with older config and branches.

New product work should not expand the raw proxy tool surface. Use mounts and providers instead.

## Security Checklist

Before adding a provider, answer these questions in code and docs:

- What exact user-selected scope does this provider expose?
- What is read-only by default?
- Which operations write, mutate, delete, move, append, or upload data?
- How are path traversal, symlink escape, bucket escape, table escape, or network escape blocked?
- How does the provider honor `exclude`, `protect`, and `writeAccess`?
- What does `mvmt doctor` verify?
- What appears in the audit log?

If the safest answer is "exclude that data from scope," prefer scope controls over post-processing filters.
