# Hermes Desktop v1.1.0

Hermes Desktop 1.1.0 is the appearance release. Your SSH-first workspace is the
same dependable Hermes, now with more room to make it feel like yours.

You can choose Light, Dark, or System appearance, tune the window opacity, use a
translucent material, and add an optional background image behind the workbench
and terminal. Terminal and Sessions Chat now share font family and size
preferences, so long work sessions can feel clearer, calmer, and more personal.

This release also polishes the Settings experience and split-pane behavior. Host
and profile controls are denser, resizing feels steadier, panels keep practical
minimum widths, and transparent terminal backgrounds stay transparent when a
background image is active.

Fixes:

- select the first saved host automatically when no active host preference
  exists
- repair stale active-host preferences on launch
- make a newly saved first host active immediately
- select the next available host after deleting the active host

Thanks to @Yunle-Lee for sharing early PRs around background images and terminal
scrolling. The appearance customization landed through a separate
implementation, and PR #56 helped shape the terminal scrolling fix included in
this release.

Hermes Desktop 1.1.0 keeps the same distribution model as v1.0.0: a universal
macOS app bundle, ad-hoc signed, not notarized by Apple, with zip checksum and
release manifest assets.
