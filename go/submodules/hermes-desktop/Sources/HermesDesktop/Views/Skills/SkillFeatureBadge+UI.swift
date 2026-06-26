import SwiftUI

extension SkillFeatureBadge {
    var color: Color {
        switch self {
        case .references:
            return .blue
        case .scripts:
            return .green
        case .templates:
            return .orange
        }
    }
}
