import CryptoKit
import Foundation
import OSLog

struct WorkflowLaunchDiagnosticsContext: Sendable {
    let runID: UUID
    let workflowID: UUID
    let workflowName: String
    let connectionLabel: String
    let hermesProfileName: String?
    let skillRelativePaths: [String]
    let commandLine: String
    let originalPromptCharacterCount: Int
    let originalPromptUTF8ByteCount: Int
    let originalPromptLineCount: Int
    let originalPromptHashPrefix: String
    let normalizedPromptCharacterCount: Int
    let normalizedPromptUTF8ByteCount: Int
    let normalizedPromptLineCount: Int
    let normalizedPromptHashPrefix: String
    let requestedAt: Date

    init(
        workflow: WorkflowPreset,
        invocation: WorkflowLaunchInvocation,
        connection: ConnectionProfile,
        requestedAt: Date = Date()
    ) {
        let originalPrompt = workflow.prompt
        let normalizedPrompt = invocation.initialInput

        self.runID = UUID()
        self.workflowID = workflow.id
        self.workflowName = workflow.name
        self.connectionLabel = connection.label
        self.hermesProfileName = connection.trimmedHermesProfile
        self.skillRelativePaths = invocation.skillRelativePaths
        self.commandLine = invocation.commandLine
        self.originalPromptCharacterCount = originalPrompt.count
        self.originalPromptUTF8ByteCount = originalPrompt.lengthOfBytes(using: .utf8)
        self.originalPromptLineCount = Self.lineCount(for: originalPrompt)
        self.originalPromptHashPrefix = Self.hashPrefix(for: originalPrompt)
        self.normalizedPromptCharacterCount = normalizedPrompt.count
        self.normalizedPromptUTF8ByteCount = normalizedPrompt.lengthOfBytes(using: .utf8)
        self.normalizedPromptLineCount = Self.lineCount(for: normalizedPrompt)
        self.normalizedPromptHashPrefix = Self.hashPrefix(for: normalizedPrompt)
        self.requestedAt = requestedAt
    }

    func elapsedMilliseconds(at date: Date = Date()) -> Int {
        max(0, Int(date.timeIntervalSince(requestedAt) * 1000))
    }

    private static func lineCount(for value: String) -> Int {
        guard !value.isEmpty else { return 0 }
        return value.components(separatedBy: .newlines).count
    }

    private static func hashPrefix(for value: String) -> String {
        let digest = SHA256.hash(data: Data(value.utf8))
        return digest.prefix(6).map { String(format: "%02x", $0) }.joined()
    }
}

enum WorkflowInitialInputDeliveryMode: String, Sendable {
    case bracketedPaste = "bracketed_paste"
    case standardSubmit = "standard_submit"
}

