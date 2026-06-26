import AppKit

@MainActor
final class HermesApplicationDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ _: Notification) {
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
        NSWindow.allowsAutomaticWindowTabbing = false
    }
}
