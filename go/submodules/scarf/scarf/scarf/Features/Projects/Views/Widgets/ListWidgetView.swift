import SwiftUI
import ScarfCore
import ScarfDesign

struct ListWidgetView: View {
    let widget: DashboardWidget

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack(spacing: 4) {
                if let icon = widget.icon {
                    Image(systemName: icon)
                        .foregroundStyle(.secondary)
                        .scarfStyle(.caption)
                }
                Text(widget.title)
                    .scarfStyle(.caption)
                    .foregroundStyle(.secondary)
            }
            if let items = widget.items {
                ForEach(items) { item in
                    ListItemRow(item: item)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(ScarfColor.backgroundSecondary)
        .clipShape(RoundedRectangle(cornerRadius: ScarfRadius.lg))
    }
}

/// One row of a list widget. Maps `item.status` through `ListItemStatus(raw:)`
/// to a typed badge (icon + color). Unknown strings render as plain text with
/// the original string preserved as a trailing badge so nothing's hidden.
struct ListItemRow: View {
    let item: ListItem

    private var typedStatus: ListItemStatus? { ListItemStatus(raw: item.status) }

    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: typedStatus?.iconName ?? "circle")
                .font(.caption2)
                .foregroundStyle(typedStatus?.tint ?? .secondary)
            Text(item.text)
                .font(.callout)
                .strikethrough(typedStatus == .done)
                .foregroundStyle(typedStatus == .done ? .secondary : .primary)
            if typedStatus == nil, let raw = item.status, !raw.isEmpty {
                Text(raw)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                    .padding(.horizontal, 6)
                    .padding(.vertical, 2)
                    .background(.quaternary.opacity(0.5))
                    .clipShape(Capsule())
            }
        }
    }
}

extension ListItemStatus {
    var iconName: String {
        switch self {
        case .success: return "checkmark.circle.fill"
        case .warning: return "exclamationmark.triangle.fill"
        case .danger:  return "xmark.octagon.fill"
        case .info:    return "info.circle.fill"
        case .pending: return "circle.dashed"
        case .done:    return "checkmark.circle.fill"
        case .neutral: return "circle"
        }
    }

    var tint: Color {
        switch self {
        case .success, .done: return ScarfColor.success
        case .warning:        return ScarfColor.warning
        case .danger:         return ScarfColor.danger
        case .info:           return ScarfColor.info
        case .pending:        return .secondary
        case .neutral:        return .secondary
        }
    }
}
