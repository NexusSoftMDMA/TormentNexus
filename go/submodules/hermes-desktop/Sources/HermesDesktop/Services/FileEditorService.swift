import Foundation

final class FileEditorService: @unchecked Sendable {
    private let sshTransport: SSHTransport
    private let maxEditableFileBytes = WorkspaceFileLimits.maxEditableFileBytes
    private let maxDirectoryEntries = 500

    init(sshTransport: SSHTransport) {
        self.sshTransport = sshTransport
    }

    static func makeReadScript(
        remotePath: String,
        maxEditableBytes: Int64
    ) throws -> String {
        try RemotePythonScript.wrap(
            FileRequest(path: remotePath, maxEditableBytes: maxEditableBytes),
            body: """
            import hashlib
            import json
            import pathlib

            def editable_file_target(path):
                if path.is_symlink():
                    try:
                        resolved = path.resolve(strict=True)
                    except FileNotFoundError:
                        fail(f"{payload['path']} is a dangling symlink.")
                    if not resolved.is_file():
                        fail(f"{payload['path']} points to a non-file target.")
                    return resolved

                return path

            try:
                requested = expand_remote_path(payload["path"]) or pathlib.Path(payload["path"])
                target = editable_file_target(requested)
                if not target.exists():
                    fail(f"{payload['path']} does not exist on the active host.")
                if not target.is_file():
                    fail(f"{payload['path']} is not a regular file.")

                size = target.stat().st_size
                max_size = int(payload.get("max_editable_bytes") or 0)
                if max_size > 0 and size > max_size:
                    size_mb = size / 1000000
                    limit_mb = max_size / 1000000
                    fail(f"This file is {size_mb:.1f} MB. Hermes Desktop can edit remote text files up to {limit_mb:g} MB.")

                raw_content = target.read_bytes()
                content_hash = hashlib.sha256(raw_content).hexdigest()
                content = raw_content.decode("utf-8")
                print(json.dumps({
                    "ok": True,
                    "content": content,
                    "content_hash": content_hash,
                }, ensure_ascii=False))
            except UnicodeDecodeError:
                fail(f"{payload['path']} is not valid UTF-8.")
            except PermissionError:
                fail(f"Permission denied while reading {payload['path']}.")
            except Exception as exc:
                fail(f"Unable to read {payload['path']}: {exc}")
            """
        )
    }

    static func makeDirectoryListScript(
        remotePath: String,
        hermesHome: String?,
        maxEntries: Int
    ) throws -> String {
        try RemotePythonScript.wrap(
            DirectoryListRequest(
                path: remotePath,
                hermesHome: hermesHome,
                maxEntries: maxEntries
            ),
            body: """
            import json
            import os
            import pathlib

            try:
                home = pathlib.Path.home()
                hermes_home = resolved_hermes_home(payload)
                requested_path = payload.get("path") or payload.get("hermes_home") or str(hermes_home)
                target = expand_remote_path(requested_path, home=home, base_dir=hermes_home)

                if not target.exists():
                    fail(f"{payload['path']} does not exist on the active host.")
                if not target.is_dir():
                    fail(f"{payload['path']} is not a directory.")

                max_entries = int(payload.get("max_entries") or 500)
                children = list(target.iterdir())

                def entry_sort_key(item):
                    try:
                        is_directory = item.is_dir()
                    except OSError:
                        is_directory = False
                    return (0 if is_directory else 1, item.name.lower())

                children.sort(key=entry_sort_key)
                limited_children = children[:max_entries]

                entries = []
                for item in limited_children:
                    stat_result = None
                    try:
                        stat_result = item.stat()
                    except OSError:
                        stat_result = None

                    try:
                        is_directory = item.is_dir()
                    except OSError:
                        is_directory = False

                    try:
                        is_file = item.is_file()
                    except OSError:
                        is_file = False

                    is_symlink = item.is_symlink()
                    if is_symlink:
                        kind = "symlink"
                    elif is_directory:
                        kind = "directory"
                    elif is_file:
                        kind = "file"
                    else:
                        kind = "other"

                    entries.append({
                        "name": item.name,
                        "path": item.as_posix(),
                        "display_path": tilde(item, home),
                        "kind": kind,
                        "size": None if is_directory or stat_result is None else stat_result.st_size,
                        "modified_at": None if stat_result is None else stat_result.st_mtime,
                        "is_readable": os.access(item, os.R_OK),
                        "is_writable": os.access(item, os.W_OK),
                        "is_symlink": is_symlink,
                    })

                parent = target.parent if target.parent != target else None

                print(json.dumps({
                    "ok": True,
                    "requested_path": requested_path,
                    "resolved_path": target.as_posix(),
                    "display_path": tilde(target, home),
                    "parent_path": None if parent is None else parent.as_posix(),
                    "parent_display_path": None if parent is None else tilde(parent, home),
                    "entries": entries,
                    "total_entry_count": len(children),
                    "is_truncated": len(children) > len(limited_children),
                }, ensure_ascii=False))
            except PermissionError:
                fail(f"Permission denied while reading {payload['path']}.")
            except Exception as exc:
                fail(f"Unable to list {payload['path']}: {exc}")
            """
        )
    }

