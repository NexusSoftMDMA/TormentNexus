# mvmtOS V1: Folder Leases

V1 makes folder leases the product surface.

```text
one lease = one recipient/use-case = one scoped token = one or more paths
```

The durable abstraction is not MCP or AI permissions. It is scoped, expiring,
audited access to local files and folders on the user's own machine.

## V1 Contract

- A lease is read-only by default.
- Upload mode is upload-only: recipients can add files but cannot browse,
  download, delete, or overwrite files.
- A read lease can point at one or more local files or folders.
- An upload lease points at one local folder.
- A lease has a required human label.
- A lease defaults to 24 hours.
- A lease can also last until revoked.
- A lease cannot expose paths outside its folder.
- A lease cannot bypass mount exclude rules or global secret-path blocks.
- Unsupported file types may still be downloaded.
- Search/indexing is allowed to support fewer file types than download.
- MCP is an adapter over the same permission engine, not the product core.

## Out Of Scope

- General write/edit access.
- Deletes.
- Team ACLs.
- Hosted relay.
- Sync.
- Backup.
- Public admin dashboard.

## First Milestone

```bash
mvmt lease create ~/Documents/Taxes --label "Sarah - tax docs"
```

The generated link opens a scoped folder view. The recipient can browse and
download files from that folder until the lease expires or is revoked.

```bash
mvmt lease create ~/Documents/Taxes ~/Documents/Receipts --label "Sarah - tax docs"
```

The same token can cover multiple files or folders. Browser recipients see one
lease namespace, and MCP clients authenticate with the same lease token.

```bash
mvmt lease add-path <lease-id> ~/Documents/MoreReceipts
```

Admins can add paths to an active read lease without rotating the token or
creating a new URL. The recipient gets the expanded namespace on the next
browser or MCP request.

```bash
mvmt lease create ~/Uploads --label "Sarah uploads" --mode upload
```

Upload leases open a drop page. The recipient can add new files, but cannot
browse, download, overwrite, edit, or delete anything in that folder.
