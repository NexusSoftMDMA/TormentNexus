# Release notes

Historical Hermes Desktop release notes live in this folder to keep the repo
root focused on the app, packaging, and trust-facing docs.

## Tag policy

- Official release tags use the `vX.Y.Z` format, for example `v0.9.0`.
- The published GitHub Release should point to the exact tagged commit used to
  build `HermesDesktop.app.zip`.
- For local release candidates before tagging, it is fine to package with an
  explicit `HERMES_VERSION=0.9.0`, but that artifact should not be presented as
  the final public release unless the tag and commit match.
- Once the final tag exists, prefer packaging from that tagged commit without
  version overrides so the app bundle, manifest, checksum, and release page all
  describe the same build.
- Create the annotated release tag only after the release commit is on `main`
  and the `main` CI run is green.
- Push the tag, require the tag-triggered macOS CI packaging and verification
  job to pass, then publish the GitHub Release from that exact tag.
- Attach all three verified assets: `HermesDesktop.app.zip`,
  `HermesDesktop.app.zip.sha256`, and
  `HermesDesktop.app.zip.manifest.json`.
