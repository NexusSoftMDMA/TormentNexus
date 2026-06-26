import SwiftUI
import ScarfCore
import ScarfDesign

/// Mac sub-view rendered between the active-skill leaderboards and the
/// last-report block on Hermes v0.13+ hosts. Lists everything currently
/// archived (`hermes curator list-archived`) with per-row Restore.
/// (Bulk "Archive idle skills" lives in the header menu, not here — it
/// archives idle *active* skills, it doesn't act on this archived list.)
///
/// Empty-state copy explains what archive means — useful when the
/// curator hasn't run yet on a fresh install (no archives ≠ a problem).
struct CuratorArchivedSection: View {
    let archived: [HermesCuratorArchivedSkill]
    let isLoading: Bool
    let onRestore: (String) -> Void

    var body: some View {
        ScarfCard {
            VStack(alignment: .leading, spacing: ScarfSpace.s2) {
                header
                if isLoading && archived.isEmpty {
                    loadingRow
                } else if archived.isEmpty {
                    emptyState
                } else {
                    rows
                }
            }
        }
    }

    private var header: some View {
        HStack(alignment: .firstTextBaseline) {
            ScarfSectionHeader("Archived")
            Spacer()
            Text("\(archived.count) skill\(archived.count == 1 ? "" : "s")")
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundMuted)
        }
    }

    private var loadingRow: some View {
        HStack(spacing: ScarfSpace.s2) {
            ProgressView().controlSize(.small)
            Text("Loading archived skills…")
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundMuted)
            Spacer()
        }
    }

    private var emptyState: some View {
        VStack(alignment: .leading, spacing: ScarfSpace.s1) {
            Text("No archived skills.")
                .scarfStyle(.body)
                .foregroundStyle(ScarfColor.foregroundMuted)
            Text("The curator moves stale or redundant skills here on its weekly review. Until then, this list stays empty.")
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundFaint)
        }
    }

    private var rows: some View {
        VStack(alignment: .leading, spacing: ScarfSpace.s1) {
            ForEach(archived) { skill in
                ArchivedSkillRow(
                    skill: skill,
                    onRestore: { onRestore(skill.name) }
                )
            }
        }
    }
}

private struct ArchivedSkillRow: View {
    let skill: HermesCuratorArchivedSkill
    let onRestore: () -> Void

    var body: some View {
        HStack(alignment: .center, spacing: ScarfSpace.s2) {
            Image(systemName: "archivebox.fill")
                .font(.system(size: 12))
                .foregroundStyle(ScarfColor.foregroundFaint)
            VStack(alignment: .leading, spacing: 2) {
                Text(skill.name)
                    .scarfStyle(.body)
                    .foregroundStyle(ScarfColor.foregroundPrimary)
                    .lineLimit(1)
                if let reason = skill.reason, !reason.isEmpty {
                    Text(reason)
                        .scarfStyle(.caption)
                        .foregroundStyle(ScarfColor.foregroundMuted)
                        .lineLimit(1)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            Text(skill.archivedAtLabel)
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundFaint)
                .frame(width: 96, alignment: .trailing)
            Text(skill.sizeLabel)
                .scarfStyle(.caption)
                .foregroundStyle(ScarfColor.foregroundFaint)
                .frame(width: 72, alignment: .trailing)
            Button("Restore") {
                onRestore()
            }
            .buttonStyle(ScarfPrimaryButton())
            .controlSize(.small)
            .help("Restore \(skill.name) to the active skill set")
        }
        .padding(.vertical, 2)
    }
}
