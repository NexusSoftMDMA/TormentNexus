import SwiftUI
import ScarfDesign

/// Replacement for the legacy "Unknown widget" placeholder. Surfaces the
/// widget's own title plus a structured reason so dashboard authors can see
/// at a glance what's wrong (unknown type, missing file, parse error, …).
///
/// Used by the `WidgetView` dispatcher's default branch and (in v2.7+) by
/// file-reading widgets that can't load their underlying data.
struct WidgetErrorCard: View {
    let title: String
    let reason: String
    var hint: String? = nil

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 6) {
                Image(systemName: "exclamationmark.triangle.fill")
                    .foregroundStyle(ScarfColor.warning)
                    .font(.caption)
                Text(title.isEmpty ? "Widget error" : title)
                    .scarfStyle(.caption)
                    .foregroundStyle(.secondary)
            }
            Text(reason)
                .font(.callout)
                .foregroundStyle(.primary)
                .fixedSize(horizontal: false, vertical: true)
            if let hint, !hint.isEmpty {
                Text(hint)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(ScarfColor.warning.opacity(0.08))
        .overlay(
            RoundedRectangle(cornerRadius: ScarfRadius.lg)
                .strokeBorder(ScarfColor.warning.opacity(0.3), lineWidth: 1)
        )
        .clipShape(RoundedRectangle(cornerRadius: ScarfRadius.lg))
    }
}
