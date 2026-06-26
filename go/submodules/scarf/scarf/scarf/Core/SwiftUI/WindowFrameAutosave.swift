import AppKit
import SwiftUI

/// Persist a SwiftUI `WindowGroup` window's frame (size + position) across
/// app launches by hooking into AppKit's `NSWindow.setFrameAutosaveName`.
///
/// **Why this exists.** SwiftUI's `WindowGroup` exposes `.defaultSize`,
/// `.windowResizability`, and (on macOS Sonoma+) various scene modifiers
/// — but not a "remember this window's size between launches" affordance.
/// Apple's documented escape hatch is AppKit's `setFrameAutosaveName(_:)`,
/// which writes the window's frame to UserDefaults on resize/move and
/// reads it back on next `makeKey`. We bridge into it from SwiftUI via an
/// invisible `NSViewRepresentable` that finds the hosting `NSWindow`
/// and stamps the autosave name once it appears.
///
/// **Usage.**
///     ContentView()
///         .windowFrameAutosave("Scarf.\(context.id)")
///
/// Pass a stable identifier per logical window. Different identifiers per
/// window are required by AppKit ("no two windows can be associated with
/// the same name simultaneously" — `NSWindow.setFrameAutosaveName(_:)`
/// docs). For Scarf's multi-window-per-server model, keying off
/// `ServerID` gives each server window its own remembered frame.
///
/// **First-launch behaviour.** No saved frame exists → AppKit leaves the
/// window at whatever frame SwiftUI's `.defaultSize` produced. After the
/// first user resize, AppKit autosaves and subsequent opens restore the
/// new frame.
///
/// **What it doesn't do.** Doesn't capture/restore fullscreen state
/// (AppKit handles that separately and reasonably). Doesn't try to
/// override window state restoration when the user has the system-level
/// "Close windows when quitting an application" setting OFF — that
/// pathway runs first and we just ride alongside.
struct WindowFrameAutosave: NSViewRepresentable {
    let name: String

    func makeNSView(context: Context) -> NSView {
        let view = NSView(frame: .zero)
        // The hosting NSWindow isn't attached to this view yet at
        // makeNSView time — SwiftUI mounts the AppKit view hierarchy
        // before the window assignment propagates. Defer one runloop
        // iteration so `view.window` is non-nil when we stamp.
        DispatchQueue.main.async { [weak view] in
            view?.window?.setFrameAutosaveName(name)
        }
        return view
    }

    func updateNSView(_ nsView: NSView, context: Context) {
        // SwiftUI may swap the host window in rare cases (window
        // restoration after a relaunch, scene reuse). Re-stamp on
        // update so we don't lose the autosave binding silently.
        // setFrameAutosaveName is idempotent for the same name on
        // the same window; assigning the same name twice is a no-op.
        DispatchQueue.main.async { [weak nsView] in
            guard let window = nsView?.window else { return }
            if window.frameAutosaveName != name {
                window.setFrameAutosaveName(name)
            }
        }
    }
}

extension View {
    /// Persist this view's hosting window's frame (size + position)
    /// across launches under `name`. See `WindowFrameAutosave` for
    /// details.
    func windowFrameAutosave(_ name: String) -> some View {
        background(WindowFrameAutosave(name: name))
    }
}
