import Foundation
import Testing
@testable import HermesDesktop

struct LocalizationCoverageTests {
    @Test
    func localizedTablesStayInSync() throws {
        let tables = try Self.localizationTables()
        let english = try #require(tables["en"])
        let englishKeys = Set(english.entries.keys)

        for locale in Self.supportedLocales {
            let table = try #require(tables[locale])

            #expect(table.duplicates.isEmpty, "\(locale) has duplicate localization keys: \(table.duplicates.joined(separator: ", "))")
            #expect(Set(table.entries.keys) == englishKeys, "\(locale) localization keys differ from en")

            for key in englishKeys {
                let englishValue = try #require(english.entries[key]?.value)
                let localizedValue = try #require(table.entries[key]?.value)
                #expect(
                    Self.placeholderSignature(englishValue) == Self.placeholderSignature(localizedValue),
                    "\(locale) placeholder mismatch for key: \(key)"
                )
            }
        }
    }

    @Test
    func directL10nStringKeysExistInEnglishTable() throws {
        let english = try #require(Self.localizationTables()["en"])
        let keys = try Self.directL10nKeys()
        let missing = keys.subtracting(english.entries.keys).sorted()

        #expect(missing.isEmpty, "Missing English localization keys: \(missing.joined(separator: ", "))")
    }

    @Test
    func wrapperLocalizedKeysExistInEnglishTable() throws {
        let english = try #require(Self.localizationTables()["en"])
        let keys = try Self.wrapperL10nKeys()
        let missing = keys.subtracting(english.entries.keys).sorted()

        #expect(missing.isEmpty, "Missing English wrapper localization keys: \(missing.joined(separator: ", "))")
    }

    @Test
    func dynamicDisplayKeysExistInEveryLocalizationTable() throws {
        let tables = try Self.localizationTables()
        let dynamicKeys = Self.dynamicDisplayKeys()

        for locale in Self.supportedLocales {
            let table = try #require(tables[locale])
            let missing = dynamicKeys.subtracting(table.entries.keys).sorted()

            #expect(missing.isEmpty, "\(locale) is missing dynamic display keys: \(missing.joined(separator: ", "))")
        }
    }

    private static let supportedLocales = ["en", "zh-Hans", "ru"]

    private struct LocalizationTable {
        var entries: [String: Entry] = [:]
        var duplicates: [String] = []
    }

    private struct Entry {
        let value: String
        let line: Int
    }

    private static func localizationTables() throws -> [String: LocalizationTable] {
        var tables: [String: LocalizationTable] = [:]

        for locale in supportedLocales {
            let url = projectRoot
                .appendingPathComponent("Sources/HermesDesktop/Resources")
                .appendingPathComponent("\(locale).lproj")
                .appendingPathComponent("Localizable.strings")
            tables[locale] = try parseLocalizationTable(at: url)
        }

        return tables
    }

    private static func parseLocalizationTable(at url: URL) throws -> LocalizationTable {
        let text = try String(contentsOf: url, encoding: .utf8)
        let lines = text.split(separator: "\n", omittingEmptySubsequences: false)
        let pattern = #"^\s*"((?:\\.|[^"\\])*)"\s*=\s*"((?:\\.|[^"\\])*)"\s*;"#
        let regex = try NSRegularExpression(pattern: pattern)
        var table = LocalizationTable()

        for (offset, line) in lines.enumerated() {
            let lineText = String(line)
            let range = NSRange(lineText.startIndex..<lineText.endIndex, in: lineText)
            guard let match = regex.firstMatch(in: lineText, range: range),
                  let keyRange = Range(match.range(at: 1), in: lineText),
                  let valueRange = Range(match.range(at: 2), in: lineText) else {
                continue
            }

            let key = String(lineText[keyRange])
            let value = String(lineText[valueRange])
            if table.entries[key] != nil {
                table.duplicates.append(key)
            }
            table.entries[key] = Entry(value: value, line: offset + 1)
        }

        return table
    }

    private static func directL10nKeys() throws -> Set<String> {
        let sourcesURL = projectRoot.appendingPathComponent("Sources/HermesDesktop")
        let sourceFiles = try swiftFiles(under: sourcesURL)
        let pattern = #"L10n\.string\(\s*"((?:\\.|[^"\\])*)""#
        let regex = try NSRegularExpression(pattern: pattern)
        var keys = Set<String>()

        for url in sourceFiles {
            let text = try String(contentsOf: url, encoding: .utf8)
            let range = NSRange(text.startIndex..<text.endIndex, in: text)
            for match in regex.matches(in: text, range: range) {
                guard let keyRange = Range(match.range(at: 1), in: text) else { continue }
                keys.insert(String(text[keyRange]))
            }
        }

        return keys
    }

    private static func wrapperL10nKeys() throws -> Set<String> {
        let patterns = [
            #"(?:HermesInspectorField|HermesLabeledValue|HermesLoadingState)\s*\((?s:.{0,500}?)\blabel:\s*"((?:\\.|[^"\\])*)""#,
            #"(?:KanbanFormField|CronFormField|SkillFormField|EditorField|ExampleValueRow)\s*\(\s*label:\s*"((?:\\.|[^"\\])*)""#,
            #"(?:HermesSurfacePanel|HermesPageHeader)\s*\((?s:.{0,500}?)\b(?:title|subtitle):\s*"((?:\\.|[^"\\])*)""#
        ]
        let sourcesURL = projectRoot.appendingPathComponent("Sources/HermesDesktop")
        let sourceFiles = try swiftFiles(under: sourcesURL)
        var keys = Set<String>()

        for pattern in patterns {
            let regex = try NSRegularExpression(pattern: pattern)

            for url in sourceFiles {
                let text = try String(contentsOf: url, encoding: .utf8)
                let range = NSRange(text.startIndex..<text.endIndex, in: text)
                for match in regex.matches(in: text, range: range) {
                    guard let keyRange = Range(match.range(at: 1), in: text) else { continue }
                    keys.insert(String(text[keyRange]))
                }
            }
        }

        return keys
    }

    private static func swiftFiles(under url: URL) throws -> [URL] {
        let resourceKeys: Set<URLResourceKey> = [.isRegularFileKey]
        let enumerator = FileManager.default.enumerator(
            at: url,
            includingPropertiesForKeys: Array(resourceKeys)
        )
        var urls: [URL] = []

        while let fileURL = enumerator?.nextObject() as? URL {
            let values = try fileURL.resourceValues(forKeys: resourceKeys)
            if values.isRegularFile == true, fileURL.pathExtension == "swift" {
                urls.append(fileURL)
            }
        }

        return urls
    }

    private static func dynamicDisplayKeys() -> Set<String> {
        var keys = Set<String>()

        keys.formUnion([
            "Connections",
            "Overview",
            "Files",
            "Sessions",
            "Workflows",
            "Cron Jobs",
            "Kanban",
            "Usage",
            "Skills",
            "Terminal",
            "New Job",
            "New Task",
            "New Chat",
            "New Workflow",
            "New Skill",
            "Add File",
            "Create a new cron job",
            "Create a Kanban task",
            "Search skills",
            "Active",
            "Paused",
            "Running",
            "Failed",
            "Error",
            "Triage",
            "Todo",
            "Ready",
            "Blocked",
            "Done",
            "Archived",
            "Scratch",
            "Worktree",
            "Directory",
            "No schedule",
            "Once in %@ %@",
            "Every %@ %@",
            "Once on %@",
            "Every hour at :%@",
            "Every day at %@",
            "Every weekday at %@",
            "Every %@ at %@",
            "On day %@ of every month at %@",
            "minute",
            "minutes",
            "hour",
            "hours",
            "day",
            "days",
            "Sun",
            "Mon",
            "Tue",
            "Wed",
            "Thu",
            "Fri",
            "Sat"
        ])

        keys.formUnion(CronSchedulePreset.allCases.map(\.title))
        keys.formUnion(CronIntervalUnit.allCases.map(\.title))
        keys.formUnion(CronDeliveryPreset.allCases.map(\.title))
        keys.formUnion(CronScheduleFormatter.weekdayPickerLabels)

        return keys
    }

    private static func placeholderSignature(_ value: String) -> [String] {
        let pattern = #"%(?:\d+\$)?[@dfiucs]"#
        guard let regex = try? NSRegularExpression(pattern: pattern) else { return [] }
        let range = NSRange(value.startIndex..<value.endIndex, in: value)

        return regex.matches(in: value, range: range).compactMap { match in
            Range(match.range, in: value).map { String(value[$0]) }
        }
    }

    private static let projectRoot: URL = {
        URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .deletingLastPathComponent()
    }()
}
