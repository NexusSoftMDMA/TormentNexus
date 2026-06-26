import Foundation

/// A Hermes v0.15 skill bundle — a named group of skills loaded together
/// by one `/<name>` slash command.
///
/// Bundles are stored as YAML files at `~/.hermes/skill-bundles/*.yaml`
/// with the schema:
///
/// ```yaml
/// name: backend-dev
/// description: ...        # optional
/// skills: [github-code-review, test-driven-development, ...]
/// instruction: |         # optional
///   extra context appended when the bundle is invoked
/// ```
///
/// The file stem is the fallback `name`, and the bundle is invoked from
/// chat as `/<slug>`. Scarf reads the YAML files directly (rather than
/// parsing `hermes bundles list` text output) so the surface works the
/// same on remote SSH contexts as on local — mirroring how
/// `SkillsScanner` walks `~/.hermes/skills/` through the transport.
public struct HermesSkillBundle: Identifiable, Sendable {
    /// Stable identity. Equal to `name` (which falls back to the file
    /// stem when the YAML omits an explicit `name:`), so it round-trips
    /// through SwiftUI `ForEach` without a separate UUID.
    public let id: String
    /// The bundle name. Falls back to the file stem when the YAML has
    /// no `name:` key.
    public let name: String
    /// Optional human-readable description (`description:` in the YAML).
    /// `nil` when absent or blank.
    public let description: String?
    /// Member skill names (`skills:` list in the YAML). Empty when the
    /// list is absent or malformed.
    public let skills: [String]
    /// Optional extra instruction text (`instruction:` block scalar).
    /// Appended to the agent's context when the bundle is invoked.
    /// `nil` when absent or blank.
    public let instruction: String?

    public init(
        id: String,
        name: String,
        description: String? = nil,
        skills: [String] = [],
        instruction: String? = nil
    ) {
        self.id = id
        self.name = name
        self.description = description
        self.skills = skills
        self.instruction = instruction
    }

    /// The `/<slug>` slash command that invokes this bundle. Derived
    /// from `name` lowercased with whitespace collapsed to hyphens —
    /// matches how Hermes slugifies a bundle name into a command.
    public var slug: String {
        let lowered = name.lowercased().trimmingCharacters(in: .whitespacesAndNewlines)
        let collapsed = lowered
            .split(whereSeparator: { $0 == " " || $0 == "_" || $0 == "\t" })
            .joined(separator: "-")
        return collapsed.isEmpty ? lowered : collapsed
    }

    /// Parse a single bundle YAML document into a `HermesSkillBundle`.
    ///
    /// Tolerant by design: a missing `name:` falls back to `stem`, a
    /// missing/blank `description` or `instruction` becomes `nil`, and a
    /// missing `skills:` list becomes `[]`. Returns `nil` only when the
    /// content is so degenerate that there's nothing to show (no name
    /// resolvable AND no skills) — callers `compactMap` over a directory
    /// so a malformed file is simply skipped, never a crash.
    ///
    /// Supports both block-style (`skills:` then `  - name`) and
    /// inline-flow (`skills: [a, b, c]`) lists for the `skills:` key,
    /// since Hermes-emitted YAML and hand-authored bundles both occur.
    public static func parse(yaml: String, stem: String) -> HermesSkillBundle? {
        let parsed = HermesYAML.parseNestedYAML(yaml)

        let rawName = parsed.values["name"]?.trimmingCharacters(in: .whitespacesAndNewlines)
        let name = (rawName?.isEmpty ?? true)
            ? stem
            : HermesYAML.stripYAMLQuotes(rawName!)

        let rawDescription = parsed.values["description"]?.trimmingCharacters(in: .whitespacesAndNewlines)
        let description: String?
        if let rawDescription, !rawDescription.isEmpty {
            description = HermesYAML.stripYAMLQuotes(rawDescription)
        } else {
            description = nil
        }

        // `skills:` may arrive as a block list (parsed into `lists`) or
        // an inline flow literal `[a, b]` (which the nested parser drops
        // into `values` as the raw bracketed string). Handle both.
        var skills = parsed.lists["skills"] ?? []
        if skills.isEmpty, let inline = parsed.values["skills"] {
            skills = parseInlineList(inline)
        }
        skills = skills
            .map { HermesYAML.stripYAMLQuotes($0).trimmingCharacters(in: .whitespaces) }
            .filter { !$0.isEmpty }

        // `instruction: |` is a block scalar. The nested parser pushes
        // `instruction` as a section header but doesn't capture the
        // indented body, so we recover the body with a dedicated scan.
        let instruction = parseBlockScalar(key: "instruction", from: yaml)

        // Degenerate guard: nothing nameable and no skills → skip.
        if name.isEmpty && skills.isEmpty { return nil }
        let resolvedName = name.isEmpty ? stem : name

        return HermesSkillBundle(
            id: resolvedName,
            name: resolvedName,
            description: description,
            skills: skills,
            instruction: instruction
        )
    }

