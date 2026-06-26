import Foundation

/// Hermes v0.13 added cross-platform recipient allowlists to the Messaging
/// Gateway. Each platform stores the list under a different YAML key
/// depending on the platform's primary noun for "addressable destination":
///
/// - **`allowed_channels`** — Slack, Mattermost, Google Chat
/// - **`allowed_chats`** — Telegram, DingTalk
/// - **`allowed_rooms`** — Matrix
///
/// `GatewayAllowlistKind` encodes the (platform → key) mapping plus a few
/// presentation hints (placeholder strings, singular noun) so the allowlist
/// editor can render the right copy without the per-platform setup view
/// needing to know the YAML shape.
public enum GatewayAllowlistKind: String, Sendable, Equatable {
    case channels   // -> allowed_channels
    case chats      // -> allowed_chats
    case rooms      // -> allowed_rooms

    /// YAML scalar key segment under top-level `<platform>.<key>`.
    public var yamlKey: String {
        switch self {
        case .channels: return "allowed_channels"
        case .chats:    return "allowed_chats"
        case .rooms:    return "allowed_rooms"
        }
    }

    /// Placeholder copy for the editor's "add row" text field. Picks the
    /// most common identifier shape per platform family — Slack channel IDs
    /// for `channels`, Telegram username/numeric for `chats`, Matrix room
    /// IDs for `rooms`. Users can paste in any platform-specific format the
    /// gateway accepts; this is a hint, not validation.
    public var inputPlaceholder: String {
        switch self {
        case .channels: return "C0123ABCD or #channel-name"
        case .chats:    return "@username or 12345678"
        case .rooms:    return "!RoomId:matrix.org"
        }
    }

    /// Singular noun for prose surfaces ("Add a channel", "1 chat allowed",
    /// "0 rooms"). Capitalization is the caller's responsibility.
    public var noun: String {
        switch self {
        case .channels: return "channel"
        case .chats:    return "chat"
        case .rooms:    return "room"
        }
    }

    /// Plural noun for headings + counts.
    public var pluralNoun: String {
        switch self {
        case .channels: return "channels"
        case .chats:    return "chats"
        case .rooms:    return "rooms"
        }
    }

    /// Map a Hermes platform identifier to the allowlist kind it supports.
    /// Returns `nil` for platforms without a chat/channel/room allowlist
    /// (`cli`, `signal`, `email`, `imessage`, `homeassistant`, `webhook`,
    /// `yuanbao`, `microsoft-teams`, `feishu`, `discord`, `whatsapp`).
    ///
    /// `whatsapp` is intentionally excluded: it gates *senders* via
    /// `allow_from` / `group_allow_from` (active only under
    /// `dm_policy: allowlist` / `group_policy: allowlist`), not an
    /// `allowed_chats` list. Writing `whatsapp.allowed_chats` is a silent
    /// no-op — Hermes never reads it (verified against v0.17
    /// `gateway/platforms/whatsapp.py`). Proper `allow_from` support belongs in
    /// the WhatsApp setup form, not this generic chat-id editor.
    ///
    /// `googlechat` and `google-chat` both map to `.channels` so we round-trip
    /// regardless of which spelling Hermes lands on. // TODO(WS-5-Q1)
    public static func kind(for platform: String) -> GatewayAllowlistKind? {
        switch platform {
        case "slack", "mattermost", "google-chat", "googlechat": return .channels
        case "telegram", "dingtalk":                             return .chats
        case "matrix":                                           return .rooms
        default: return nil
        }
    }
}
