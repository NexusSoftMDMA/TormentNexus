import SwiftUI
import ScarfCore
import ScarfDesign

/// Confirm sheet for `hermes curator prune` — which **bulk-archives** the
/// agent-created skills idle for ≥ N days. Archiving is reversible (each skill
/// can be restored from the Archived list), so this is a non-destructive
/// tidy-up: a primary confirm, not a red "permanently delete" gate. We still
/// enumerate the affected skills + their idle age so the user sees exactly what
/// moves before confirming. Cancel owns the keyboard default so an accidental
/// Enter-press doesn't archive a batch.
struct CuratorPruneConfirmSheet: View {
    @Environment(\.dismiss) private var dismiss
    let summary: CuratorPruneSummary
    let isPruning: Bool
    let onConfirm: () -> Void
    let onCancel: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
                .padding(.bottom, ScarfSpace.s2)
            ScarfDivider()
            ScrollView {
                VStack(alignment: .leading, spacing: ScarfSpace.s2) {
                    ForEach(summary.candidates) { candidate in
                        row(candidate: candidate)
                    }
                    if summary.candidates.isEmpty {
                        Text("No skills have been idle for at least \(summary.days) days. Nothing to archive.")
                            .scarfStyle(.caption)
                            .foregroundStyle(ScarfColor.foregroundMuted)
                            .padding(.vertical, ScarfSpace.s2)
                    }
                }
                .padding(.vertical, ScarfSpace.s2)
            }
            ScarfDivider()
            footer
                .padding(.top, ScarfSpace.s2)
        }
        .frame(minWidth: 520, minHeight: 380)
        .padding(ScarfSpace.s4)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: ScarfSpace.s1) {
            HStack(alignment: .firstTextBaseline) {
                Text("Archive idle skills")
                    .scarfStyle(.title2)
                    .foregroundStyle(ScarfColor.foregroundPrimary)
                Spacer()
                if summary.count > 0 {
                    ScarfBadge("\(summary.count)", kind: .info)
                }
            }
            Text("These skills haven't been used in at least \(summary.days) days. Archiving moves them out of the active set to keep the curator focused — you can restore any of them later from the Archived list.")
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundMuted)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    private func row(candidate: CuratorPruneCandidate) -> some View {
        HStack(spacing: ScarfSpace.s2) {
            Image(systemName: "archivebox")
                .foregroundStyle(ScarfColor.foregroundFaint)
                .font(.caption)
            Text(candidate.name)
                .scarfStyle(.body)
                .foregroundStyle(ScarfColor.foregroundPrimary)
                .lineLimit(1)
            Spacer()
            Text(candidate.idleLabel)
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundFaint)
                .frame(width: 96, alignment: .trailing)
        }
    }

    private var footer: some View {
        HStack {
            Button("Cancel") {
                onCancel()
                dismiss()
            }
            .buttonStyle(ScarfGhostButton())
            // Cancel owns .defaultAction so an accidental Enter doesn't archive
            // a whole batch, even though the action is reversible.
            .keyboardShortcut(.defaultAction)
            .disabled(isPruning)
            Spacer()
            if isPruning {
                ProgressView().controlSize(.small)
            }
            Button(summary.count == 1 ? "Archive 1 skill" : "Archive \(summary.count) skills") {
                onConfirm()
            }
            .buttonStyle(ScarfPrimaryButton())
            .disabled(isPruning || summary.candidates.isEmpty)
            .accessibilityIdentifier("curatorPrune.confirm")
        }
    }
}
