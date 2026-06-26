import SwiftUI
import ScarfCore
import ScarfDesign
import UIKit

/// In-app Diagnostics → Performance panel. Lets users flip the
/// ScarfMon backend mode, watch live aggregated stats from the ring
/// buffer, and copy a JSON dump to paste into a feedback thread.
///
/// Data never leaves the device unless the user taps "Copy as JSON" —
/// no remote upload, no analytics. Same source-of-truth as the Mac
/// panel; both sides read `ScarfMonBoot.sharedRingBuffer`.
struct ScarfMonDiagnosticsView: View {
    @State private var mode: ScarfMonBoot.Mode = ScarfMonBoot.currentMode()
    @State private var stats: [ScarfMonStat] = []
    @State private var copiedToast: Bool = false

    /// Ring buffer is process-wide; we read from it on a 1s timer
    /// while the panel is foregrounded. No live tail; this view only
    /// re-aggregates the in-memory snapshot.
    private let refreshInterval: TimeInterval = 1.0

    var body: some View {
        List {
            modeSection
            if mode == .full {
                summarySection
                actionsSection
            } else {
                Section {
                    Text("Switch to **Full** above to see live stats and copy a JSON dump. Off and Signpost-only modes don't keep an in-memory ring buffer.")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                }
            }
        }
        .navigationTitle("Performance")
        .navigationBarTitleDisplayMode(.inline)
        .task(id: mode) {
            // Re-aggregate while the view is visible. SwiftUI cancels
            // this task on disappear, so the timer stops eating cycles
            // when the user backs out.
            guard mode == .full else { return }
            while !Task.isCancelled {
                refresh()
                try? await Task.sleep(nanoseconds: UInt64(refreshInterval * 1_000_000_000))
            }
        }
        .overlay(alignment: .top) {
            if copiedToast {
                Text("Copied to clipboard")
                    .font(.caption)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .background(.regularMaterial)
                    .clipShape(Capsule())
                    .padding(.top, 8)
            }
        }
    }

    @ViewBuilder
    private var modeSection: some View {
        Section {
            Picker("Mode", selection: $mode) {
                Text("Off").tag(ScarfMonBoot.Mode.off)
                Text("Signpost only").tag(ScarfMonBoot.Mode.signpostOnly)
                Text("Full").tag(ScarfMonBoot.Mode.full)
            }
            .pickerStyle(.segmented)
            .onChange(of: mode) { _, newValue in
                ScarfMonBoot.setMode(newValue)
            }
        } header: {
            Text("Recording mode")
        } footer: {
            Text("**Signpost only** is the default — Instruments can attach and read the Points of Interest track without any other overhead. **Full** also keeps a 4096-entry in-memory ring you can browse below and copy as JSON.")
                .font(.caption)
        }
    }

    @ViewBuilder
    private var summarySection: some View {
        Section {
            if stats.isEmpty {
                Text("No samples yet. Use the app for a few seconds and the table will populate.")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(stats.prefix(20), id: \.self) { stat in
                    StatRow(stat: stat)
                }
            }
        } header: {
            Text("Top 20 by p95")
        } footer: {
            Text("Sorted by 95th-percentile duration. Counts include events; intervals are everything wrapped in `ScarfMon.measure`.")
                .font(.caption)
        }
    }

    @ViewBuilder
    private var actionsSection: some View {
        Section {
            Button {
                copyJSON()
            } label: {
                Label("Copy ring buffer as JSON", systemImage: "doc.on.clipboard")
            }
            Button(role: .destructive) {
                ScarfMonBoot.sharedRingBuffer?.reset()
                refresh()
            } label: {
                Label("Reset ring buffer", systemImage: "trash")
            }
        }
    }

    private func refresh() {
        stats = ScarfMonBoot.sharedRingBuffer?.summary() ?? []
    }

    private func copyJSON() {
        guard let json = ScarfMonBoot.sharedRingBuffer?.exportJSON() else { return }
        UIPasteboard.general.string = json
        copiedToast = true
        Task { @MainActor in
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            copiedToast = false
        }
    }
}

private struct StatRow: View {
    let stat: ScarfMonStat

    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            HStack {
                Text(stat.name)
                    .font(.system(.body, design: .monospaced))
                Spacer()
                Text("p95 \(formatMs(stat.p95Ms))")
                    .font(.caption.monospaced())
                    .foregroundStyle(.secondary)
            }
            HStack(spacing: 12) {
                Text(stat.category.rawValue)
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
                Text("count \(stat.count)")
                    .font(.caption2.monospaced())
                    .foregroundStyle(.tertiary)
                if stat.kind == .interval {
                    Text("p50 \(formatMs(stat.p50Ms))")
                        .font(.caption2.monospaced())
                        .foregroundStyle(.tertiary)
                    Text("max \(formatMs(stat.maxMs))")
                        .font(.caption2.monospaced())
                        .foregroundStyle(.tertiary)
                }
                if stat.totalBytes > 0 {
                    Text("bytes \(stat.totalBytes)")
                        .font(.caption2.monospaced())
                        .foregroundStyle(.tertiary)
                }
            }
        }
    }

    private func formatMs(_ ms: Double) -> String {
        if ms >= 100 { return String(format: "%.0fms", ms) }
        if ms >= 1   { return String(format: "%.1fms", ms) }
        return String(format: "%.2fms", ms)
    }
}
