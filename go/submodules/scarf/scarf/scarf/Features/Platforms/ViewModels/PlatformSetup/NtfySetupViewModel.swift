import Foundation
import ScarfCore
import os

/// ntfy setup (Hermes v0.15, 23rd platform). Pub/sub push via an
/// ntfy.sh-compatible server.
///
/// `topic` + `server` are settable via env (`NTFY_TOPIC` /
/// `NTFY_SERVER_URL`), which win over config.yaml. `publish_topic`,
/// `token`, and `markdown` live under `platforms.ntfy.extra` in config.yaml.
/// `token` is a bearer token, or `user:pass` for HTTP Basic auth.
///
/// Field reference: https://hermes-agent.nousresearch.com/docs/user-guide/messaging/ntfy
@Observable
@MainActor
final class NtfySetupViewModel {
    let context: ServerContext
    init(context: ServerContext = .local) { self.context = context }

    // Required
    var topic: String = ""
    // Optional
    var server: String = "https://ntfy.sh"
    var publishTopic: String = ""
    var token: String = ""
    var markdown: Bool = false

    var message: String?

    func load() {
        let env = HermesEnvService(context: context).load()
        let cfg = HermesFileService(context: context).loadConfig().ntfy

        // env wins over config.yaml for topic + server.
        topic = env["NTFY_TOPIC"] ?? cfg.topic
        server = env["NTFY_SERVER_URL"] ?? (cfg.server.isEmpty ? "https://ntfy.sh" : cfg.server)
        publishTopic = cfg.publishTopic
        token = cfg.token
        markdown = cfg.markdown
    }

    func save() {
        let envPairs: [String: String] = [
            "NTFY_TOPIC": topic,
            // Don't persist the default server as an env override.
            "NTFY_SERVER_URL": server == "https://ntfy.sh" ? "" : server
        ]
        let configKV: [String: String] = [
            "platforms.ntfy.extra.topic": topic,
            "platforms.ntfy.extra.server": server,
            "platforms.ntfy.extra.publish_topic": publishTopic,
            "platforms.ntfy.extra.token": token,
            "platforms.ntfy.extra.markdown": PlatformSetupHelpers.envBool(markdown)
        ]
        message = PlatformSetupHelpers.saveForm(context: context, envPairs: envPairs, configKV: configKV)
        DispatchQueue.main.asyncAfter(deadline: .now() + 3) { [weak self] in
            self?.message = nil
        }
    }
}
