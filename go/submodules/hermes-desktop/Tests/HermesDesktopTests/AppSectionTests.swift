import Testing

@testable import HermesDesktop

struct AppSectionTests {
    @Test
    func navigationShortcutsStayStableAcrossCommandVisibleSections() {
        #expect(AppSection.navigationCases == [.connections] + AppSection.customizableSidebarSections)
        #expect(AppSection.connections.navigationShortcutKey == "1")
        #expect(AppSection.sessions.navigationShortcutKey == "3")
        #expect(AppSection.workflows.navigationShortcutKey == "4")
        #expect(AppSection.cronjobs.navigationShortcutKey == "5")
        #expect(AppSection.kanban.navigationShortcutKey == "6")
        #expect(AppSection.files.navigationShortcutKey == "7")
        #expect(AppSection.usage.navigationShortcutKey == "8")
        #expect(AppSection.skills.navigationShortcutKey == "9")
        #expect(AppSection.terminal.navigationShortcutKey == "0")
    }
}
