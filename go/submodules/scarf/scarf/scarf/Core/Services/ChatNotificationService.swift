import Foundation
import UserNotifications
import os
#if canImport(AppKit)
import AppKit
#endif

/// Posts a "Hermes finished responding" local notification when an
/// agent prompt completes while Scarf is not in the foreground
/// (issue #64). Users can switch to other work and learn when their
/// prompt has landed without polling the chat pane.
///
/// Authorization is requested lazily on first use. The user's global
/// toggle (`scarf.chat.notifyOnComplete`, default on) gates posting,
/// and notifications are suppressed when `NSApp.isActive` so users
/// who happen to be looking at the chat aren't pinged for nothing.
@MainActor
final class ChatNotificationService {
    static let shared = ChatNotificationService()

    private let logger = Logger(subsystem: "com.scarf", category: "ChatNotifications")
    private let center = UNUserNotificationCenter.current()
    private var hasRequestedAuthorization = false
    private var isAuthorized = false

    /// AppStorage-shared key for the "notify on completion" toggle.
    /// Default true; the toggle lives under Settings → Display.
    static let toggleKey = "scarf.chat.notifyOnComplete"

    private init() {}

    /// Post a local notification announcing prompt completion. Quietly
    /// no-ops when:
    ///   - The user has disabled the toggle.
    ///   - Scarf is the foreground app (the in-chat status indicator
    ///     is sufficient).
    ///   - The system has not yet granted (or has denied) notification
    ///     authorization.
    /// `preview` is the first line of the assistant's reply, truncated
    /// to a sensible length for the lock-screen / notification center.
    func postPromptCompleted(sessionTitle: String?, preview: String) {
        let enabled = UserDefaults.standard.object(forKey: Self.toggleKey) as? Bool ?? true
        guard enabled else { return }

        #if canImport(AppKit)
        if NSApp?.isActive == true { return }
        #endif

        Task { [weak self] in
            guard let self else { return }
            let granted = await self.ensureAuthorized()
            guard granted else { return }

            let content = UNMutableNotificationContent()
            content.title = sessionTitle?.isEmpty == false
                ? "Hermes finished — \(sessionTitle ?? "")"
                : "Hermes finished responding"
            content.body = Self.trimmedPreview(preview)
            content.sound = .default

            let request = UNNotificationRequest(
                identifier: UUID().uuidString,
                content: content,
                trigger: nil
            )
            do {
                try await self.center.add(request)
            } catch {
                self.logger.warning("Notification post failed: \(error.localizedDescription, privacy: .public)")
            }
        }
    }

    private func ensureAuthorized() async -> Bool {
        if isAuthorized { return true }
        if hasRequestedAuthorization {
            // Already asked once this run; respect the current settings.
            let settings = await center.notificationSettings()
            isAuthorized = settings.authorizationStatus == .authorized
            return isAuthorized
        }
        hasRequestedAuthorization = true
        do {
            let granted = try await center.requestAuthorization(options: [.alert, .sound])
            isAuthorized = granted
            return granted
        } catch {
            logger.warning("Notification authorization failed: \(error.localizedDescription, privacy: .public)")
            return false
        }
    }

    /// First non-empty line, capped at ~140 chars so the notification
    /// surface stays readable on every macOS notification style.
    static func trimmedPreview(_ raw: String) -> String {
        let firstLine = raw
            .split(whereSeparator: \.isNewline)
            .first
            .map(String.init) ?? raw
        let trimmed = firstLine.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.count <= 140 { return trimmed }
        let prefix = trimmed.prefix(140).trimmingCharacters(in: .whitespacesAndNewlines)
        return prefix + "…"
    }
}
