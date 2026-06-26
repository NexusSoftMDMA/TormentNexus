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
                        .foregroundStyle(ScarfColor.foregroundMuted)
                        .font(.caption)
                }
                Text(widget.title)
                    .font(.caption)
                    .foregroundStyle(ScarfColor.foregroundMuted)
            }
            if let items = widget.items {
                ForEach(items) { item in
                    ListItemRow(item: item)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(.quaternary.opacity(0.5))
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

private struct ListItemRow: View {
    let item: ListItem

    private var typedStatus: ListItemStatus? { ListItemStatus(raw: item.status) }

    var body: some View {
        HStack(spacing: 6) {
            Image(systemName: iconName)
                .font(.caption2)
                .foregroundStyle(tint)
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

    private var iconName: String {
        switch typedStatus {
        case .success, .done: return "checkmark.circle.fill"
        case .warning:        return "exclamationmark.triangle.fill"
        case .danger:         return "xmark.octagon.fill"
        case .info:           return "info.circle.fill"
        case .pending:        return "circle.dashed"
        case .neutral, nil:   return "circle"
        }
    }

    private var tint: Color {
        switch typedStatus {
        case .success, .done: return ScarfColor.success
        case .warning:        return ScarfColor.warning
        case .danger:         return ScarfColor.danger
        case .info:           return ScarfColor.info
        case .pending, .neutral, nil: return .secondary
        }
    }
}