    static func makeWriteScript(
        remotePath: String,
        content: String,
        expectedContentHash: String?
    ) throws -> String {
        try RemotePythonScript.wrap(
            FileWriteRequest(
                path: remotePath,
                content: content,
                expectedContentHash: expectedContentHash,
                atomic: true
            ),
            body: """
            import hashlib
            import json
            import os
            import pathlib
            import tempfile

            temp_name = None
            directory_fd = None
            content_bytes = payload["content"].encode("utf-8")
            expected_hash = payload.get("expected_content_hash")

            def editable_file_target(path):
                if path.is_symlink():
                    try:
                        resolved = path.resolve(strict=True)
                    except FileNotFoundError:
                        fail(f"{payload['path']} is a dangling symlink.")
                    if not resolved.is_file():
                        fail(f"{payload['path']} points to a non-file target.")
                    return resolved

                return path

            try:
                requested = expand_remote_path(payload["path"]) or pathlib.Path(payload["path"])
                target = editable_file_target(requested)

                if expected_hash is not None:
                    if not target.exists():
                        fail(f"{payload['path']} was removed on the active host after it was loaded. Reload from Remote before saving.")
                    if not target.is_file():
                        fail(f"{payload['path']} is not a regular file anymore. Reload from Remote before saving.")

                    current_bytes = target.read_bytes()
                    current_hash = hashlib.sha256(current_bytes).hexdigest()
                    if current_hash != expected_hash:
                        fail(f"{payload['path']} changed on the active host after it was loaded. Reload from Remote before saving.")

                target.parent.mkdir(parents=True, exist_ok=True)

                fd, temp_name = tempfile.mkstemp(
                    dir=str(target.parent),
                    prefix=f".{target.name}.",
                    suffix=".tmp",
                )

                with os.fdopen(fd, "wb") as handle:
                    handle.write(content_bytes)
                    handle.flush()
                    os.fsync(handle.fileno())

                if target.exists():
                    os.chmod(temp_name, target.stat().st_mode)

                os.replace(temp_name, target)

                directory_fd = os.open(target.parent, os.O_RDONLY)
                os.fsync(directory_fd)

                print(json.dumps({
                    "ok": True,
                    "path": payload["path"],
                    "content_hash": hashlib.sha256(content_bytes).hexdigest(),
                }, ensure_ascii=False))
            except PermissionError:
                fail(f"Permission denied while writing {payload['path']}.")
            except Exception as exc:
                fail(f"Unable to write {payload['path']}: {exc}")
            finally:
                if directory_fd is not None:
                    os.close(directory_fd)
                if temp_name and os.path.exists(temp_name):
                    os.unlink(temp_name)
            """
        )
    }

    func read(
        remotePath: String,
        connection: ConnectionProfile
    ) async throws -> FileSnapshot {
        let script = try Self.makeReadScript(
            remotePath: remotePath,
            maxEditableBytes: maxEditableFileBytes
        )

        let response = try await sshTransport.executeJSON(
            on: connection,
            pythonScript: script,
            responseType: FileReadResponse.self
        )

        return FileSnapshot(
            content: response.content,
            contentHash: response.contentHash
        )
    }

    func read(
        file: RemoteTrackedFile,
        remotePath: String,
        connection: ConnectionProfile
    ) async throws -> FileSnapshot {
        try await read(remotePath: remotePath, connection: connection)
    }

    func listDirectory(
        remotePath: String,
        hermesHome: String?,
        connection: ConnectionProfile
    ) async throws -> RemoteDirectoryListing {
        let script = try Self.makeDirectoryListScript(
            remotePath: remotePath,
            hermesHome: hermesHome,
            maxEntries: maxDirectoryEntries
        )

        return try await sshTransport.executeJSON(
            on: connection,
            pythonScript: script,
            responseType: RemoteDirectoryListing.self
        )
    }

    func write(
        remotePath: String,
        content: String,
        expectedContentHash: String?,
        connection: ConnectionProfile
    ) async throws -> FileSaveResult {
        let script = try Self.makeWriteScript(
            remotePath: remotePath,
            content: content,
            expectedContentHash: expectedContentHash
        )

        let response = try await sshTransport.executeJSON(
            on: connection,
            pythonScript: script,
            responseType: FileWriteResponse.self
        )

        return FileSaveResult(
            path: response.path,
            contentHash: response.contentHash
        )
    }

    func write(
        file: RemoteTrackedFile,
        remotePath: String,
        content: String,
        expectedContentHash: String?,
        connection: ConnectionProfile
    ) async throws -> FileSaveResult {
        try await write(
            remotePath: remotePath,
            content: content,
            expectedContentHash: expectedContentHash,
            connection: connection
        )
    }
}

private struct FileRequest: Encodable {
    let path: String
    let maxEditableBytes: Int64

    enum CodingKeys: String, CodingKey {
        case path
        case maxEditableBytes = "max_editable_bytes"
    }
}

private struct DirectoryListRequest: Encodable {
    let path: String
    let hermesHome: String?
    let maxEntries: Int

    enum CodingKeys: String, CodingKey {
        case path
        case hermesHome = "hermes_home"
        case maxEntries = "max_entries"
    }
}

private struct FileWriteRequest: Encodable {
    let path: String
    let content: String
    let expectedContentHash: String?
    let atomic: Bool

    enum CodingKeys: String, CodingKey {
        case path
        case content
        case expectedContentHash = "expected_content_hash"
        case atomic
    }
}

private struct FileReadResponse: Decodable {
    let ok: Bool
    let content: String
    let contentHash: String

    enum CodingKeys: String, CodingKey {
        case ok
        case content
        case contentHash = "content_hash"
    }
}

private struct FileWriteResponse: Decodable {
    let ok: Bool
    let path: String
    let contentHash: String

    enum CodingKeys: String, CodingKey {
        case ok
        case path
        case contentHash = "content_hash"
    }
}

struct FileSnapshot {
    let content: String
    let contentHash: String
}

struct FileSaveResult {
    let path: String
    let contentHash: String
}
