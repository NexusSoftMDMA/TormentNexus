import Testing
@testable import HermesDesktop

struct HermesTerminalChatInvocationTests {
    @Test
    func terminalResumeInvocationUsesDefaultHermesCommand() {
        let connection = ConnectionProfile(
            label: "Host",
            sshHost: "example.local"
        )
        let invocation = HermesSessionResumeInvocation(
            sessionID: "20260503_161557_453be2",
            connection: connection
        )

        #expect(invocation.arguments == ["--resume", "20260503_161557_453be2"])
        #expect(invocation.commandLine == "hermes --resume 20260503_161557_453be2")
    }

    @Test
    func terminalResumeInvocationPinsCustomHermesProfile() {
        let connection = ConnectionProfile(
            label: "Host",
            sshHost: "example.local",
            hermesProfile: "researcher"
        )
        let invocation = HermesSessionResumeInvocation(
            sessionID: "debug session's final turn",
            connection: connection
        )

        #expect(invocation.arguments == [
            "--profile",
            "researcher",
            "--resume",
            "debug session's final turn"
        ])
        #expect(invocation.commandLine == "hermes --profile researcher --resume 'debug session'\\''s final turn'")
    }

    @Test
    func terminalResumeInvocationSkipsProfileArgumentForCustomHermesHome() {
        let connection = ConnectionProfile(
            label: "Host",
            sshHost: "example.local",
            hermesProfile: "researcher",
            customHermesHomePath: "~/.hermes-work"
        )
        let invocation = HermesSessionResumeInvocation(
            sessionID: "session-123",
            connection: connection
        )

        #expect(invocation.arguments == ["--resume", "session-123"])
        #expect(invocation.commandLine == "hermes --resume session-123")
        #expect(invocation.startupCommandLine.contains(#"$HERMES_HOME/hermes-agent/venv/bin/hermes"#))
        #expect(invocation.startupCommandLine.contains(#""$HERMES_BIN" --resume session-123"#))
    }

    @Test
    func tuiInvocationStartsDefaultProfileNewChat() {
        let connection = ConnectionProfile(label: "Host", sshHost: "example.local")
        let invocation = HermesTUIInvocation(sessionID: nil, connection: connection)

        #expect(invocation.arguments == ["--tui"])
        #expect(invocation.commandLine == "hermes --tui")
        #expect(invocation.startupCommandLine.contains(#""$HERMES_BIN" --tui"#))
    }

    @Test
    func tuiInvocationResumesDefaultProfileSession() {
        let connection = ConnectionProfile(label: "Host", sshHost: "example.local")
        let invocation = HermesTUIInvocation(sessionID: "session-123", connection: connection)

        #expect(invocation.arguments == ["--tui", "--resume", "session-123"])
        #expect(invocation.commandLine == "hermes --tui --resume session-123")
    }

    @Test
    func tuiInvocationStartsNamedProfileNewChat() {
        let connection = ConnectionProfile(
            label: "Host",
            sshHost: "example.local",
            hermesProfile: "work"
        )
        let invocation = HermesTUIInvocation(sessionID: nil, connection: connection)

        #expect(invocation.arguments == ["--profile", "work", "--tui"])
        #expect(invocation.commandLine == "hermes --profile work --tui")
    }

    @Test
    func tuiInvocationResumesNamedProfileSession() {
        let connection = ConnectionProfile(
            label: "Host",
            sshHost: "example.local",
            hermesProfile: "researcher"
        )
        let invocation = HermesTUIInvocation(sessionID: "session-123", connection: connection)

        #expect(invocation.arguments == ["--profile", "researcher", "--tui", "--resume", "session-123"])
        #expect(invocation.commandLine == "hermes --profile researcher --tui --resume session-123")
    }

    @Test
    func tuiInvocationSkipsProfileArgumentForCustomHermesHome() {
        let connection = ConnectionProfile(
            label: "Host",
            sshHost: "example.local",
            hermesProfile: "researcher",
            customHermesHomePath: "~/.hermes-work"
        )
        let invocation = HermesTUIInvocation(sessionID: "session-123", connection: connection)

        #expect(invocation.arguments == ["--tui", "--resume", "session-123"])
        #expect(invocation.commandLine == "hermes --tui --resume session-123")
        #expect(invocation.startupCommandLine.contains(#""$HERMES_BIN" --tui --resume session-123"#))
    }
}
