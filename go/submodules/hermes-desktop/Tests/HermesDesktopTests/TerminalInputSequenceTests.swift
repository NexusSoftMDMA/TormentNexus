import Foundation
import Testing

@testable import HermesDesktop

struct TerminalInputSequenceTests {
    private let bracketedPasteStart: [UInt8] = [0x1B, 0x5B, 0x32, 0x30, 0x30, 0x7E]
    private let bracketedPasteEnd: [UInt8] = [0x1B, 0x5B, 0x32, 0x30, 0x31, 0x7E]

    @Test
    func bracketedPasteSubmissionPreservesBlankLinesAndSubmitsOnce() {
        let prompt = """
        Check this GitHub repository: https://github.com/dodo-reach/hermes-desktop


        inspect and summarize the existing PRs and Issues.
        """

        let payload = TerminalInputSequence.bracketedPasteSubmission(for: prompt)
        let expectedText = String(
            decoding: payload
                .dropFirst(bracketedPasteStart.count)
                .dropLast(bracketedPasteEnd.count + 1),
            as: UTF8.self
        )

        #expect(Array(payload.prefix(bracketedPasteStart.count)) == bracketedPasteStart)
        #expect(Array(payload.suffix(bracketedPasteEnd.count + 1).dropLast()) == bracketedPasteEnd)
        #expect(payload.last == UInt8(ascii: "\r"))
        #expect(expectedText == prompt)
        #expect(payload.filter { $0 == UInt8(ascii: "\r") }.count == 1)
    }

    @Test
    func standardSubmissionAppendsSingleReturn() {
        let payload = TerminalInputSequence.standardSubmission(for: "hello")

        #expect(String(decoding: payload.dropLast(), as: UTF8.self) == "hello")
        #expect(payload.last == UInt8(ascii: "\r"))
        #expect(payload.count == 6)
    }

    @Test
    func bracketedPasteSubmissionPreservesVeryLongPromptWithoutTruncation() {
        let prompt = (0..<4_000)
            .map { "segment-\($0)-abcdefghij" }
            .joined(separator: " ")

        let payload = TerminalInputSequence.bracketedPasteSubmission(for: prompt)
        let extracted = String(
            decoding: payload
                .dropFirst(bracketedPasteStart.count)
                .dropLast(bracketedPasteEnd.count + 1),
            as: UTF8.self
        )

        #expect(extracted == prompt)
        #expect(extracted.count == prompt.count)
        #expect(extracted.hasSuffix("segment-3999-abcdefghij"))
        #expect(payload.last == UInt8(ascii: "\r"))
    }
}