actor WorkflowLaunchDiagnostics {
    nonisolated let logFileURL: URL

    private let fileManager: FileManager
    private let logger: Logger
    private let dateFormatter: ISO8601DateFormatter

    init(logFileURL: URL, fileManager: FileManager = .default) {
        let diagnosticsDirectoryURL = logFileURL.deletingLastPathComponent()
        let subsystem = Bundle.main.bundleIdentifier ?? "HermesDesktop"

        self.logFileURL = logFileURL
        self.fileManager = fileManager
        self.logger = Logger(subsystem: subsystem, category: "WorkflowLaunch")
        self.dateFormatter = ISO8601DateFormatter()
        self.dateFormatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]

        try? fileManager.createDirectory(at: diagnosticsDirectoryURL, withIntermediateDirectories: true)
        try? Data().write(to: logFileURL, options: .atomic)

        let sessionStartedLine = Self.makeLine(
            dateFormatter: dateFormatter,
            event: "diagnostics_session_started",
            fields: [
                "banner": "=== Hermes Desktop workflow launch diagnostics session started ===",
                "log_path": logFileURL.path
            ]
        )
        Self.appendLineSynchronously(
            sessionStartedLine,
            logger: logger,
            fileManager: fileManager,
            logFileURL: logFileURL
        )
    }

    func recordWorkflowRunRequested(_ context: WorkflowLaunchDiagnosticsContext) {
        record(
            event: "workflow_run_requested",
            context: context,
            fields: [
                "command_line": context.commandLine,
                "connection": context.connectionLabel,
                "hermes_profile": context.hermesProfileName ?? "default",
                "normalized_prompt_chars": "\(context.normalizedPromptCharacterCount)",
                "normalized_prompt_hash": context.normalizedPromptHashPrefix,
                "normalized_prompt_lines": "\(context.normalizedPromptLineCount)",
                "normalized_prompt_utf8_bytes": "\(context.normalizedPromptUTF8ByteCount)",
                "original_prompt_chars": "\(context.originalPromptCharacterCount)",
                "original_prompt_hash": context.originalPromptHashPrefix,
                "original_prompt_lines": "\(context.originalPromptLineCount)",
                "original_prompt_utf8_bytes": "\(context.originalPromptUTF8ByteCount)",
                "skill_count": "\(context.skillRelativePaths.count)",
                "skills": context.skillRelativePaths.joined(separator: ","),
                "workflow_id": context.workflowID.uuidString.lowercased(),
                "workflow_name": context.workflowName
            ]
        )
    }

    func recordTerminalProcessStarted(_ context: WorkflowLaunchDiagnosticsContext) {
        record(
            event: "terminal_process_started",
            context: context,
            fields: [:]
        )
    }

    func recordInitialInputWaitStarted(_ context: WorkflowLaunchDiagnosticsContext, deadlineMilliseconds: Int) {
        record(
            event: "initial_input_wait_started",
            context: context,
            fields: [
                "deadline_ms": "\(deadlineMilliseconds)"
            ]
        )
    }

    func recordBracketedPasteModeObserved(
        _ context: WorkflowLaunchDiagnosticsContext,
        stage: String
    ) {
        record(
            event: "bracketed_paste_mode_observed",
            context: context,
            fields: [
                "stage": stage
            ]
        )
    }

    func recordInitialInputSent(
        _ context: WorkflowLaunchDiagnosticsContext,
        deliveryMode: WorkflowInitialInputDeliveryMode,
        reason: String,
        bracketedPasteModeAtSend: Bool
    ) {
        record(
            event: "initial_input_sent",
            context: context,
            fields: [
                "bracketed_paste_mode_at_send": bracketedPasteModeAtSend ? "true" : "false",
                "delivery_mode": deliveryMode.rawValue,
                "reason": reason
            ]
        )
    }

    func recordInitialInputAborted(
        _ context: WorkflowLaunchDiagnosticsContext,
        reason: String
    ) {
        record(
            event: "initial_input_aborted",
            context: context,
            fields: [
                "reason": reason
            ]
        )
    }

    func recordTerminalProcessExited(
        _ context: WorkflowLaunchDiagnosticsContext,
        exitCode: Int32?
    ) {
        record(
            event: "terminal_process_exited",
            context: context,
            fields: [
                "exit_code": exitCode.map(String.init) ?? "nil"
            ]
        )
    }

    private func record(
        event: String,
        context: WorkflowLaunchDiagnosticsContext,
        fields: [String: String]
    ) {
        var values = fields
        values["elapsed_ms"] = "\(context.elapsedMilliseconds())"
        Self.appendLineSynchronously(Self.makeLine(
            dateFormatter: dateFormatter,
            event: event,
            runID: context.runID,
            fields: values
        ), logger: logger, fileManager: fileManager, logFileURL: logFileURL)
    }

    private static func makeLine(
        dateFormatter: ISO8601DateFormatter,
        event: String,
        runID: UUID? = nil,
        fields: [String: String]
    ) -> String {
        var segments = [
            "ts=\(dateFormatter.string(from: Date()))",
            "event=\(event)"
        ]

        if let runID {
            segments.append("run=\(runID.uuidString.lowercased())")
        }

        for key in fields.keys.sorted() {
            guard let value = fields[key] else { continue }
            segments.append("\(key)=\(Self.sanitize(value))")
        }

        return segments.joined(separator: " | ")
    }

    private static func sanitize(_ value: String) -> String {
        let cleaned = value
            .replacingOccurrences(of: "\n", with: "\\n")
            .replacingOccurrences(of: "\r", with: "\\r")
            .replacingOccurrences(of: "|", with: "/")
            .replacingOccurrences(of: "\"", with: "'")
        return "\"\(cleaned)\""
    }

    private static func appendLineSynchronously(
        _ line: String,
        logger: Logger,
        fileManager: FileManager,
        logFileURL: URL
    ) {
        logger.notice("\(line, privacy: .public)")

        let payload = Data((line + "\n").utf8)
        if !fileManager.fileExists(atPath: logFileURL.path) {
            fileManager.createFile(atPath: logFileURL.path, contents: Data(), attributes: nil)
        }

        guard let handle = try? FileHandle(forWritingTo: logFileURL) else { return }
        defer {
            try? handle.close()
        }

        do {
            try handle.seekToEnd()
            try handle.write(contentsOf: payload)
        } catch {
            return
        }
    }
}
