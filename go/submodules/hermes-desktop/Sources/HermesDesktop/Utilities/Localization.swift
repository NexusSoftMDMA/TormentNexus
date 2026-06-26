import Foundation

enum L10n {
    static func string(_ key: String) -> String {
        for bundle in localizationBundles {
            let value = NSLocalizedString(key, bundle: bundle, value: "", comment: "")
            if !value.isEmpty, value != key {
                return value
            }
        }

        return key
    }

    static func string(_ key: String, _ arguments: CVarArg...) -> String {
        String(format: string(key), arguments: arguments)
    }

    private static let localizationBundles: [Bundle] = {
        let resourceBundleName = "HermesDesktop_HermesDesktop.bundle"
        let candidateURLs = [
            Bundle.main.resourceURL?.appendingPathComponent(resourceBundleName),
            Bundle.main.bundleURL.appendingPathComponent(resourceBundleName)
        ].compactMap { $0 }

        let resourceBundles = candidateURLs.compactMap(Bundle.init(url:))
        return [.main] + resourceBundles
    }()
}
