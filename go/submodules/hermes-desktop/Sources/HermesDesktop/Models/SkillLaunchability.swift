import Foundation

struct LaunchableSkillRecord: Equatable, Hashable, Sendable {
    let name: String
    let category: String?
    let source: String
    let status: String

    var launchIdentifier: String {
        guard let category,
              !category.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return name
        }

        return "\(category)/\(name)"
    }
}

enum LaunchableSkillInventoryParser {
    static func parse(_ output: String) -> [LaunchableSkillRecord] {
        output
            .split(whereSeparator: \.isNewline)
            .compactMap(parseRecord)
    }

    static func filterDiscoveredSkills(
        _ discovered: [SkillSummary],
        using launchableRecords: [LaunchableSkillRecord]
    ) -> [SkillSummary] {
        let allowedIdentifiers = Set(launchableRecords.map(\.launchIdentifier))
        return discovered.filter { allowedIdentifiers.contains($0.relativePath) }
    }

    private static func parseRecord(_ rawLine: Substring) -> LaunchableSkillRecord? {
        let line = rawLine.trimmingCharacters(in: .whitespacesAndNewlines)
        guard line.hasPrefix("│"), line.hasSuffix("│") else { return nil }

        let columns = line
            .split(separator: "│", omittingEmptySubsequences: false)
            .dropFirst()
            .dropLast()
            .map { $0.trimmingCharacters(in: .whitespacesAndNewlines) }

        guard columns.count == 5 else { return nil }
        guard columns[0] != "Name" else { return nil }
        guard columns[4] == "enabled" else { return nil }
        guard !columns[0].isEmpty else { return nil }

        let category = columns[1].isEmpty ? nil : columns[1]

        return LaunchableSkillRecord(
            name: columns[0],
            category: category,
            source: columns[2],
            status: columns[4]
        )
    }
}
