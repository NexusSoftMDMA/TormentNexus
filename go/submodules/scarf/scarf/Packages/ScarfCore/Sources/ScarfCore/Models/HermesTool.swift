import Foundation

public struct HermesToolset: Identifiable, Sendable {
    public var id: String { name }
    public let name: String
    public let description: String
    public let icon: String
    public var enabled: Bool

    public init(
        name: String,
        description: String,
        icon: String,
        enabled: Bool
    ) {
        self.name = name
        self.description = description
        self.icon = icon
        self.enabled = enabled
    }
}

public struct HermesToolPlatform: Identifiable, Sendable {
    public var id: String { name }
    public let name: String
    public let displayName: String
    public let icon: String

    public init(
        name: String,
        displayName: String,
        icon: String
    ) {
        self.name = name
        self.displayName = displayName
        self.icon = icon
    }
}

public enum KnownPlatforms {
    public static let cli = HermesToolPlatform(name: "cli", displayName: "CLI", icon: "terminal")
    public static let all: [HermesToolPlatform] = [
        cli,
        HermesToolPlatform(name: "telegram", displayName: "Telegram", icon: "paperplane"),
        HermesToolPlatform(name: "discord", displayName: "Discord", icon: "bubble.left.and.bubble.right"),
        HermesToolPlatform(name: "slack", displayName: "Slack", icon: "number"),
        HermesToolPlatform(name: "whatsapp", displayName: "WhatsApp", icon: "phone.bubble"),
        HermesToolPlatform(name: "signal", displayName: "Signal", icon: "lock.shield"),
        HermesToolPlatform(name: "email", displayName: "Email", icon: "envelope"),
        HermesToolPlatform(name: "homeassistant", displayName: "Home Assistant", icon: "house"),
        HermesToolPlatform(name: "webhook", displayName: "Webhook", icon: "arrow.up.right.square"),
        HermesToolPlatform(name: "matrix", displayName: "Matrix", icon: "lock.rectangle.stack"),
        HermesToolPlatform(name: "feishu", displayName: "Feishu", icon: "message.badge.circle"),
        HermesToolPlatform(name: "mattermost", displayName: "Mattermost", icon: "bubble.left.and.exclamationmark.bubble.right"),
        HermesToolPlatform(name: "imessage", displayName: "iMessage", icon: "message.fill"),
        // -- v0.12 additions ---------------------------------------------
        // Yuanbao is a native gateway adapter (18th platform); Microsoft
        // Teams ships as a plugin (19th). PlatformDetail surfaces the
        // distinction in the setup copy. Names match Hermes's gateway
        // platform identifiers.
        HermesToolPlatform(name: "yuanbao", displayName: "Yuanbao 元宝", icon: "bubble.left.and.bubble.right.fill"),
        HermesToolPlatform(name: "microsoft-teams", displayName: "Microsoft Teams", icon: "person.2.fill"),
        // -- v0.13 additions ---------------------------------------------
        // Google Chat is the 20th gateway platform. It's a generic
        // `env_enablement_fn` / `cron_deliver_env_var`-driven adapter; setup
        // runs through `hermes setup` rather than per-field forms because
        // the auth dance is OAuth-style and lives outside Scarf. Identifier
        // is `google-chat` (kebab-case, mirroring `microsoft-teams`).
        // TODO(WS-5-Q1): verify identifier against Hermes v0.13 GA — if it
        // ships as `googlechat` instead, update both this entry and
        // `KnownPlatforms.icon(for:)` below. `GatewayAllowlistKind.kind(for:)`
        // already accepts both spellings defensively.
        HermesToolPlatform(name: "google-chat", displayName: "Google Chat", icon: "bubble.left.fill"),
        // -- v0.14 additions ---------------------------------------------
        // LINE Messaging API (21st platform, first-class native adapter)
        // and SimpleX Chat (22nd platform, talks to a local
        // `simplex-chat` daemon in WebSocket mode). Identifiers match
        // Hermes's gateway platform names verbatim.
        HermesToolPlatform(name: "line", displayName: "LINE", icon: "bubble.left.and.text.bubble.right"),
        HermesToolPlatform(name: "simplex", displayName: "SimpleX Chat", icon: "lock.shield.fill"),
        // -- v0.15 additions ---------------------------------------------
        // ntfy (23rd platform) — pub/sub push via an ntfy.sh-compatible
        // server. Outbound-capable with an optional separate publish
        // topic; auth is an optional bearer token or `user:pass` Basic.
        // Identifier matches Hermes's gateway platform name verbatim.
        HermesToolPlatform(name: "ntfy", displayName: "ntfy", icon: "bell.badge"),
        // -- v0.17 additions ---------------------------------------------
        // WhatsApp Business Cloud API (25th platform) — Meta's hosted webhook
        // path, distinct from the older `whatsapp` web-bridge. iMessage via
        // Photon (24th) is intentionally not surfaced yet (moving protocol).
        HermesToolPlatform(name: "whatsapp_cloud", displayName: "WhatsApp Cloud", icon: "phone.bubble.fill"),
    ]

    public static func icon(for platform: String) -> String {
        switch platform {
        case "cli": return "terminal"
        case "telegram": return "paperplane"
        case "discord": return "bubble.left.and.bubble.right"
        case "slack": return "number"
        case "whatsapp": return "phone.bubble"
        case "signal": return "lock.shield"
        case "email": return "envelope"
        case "homeassistant": return "house"
        case "webhook": return "arrow.up.right.square"
        case "matrix": return "lock.rectangle.stack"
        case "feishu": return "message.badge.circle"
        case "mattermost": return "bubble.left.and.exclamationmark.bubble.right"
        case "imessage": return "message.fill"
        case "yuanbao": return "bubble.left.and.bubble.right.fill"
        case "microsoft-teams": return "person.2.fill"
        case "google-chat", "googlechat": return "bubble.left.fill"
        case "line": return "bubble.left.and.text.bubble.right"
        case "simplex": return "lock.shield.fill"
        case "ntfy": return "bell.badge"
        case "whatsapp_cloud": return "phone.bubble.fill"
        default: return "bubble.left"
        }
    }
}