    /// Parse an inline flow list literal like `[a, b, c]` into its
    /// elements. Leading/trailing brackets are optional (tolerant); each
    /// element is comma-split, quote-stripped, and whitespace-trimmed.
    private static func parseInlineList(_ raw: String) -> [String] {
        var body = raw.trimmingCharacters(in: .whitespaces)
        if body.hasPrefix("[") { body.removeFirst() }
        if body.hasSuffix("]") { body.removeLast() }
        return body
            .split(separator: ",")
            .map { HermesYAML.stripYAMLQuotes($0.trimmingCharacters(in: .whitespaces)) }
            .filter { !$0.isEmpty }
    }

    /// Recover a YAML block scalar body (`key: |` followed by indented
    /// lines) from the raw text. Returns `nil` when the key is absent or
    /// the body is blank. Indentation is de-dented by the smallest
    /// leading-space count across the block so the returned text reads
    /// naturally. Stops at the first line whose indent is <= the key's.
    private static func parseBlockScalar(key: String, from yaml: String) -> String? {
        let lines = yaml.components(separatedBy: "\n")
        guard let headerIdx = lines.firstIndex(where: { line in
            let trimmed = line.trimmingCharacters(in: .whitespaces)
            guard trimmed.hasPrefix("\(key):") else { return false }
            let after = trimmed.dropFirst("\(key):".count).trimmingCharacters(in: .whitespaces)
            // Only treat `|` / `>` (optionally with chomp indicators) or
            // an empty value as a block-scalar header.
            return after.isEmpty || after.hasPrefix("|") || after.hasPrefix(">")
        }) else { return nil }

        let headerIndent = lines[headerIdx].prefix { $0 == " " }.count
        var body: [String] = []
        var idx = headerIdx + 1
        while idx < lines.count {
            let line = lines[idx]
            let trimmed = line.trimmingCharacters(in: .whitespaces)
            if trimmed.isEmpty {
                body.append("")
                idx += 1
                continue
            }
            let indent = line.prefix { $0 == " " }.count
            if indent <= headerIndent { break }
            body.append(line)
            idx += 1
        }

        // Trim trailing blank lines.
        while let last = body.last, last.isEmpty { body.removeLast() }
        guard !body.isEmpty else { return nil }

        // De-dent by the minimum indent of the non-blank lines.
        let minIndent = body
            .filter { !$0.trimmingCharacters(in: .whitespaces).isEmpty }
            .map { $0.prefix { $0 == " " }.count }
            .min() ?? 0
        let dedented = body.map { line -> String in
            guard line.count >= minIndent else { return line }
            return String(line.dropFirst(minIndent))
        }
        let joined = dedented.joined(separator: "\n").trimmingCharacters(in: .whitespacesAndNewlines)
        return joined.isEmpty ? nil : joined
    }
}
