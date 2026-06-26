# Hermes Desktop v1.0.1

Hermes Desktop 1.0.1 fixes a host selection regression introduced in the
v1.0.0 Settings refresh.

On a fresh install, saving the first host could write `connections.json`
successfully but leave `preferences.json` without an active `lastConnectionID`.
The host was saved locally, but the Settings UI still behaved as if no host was
available.

Fixes:

- select the first saved host automatically when no active host preference
  exists
- repair stale active-host preferences on launch
- make a newly saved first host active immediately
- select the next available host after deleting the active host

This release keeps the same distribution model as v1.0.0: a universal macOS app
bundle, ad-hoc signed, not notarized by Apple, with zip checksum and release
manifest assets.
