import Foundation

/// Cross-profile snapshot derived from `hermes gateway list` (Hermes v0.16).
/// Each profile is one configured Messaging Gateway instance — most users
/// have a single `default` profile, but power users keep separate profiles
/// for work / personal / project-specific accounts.
public struct GatewayListSnapshot: Sendable, Equatable {
    public struct ProfileEntry: Sendable, Equatable {
        public let profile: String
        public let isRunning: Bool
        public let pid: Int?
        public let platforms: [String]   // always empty: text output omits it

        public init(
            profile: String,
            isRunning: Bool,
            pid: Int?,
            platforms: [String]
        ) {
            self.profile = profile
            self.isRunning = isRunning
            self.pid = pid
            self.platforms = platforms
        }
    }
    public let profiles: [ProfileEntry]
    public let detectedAt: Date

    public init(profiles: [ProfileEntry], detectedAt: Date = Date()) {
        self.profiles = profiles
        self.detectedAt = detectedAt
    }

    /// One-line digest for the Messaging Gateway page header. Format depends
    /// on shape:
    /// - 0 profiles: `"no profiles configured"`
    /// - 1 profile, running: `"default profile · running · slack, telegram"`
    /// - 1 profile, stopped: `"default profile · stopped"`
    /// - >1 profile: `"3 profiles (2 running) · default: slack, telegram"`
    public var headerDigest: String {
        if profiles.isEmpty { return "no profiles configured" }

        if profiles.count == 1 {
            let p = profiles[0]
            let state = p.isRunning ? "running" : "stopped"
            if p.isRunning && !p.platforms.isEmpty {
                let plats = p.platforms.joined(separator: ", ")
                return "\(p.profile) profile · \(state) · \(plats)"
            }
            return "\(p.profile) profile · \(state)"
        }

        let runningCount = profiles.filter(\.isRunning).count
        // Surface the platforms of the first running profile (or first profile
        // if none are running) so the digest carries one specimen of context
        // beyond just counts.
        let highlight = profiles.first(where: \.isRunning) ?? profiles[0]
        let platsClause: String
        if highlight.platforms.isEmpty {
            platsClause = ""
        } else {
            platsClause = " · \(highlight.profile): \(highlight.platforms.joined(separator: ", "))"
        }
        return "\(profiles.count) profiles (\(runningCount) running)\(platsClause)"
    }
}

/// Pure parser + sync fetcher for `hermes gateway list` (Hermes v0.16).
/// `hermes gateway list` has no `--json` flag — it prints a text table — so
/// the parser reads that text directly. The fetcher returns `nil` on a
/// non-zero exit (host without the subcommand) so the digest row hides
/// itself.
///
/// Expected text shape:
/// ```
/// Gateways:
///   ✓ default (current)        — PID 44417
///   ✗ scarfbox-smoke           — not running
///   ✗ scarfbox-test            — not running
/// ```
/// `✓`/`✗` gives `isRunning`; the word after it is the profile name (a
/// trailing `(current)` marker is stripped); `— PID <n>` (em dash, U+2014)
/// carries the pid on running lines. Text output has no per-profile platform
/// list, so `platforms` is always `[]`.
///
/// The detection is **synchronous** — run from a `Task.detached` to avoid
/// blocking MainActor on remote SSH round-trips. The pure `parse(_:)`
/// helper has no I/O and can be used in tests against canned text.
public enum HermesGatewayListService {

    /// Parse the text table from `hermes gateway list` into a snapshot.
    /// Skips the `Gateways:` header; each subsequent profile line yields a
    /// `ProfileEntry`. `platforms` is always `[]` (the text output omits
    /// it). Returns `nil` for empty / whitespace-only input.
    public static func parse(_ text: String) -> GatewayListSnapshot? {
        let trimmedWhole = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmedWhole.isEmpty else { return nil }

        var entries: [GatewayListSnapshot.ProfileEntry] = []
        for rawLine in text.components(separatedBy: "\n") {
            let line = rawLine.trimmingCharacters(in: .whitespaces)
            if line.isEmpty { continue }
            // Skip the `Gateways:` header (and any other non-profile line
            // that lacks a running marker).
            guard line.hasPrefix("✓") || line.hasPrefix("✗") else { continue }

            let isRunning = line.hasPrefix("✓")

            // Strip the marker, then split off the trailing `— …` clause
            // (em dash, U+2014) which carries pid / status.
            var rest = String(line.dropFirst()).trimmingCharacters(in: .whitespaces)
            var pid: Int?
            if let dashRange = rest.range(of: "—") {
                let after = rest[dashRange.upperBound...]
                    .trimmingCharacters(in: .whitespaces)
                rest = String(rest[..<dashRange.lowerBound])
                    .trimmingCharacters(in: .whitespaces)
                // Running lines read `PID <n>`; stopped lines `not running`.
                if after.hasPrefix("PID") {
                    let digits = after.drop(while: { !$0.isNumber })
                    pid = Int(digits.prefix(while: { $0.isNumber }))
                }
            }

            // The profile name is the first whitespace-delimited token; a
            // trailing `(current)` marker is a separate token, so dropping
            // everything after the first space removes it.
            let profile = rest
                .split(whereSeparator: { $0 == " " || $0 == "\t" })
                .first
                .map(String.init) ?? ""
            guard !profile.isEmpty else { continue }

            entries.append(GatewayListSnapshot.ProfileEntry(
                profile: profile,
                isRunning: isRunning,
                pid: pid,
                platforms: []
            ))
        }

        // No recognizable profile lines (e.g. garbage input) → nil.
        guard !entries.isEmpty else { return nil }
        return GatewayListSnapshot(profiles: entries)
    }

    /// Synchronous fetch helper — call from a `Task.detached`. Returns
    /// `nil` when the subcommand fails (host without `gateway list`) or when
    /// the output has no recognizable profile lines.
    public static func fetch(context: ServerContext) -> GatewayListSnapshot? {
        let transport = context.makeTransport()
        let executable = context.paths.hermesBinary
        do {
            let result = try transport.runProcess(
                executable: executable,
                args: ["gateway", "list"],
                stdin: nil,
                timeout: 10
            )
            guard result.exitCode == 0 else { return nil }
            return parse(result.stdoutString)
        } catch {
            return nil
        }
    }
}
