# Hermes Desktop v1.0.0

Hermes Desktop 1.0 is a stability and trust release for the native macOS app.

The app keeps its direct SSH-first model: no gateway API, no host daemon, no
local mirror, and no custom sync layer. Your selected Hermes host remains the
source of truth.

This release also refreshes the app icon, so Hermes Desktop may look different
in your Dock and Applications folder.

Highlights:

- redesigned Settings for hosts, profiles, diagnostics, sidebar preferences,
  app appearance, terminal theme, and update checks
- added app theme preferences and richer terminal surface customization for
  colors and font size
- removed the old Overview surface and experimental Desktop chat backend
- polished the main app views across Sessions, Files, Skills, Cron, Kanban,
  Usage, Terminal, and Workflows
- improved release packaging with manifest and verification checks

Hermes Desktop is still ad-hoc signed and not notarized by Apple. On first
launch, macOS may require right-click > Open.
