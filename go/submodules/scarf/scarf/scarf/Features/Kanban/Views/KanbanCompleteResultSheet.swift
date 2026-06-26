import SwiftUI
import ScarfCore
import ScarfDesign

/// Modal sheet that prompts for an optional "result summary" before
/// firing `kanban complete`. Optional — leaving it blank still
/// completes the task; the field captures the most useful Hermes
/// flag for downstream child tasks.
struct KanbanCompleteResultSheet: View {
    @Environment(\.dismiss) private var dismiss

    let taskTitle: String
    let onSubmit: (String?) -> Void

    @State private var result: String = ""
    @FocusState private var fieldFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: ScarfSpace.s3) {
            VStack(alignment: .leading, spacing: 4) {
                Text("Complete task")
                    .scarfStyle(.title3)
                    .foregroundStyle(ScarfColor.foregroundPrimary)
                Text(taskTitle)
                    .scarfStyle(.caption)
                    .foregroundStyle(ScarfColor.foregroundMuted)
                    .lineLimit(2)
            }

            ScarfTextField("Result summary (optional)", text: $result)
                .focused($fieldFocused)

            Text("If this task has child tasks, the result is handed to them as upstream context. Leave blank for a quiet completion.")
                .scarfStyle(.footnote)
                .foregroundStyle(ScarfColor.foregroundFaint)

            HStack {
                Spacer()
                Button("Cancel") {
                    dismiss()
                }
                .keyboardShortcut(.cancelAction)
                .buttonStyle(ScarfSecondaryButton())
                Button("Complete") {
                    onSubmit(result.trimmingCharacters(in: .whitespacesAndNewlines))
                    dismiss()
                }
                .keyboardShortcut(.defaultAction)
                .buttonStyle(ScarfPrimaryButton())
            }
        }
        .padding(ScarfSpace.s5)
        .frame(width: 460)
        .onAppear { fieldFocused = true }
    }
}
