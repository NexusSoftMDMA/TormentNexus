import Foundation

extension ISO8601DateFormatter {
    static func fractionalSecondsFormatter() -> ISO8601DateFormatter {
        let formatter = ISO8601DateFormatter()
        formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return formatter
    }
}

enum DateFormatters {
    static func relativeFormatter() -> RelativeDateTimeFormatter {
        let cacheKey = "HermesDesktop.relativeFormatter"
        if let formatter = Thread.current.threadDictionary[cacheKey] as? RelativeDateTimeFormatter {
            return formatter
        }

        let formatter = RelativeDateTimeFormatter()
        formatter.unitsStyle = .short
        Thread.current.threadDictionary[cacheKey] = formatter
        return formatter
    }

    static func shortDateTimeFormatter() -> DateFormatter {
        let cacheKey = "HermesDesktop.shortDateTimeFormatter"
        if let formatter = Thread.current.threadDictionary[cacheKey] as? DateFormatter {
            return formatter
        }

        let formatter = DateFormatter()
        formatter.dateStyle = .medium
        formatter.timeStyle = .short
        Thread.current.threadDictionary[cacheKey] = formatter
        return formatter
    }

    static func shortDateTimeString(from date: Date) -> String {
        shortDateTimeFormatter().string(from: date)
    }
}
