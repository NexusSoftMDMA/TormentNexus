import AppKit
import SwiftUI

@main
struct HermesDesktopApp: App {
    @NSApplicationDelegateAdaptor(HermesApplicationDelegate.self) private var appDelegate
    @StateObject private var appState = AppState()

    var body: some Scene {
        WindowGroup("Hermes Desktop") {
            RootView()
                .environmentObject(appState)
                .frame(minWidth: HermesSplitMetrics.minimumWindowWidth, minHeight: 520)
                .background(
                    HermesWindowTitleBarConfigurator(
                        backgroundImageActive: appState.connectionStore.isBackgroundImageActive,
                        windowOpacity: appState.connectionStore.windowOpacity,
                        windowMaterial: appState.connectionStore.windowMaterial
                    )
                )
        }
        .defaultSize(width: 1360, height: 860)
        .commands {
            HermesDesktopCommands(appState: appState)
        }
    }
}
