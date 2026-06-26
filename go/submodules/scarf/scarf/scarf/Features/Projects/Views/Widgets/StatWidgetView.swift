import SwiftUI
import ScarfCore
import ScarfDesign

/// Tiny inline trend line drawn under a `stat` widget's value. Pure SwiftUI
/// `Path`, no Swift Charts dependency — stays light enough to render
/// dozens per dashboard without measurable cost.
struct SparklineView: View {
    let values: [Double]
    let tint: Color

    var body: some View {
        GeometryReader { geo in
            let minV = values.min() ?? 0
            let maxV = values.max() ?? 1
            let span = max(0.0001, maxV - minV)
            let stepX = values.count > 1 ? geo.size.width / CGFloat(values.count - 1) : 0
            Path { path in
                for (i, v) in values.enumerated() {
                    let x = CGFloat(i) * stepX
                    let normalized = (v - minV) / span
                    let y = geo.size.height - CGFloat(normalized) * geo.size.height
                    if i == 0 { path.move(to: CGPoint(x: x, y: y)) }
                    else { path.addLine(to: CGPoint(x: x, y: y)) }
                }
            }
            .stroke(tint.opacity(0.85), lineWidth: 1.2)
        }
    }
}

struct StatWidgetView: View {
    let widget: DashboardWidget

    private var widgetColor: Color {
        parseColor(widget.color)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 4) {
                if let icon = widget.icon {
                    Image(systemName: icon)
                        .foregroundStyle(widgetColor)
                        .scarfStyle(.caption)
                }
                Text(widget.title)
                    .scarfStyle(.caption)
                    .foregroundStyle(.secondary)
            }
            if let value = widget.value {
                Text(value.displayString)
                    .font(.system(.title2, design: .monospaced, weight: .semibold))
            }
            if let subtitle = widget.subtitle {
                Text(subtitle)
                    .font(.caption2)
                    .foregroundStyle(widgetColor)
            }
            if let sparkline = widget.sparkline, sparkline.count >= 2 {
                SparklineView(values: sparkline, tint: widgetColor)
                    .frame(height: 18)
                    .padding(.top, 2)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(ScarfColor.backgroundSecondary)
        .clipShape(RoundedRectangle(cornerRadius: ScarfRadius.lg))
    }
}
