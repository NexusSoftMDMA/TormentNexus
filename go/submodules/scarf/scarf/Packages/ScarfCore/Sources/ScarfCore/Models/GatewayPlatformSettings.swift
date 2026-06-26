import Foundation

/// Per-platform Messaging Gateway settings introduced in Hermes v0.13. Bundles
/// the allowlist (the platform-appropriate flavor of `allowed_channels` /
/// `allowed_chats` / `allowed_rooms`) and three behavior toggles
/// (`busy_ack_enabled`, `gateway_restart_notification`,
/// `slash_command_notice_ttl_seconds`).
///
/// The struct carries all three list fields so a single shape fits every
/// platform; only the field matching `GatewayAllowlistKind.kind(for:)` is
/// surfaced in the editor for a given platform. The other two stay empty
/// and round-trip through the YAML parser unchanged.
///
/// **Defaults track Hermes v0.13.** `busyAckEnabled = true`,
/// `gatewayRestartNotification = false`, `slashCommandNoticeTTLSeconds = 0`
/// (disabled). An "all-default" instance therefore produces no `gateway:`
/// block in YAML — see `HermesConfig+YAML` parsing logic which only inserts
/// an entry into `gatewayPlatforms` when at least one v0.13 key is present
/// in the file.
public struct GatewayPlatformSettings: Sendable, Equatable {
    /// `gateway.platforms.<platform>.allowed_channels` — Slack, Mattermost,
    /// Google Chat. Empty when the platform doesn't use channels.
    public var allowedChannels: [String]
    /// `gateway.platforms.<platform>.allowed_chats` — Telegram, WhatsApp.
    /// Empty when the platform doesn't use chats.
    public var allowedChats: [String]
    /// `gateway.platforms.<platform>.allowed_rooms` — Matrix, DingTalk.
    /// Empty when the platform doesn't use rooms.
    public var allowedRooms: [String]
    /// `gateway.platforms.<platform>.busy_ack_enabled`. Default `true` — set
    /// to `false` to suppress per-message "agent is working…" acks.
    public var busyAckEnabled: Bool
    /// `gateway.platforms.<platform>.gateway_restart_notification`. Default
    /// `false` — set to `true` to post a "Gateway restarted" notice on boot.
    public var gatewayRestartNotification: Bool
    /// `gateway.platforms.<platform>.slash_command_notice_ttl_seconds`.
    /// Default `0` (disabled). Positive values auto-delete slash-command
    /// notices after N seconds.
    public var slashCommandNoticeTTLSeconds: Int

    public init(
        allowedChannels: [String] = [],
        allowedChats: [String] = [],
        allowedRooms: [String] = [],
        busyAckEnabled: Bool = true,
        gatewayRestartNotification: Bool = false,
        slashCommandNoticeTTLSeconds: Int = 0
    ) {
        self.allowedChannels = allowedChannels
        self.allowedChats = allowedChats
        self.allowedRooms = allowedRooms
        self.busyAckEnabled = busyAckEnabled
        self.gatewayRestartNotification = gatewayRestartNotification
        self.slashCommandNoticeTTLSeconds = slashCommandNoticeTTLSeconds
    }

    /// All-default instance. `HermesConfig.empty` initializes
    /// `gatewayPlatforms: [:]` so this is rarely used directly; provided
    /// for symmetry with the other settings types.
    public static let empty = GatewayPlatformSettings()

    /// The list field matching this allowlist kind, or `nil` for
    /// platforms without an allowlist surface.
    public func items(for kind: GatewayAllowlistKind) -> [String] {
        switch kind {
        case .channels: return allowedChannels
        case .chats:    return allowedChats
        case .rooms:    return allowedRooms
        }
    }
}
