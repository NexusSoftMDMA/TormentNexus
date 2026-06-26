import SwiftUI

struct MarkdownContentView: View {
    let content: String

    /// Skip the block-AST pipeline mid-stream. `parseBlocks()` walks the
    /// full content on every body re-eval; at Hermes's ~30–60 chunks/sec
    /// that's quadratic over the message lifetime and was a dominant
    /// per-token render cost (cross-reference with the bubble-level
    /// `parseContentBlocks` skip in `RichMessageBubble.contentView`).
    /// Inline bold/italic/code/links still render via
    /// `MarkdownRenderer.inlineAttributedString`; tables, code fences,
    /// headings, lists materialize on finalize when the caller flips
    /// this back to the full pipeline.
    var streaming: Bool = false

    /// Chat font scale plumbed from `RichChatView` (issue #68). Defaults
    /// to 1.0 when this view is used outside the chat surface so other
    /// callers see the un-scaled rendering.
    @Environment(\.chatFontScale) private var chatFontScale: Double

    var body: some View {
        if streaming {
            Text(MarkdownRenderer.inlineAttributedString(content))
                .textSelection(.enabled)
                .font(ChatFontScale.body(chatFontScale))
        } else {
            VStack(alignment: .leading, spacing: 6) {
                // Coalesce consecutive `.paragraph` blocks (with optional
                // `.blank` between them) into a single `Text(AttributedString)`
                // so the cursor can select across paragraphs (issue #93).
                // SwiftUI's `.textSelection(.enabled)` is per-Text — without
                // this pre-pass, every `\n\n` in the agent's reply silently
                // terminates the selection.
                let units = Self.coalesceParagraphs(parseBlocks())
                ForEach(Array(units.enumerated()), id: \.offset) { _, unit in
                    unitView(unit)
                }
            }
            // Paragraphs are rendered as plain `Text(AttributedString)` and
            // inherit whatever font is set on the enclosing scope. Pin the
            // scope to the scaled body font so the chat slider actually
            // moves the visible text.
            .font(ChatFontScale.body(chatFontScale))
        }
    }

    @ViewBuilder
    private func unitView(_ unit: RenderableUnit) -> some View {
        switch unit {
        case .block(let block):
            blockView(block)
        case .paragraphGroup(let texts):
            paragraphGroupView(texts: texts)
        }
    }

    /// Render a run of consecutive paragraphs as ONE Text, joining the
    /// per-paragraph AttributedStrings with `\n\n`. The single-Text
    /// shape is what makes selection span paragraph breaks — that's
    /// issue #93's whole point.
    private func paragraphGroupView(texts: [String]) -> some View {
        var combined = AttributedString()
        for (idx, text) in texts.enumerated() {
            if idx > 0 {
                combined.append(AttributedString("\n\n"))
            }
            combined.append(MarkdownRenderer.inlineAttributedString(text))
        }
        return Text(combined)
            .textSelection(.enabled)
    }

    @ViewBuilder
    private func blockView(_ block: MarkdownBlock) -> some View {
        switch block {
        case .heading(let level, let text):
            headingView(level: level, text: text)
        case .paragraph(let text):
            // Reached only for a paragraph that wasn't coalesced (e.g.
            // a single paragraph adjacent to non-paragraph blocks). Same
            // selection semantics — the coalesced path is just a wider
            // selection scope.
            Text(MarkdownRenderer.inlineAttributedString(text))
                .textSelection(.enabled)
        case .codeBlock(let code, let language):
            codeBlockView(code: code, language: language)
        case .bulletItem(let text, let indent):
            bulletView(text: text, indent: indent)
        case .numberedItem(let number, let text):
            numberedView(number: number, text: text)
        case .blockquote(let text):
            blockquoteView(text: text)
        case .horizontalRule:
            Divider().padding(.vertical, 4)
        case .blank:
            Spacer().frame(height: 4)
        }
    }

    // MARK: - Block Views

    private func headingView(level: Int, text: String) -> some View {
        // Heading sizes scale with `chatFontScale` (issue #68). Bases
        // mirror the SwiftUI semantic tokens we used previously
        // (`.title` ≈ 28, `.title2` ≈ 22, `.title3` ≈ 20, `.headline`
        // ≈ 17, `.subheadline` ≈ 15) so 100% matches today's UI.
        let baseSize: CGFloat = switch level {
        case 1: 28
        case 2: 22
        case 3: 20
        case 4: 17
        default: 15
        }
        return Text(MarkdownRenderer.inlineAttributedString(text))
            .font(.system(size: baseSize * chatFontScale, weight: .semibold))
            .textSelection(.enabled)
            .padding(.top, level <= 2 ? 8 : 4)
    }

    private func codeBlockView(code: String, language: String?) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            if let lang = language, !lang.isEmpty {
                Text(lang)
                    .font(ChatFontScale.caption2(chatFontScale).bold())
                    .foregroundStyle(.secondary)
            }
            Text(code)
                .font(ChatFontScale.codeInline(chatFontScale))
                .textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(10)
        .background(Color(.textBackgroundColor).opacity(0.5))
        .clipShape(RoundedRectangle(cornerRadius: 6))
        .overlay(
            RoundedRectangle(cornerRadius: 6)
                .strokeBorder(.quaternary, lineWidth: 1)
        )
    }

    private func bulletView(text: String, indent: Int) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: 6) {
            Text("\u{2022}")
                .foregroundStyle(.secondary)
            Text(MarkdownRenderer.inlineAttributedString(text))
                .textSelection(.enabled)
        }
        .padding(.leading, CGFloat(indent) * 16)
    }

    private func numberedView(number: Int, text: String) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: 6) {
            Text("\(number).")
                .foregroundStyle(.secondary)
                .frame(width: 20, alignment: .trailing)
            Text(MarkdownRenderer.inlineAttributedString(text))
                .textSelection(.enabled)
        }
    }

    private func blockquoteView(text: String) -> some View {
        HStack(spacing: 0) {
            RoundedRectangle(cornerRadius: 1)
                .fill(.blue.opacity(0.5))
                .frame(width: 3)
            Text(MarkdownRenderer.inlineAttributedString(text))
                .foregroundStyle(.secondary)
                .textSelection(.enabled)
                .padding(.leading, 10)
        }
        .padding(.vertical, 2)
    }

    // MARK: - Parser

    private func parseBlocks() -> [MarkdownBlock] {
        var blocks: [MarkdownBlock] = []
        let lines = content.components(separatedBy: "\n")
        var i = 0

        // Skip YAML frontmatter (--- delimited block at start of file)
        if i < lines.count && lines[i].trimmingCharacters(in: .whitespaces) == "---" {
            i += 1
            while i < lines.count {
                if lines[i].trimmingCharacters(in: .whitespaces) == "---" {
                    i += 1
                    break
                }
                i += 1
            }
        }

        while i < lines.count {
            let line = lines[i]
            let trimmed = line.trimmingCharacters(in: .whitespaces)

            // Blank line
            if trimmed.isEmpty {
                if blocks.last != .blank {
                    blocks.append(.blank)
                }
                i += 1
                continue
            }

            // Code block (fenced)
            if trimmed.hasPrefix("```") {
                let language = String(trimmed.dropFirst(3)).trimmingCharacters(in: .whitespaces)
                var codeLines: [String] = []
                i += 1
                while i < lines.count {
                    if lines[i].trimmingCharacters(in: .whitespaces).hasPrefix("```") {
                        i += 1
                        break
                    }
                    codeLines.append(lines[i])
                    i += 1
                }
                blocks.append(.codeBlock(codeLines.joined(separator: "\n"), language: language.isEmpty ? nil : language))
                continue
            }

            // Heading
            if let heading = parseHeading(trimmed) {
                blocks.append(heading)
                i += 1
                continue
            }

            // Horizontal rule
            if isHorizontalRule(trimmed) {
                blocks.append(.horizontalRule)
                i += 1
                continue
            }

            // Blockquote
            if trimmed.hasPrefix("> ") {
                var quoteLines: [String] = []
                while i < lines.count {
                    let l = lines[i].trimmingCharacters(in: .whitespaces)
                    if l.hasPrefix("> ") {
                        quoteLines.append(String(l.dropFirst(2)))
                    } else if l.hasPrefix(">") {
                        quoteLines.append(String(l.dropFirst(1)))
                    } else {
                        break
                    }
                    i += 1
                }
                blocks.append(.blockquote(quoteLines.joined(separator: " ")))
                continue
            }

            // Bullet list
            if let bullet = parseBullet(line) {
                blocks.append(bullet)
                i += 1
                continue
            }

            // Numbered list
            if let numbered = parseNumbered(trimmed) {
                blocks.append(numbered)
                i += 1
                continue
            }

            // Paragraph — each line is its own paragraph to preserve line breaks
            blocks.append(.paragraph(trimmed))
            i += 1
        }

        return blocks
    }

    private func parseHeading(_ line: String) -> MarkdownBlock? {
        let levels: [(prefix: String, level: Int)] = [
            ("##### ", 5), ("#### ", 4), ("### ", 3), ("## ", 2), ("# ", 1)
        ]
        for (prefix, level) in levels {
            if line.hasPrefix(prefix) {
                return .heading(level, String(line.dropFirst(prefix.count)))
            }
        }
        return nil
    }

    private func parseBullet(_ line: String) -> MarkdownBlock? {
        let indent = line.prefix(while: { $0 == " " }).count / 2
        let trimmed = line.trimmingCharacters(in: .whitespaces)
        if trimmed.hasPrefix("- ") {
            return .bulletItem(String(trimmed.dropFirst(2)), indent: indent)
        }
        if trimmed.hasPrefix("* ") {
            return .bulletItem(String(trimmed.dropFirst(2)), indent: indent)
        }
        return nil
    }

    private func parseNumbered(_ line: String) -> MarkdownBlock? {
        guard let dotIdx = line.firstIndex(of: ".") else { return nil }
        let numStr = String(line[line.startIndex..<dotIdx])
        guard let num = Int(numStr), line[line.index(after: dotIdx)...].hasPrefix(" ") else { return nil }
        let text = String(line[line.index(dotIdx, offsetBy: 2)...])
        return .numberedItem(num, text)
    }

    private func isHorizontalRule(_ line: String) -> Bool {
        let stripped = line.replacingOccurrences(of: " ", with: "")
        return (stripped.allSatisfy({ $0 == "-" }) && stripped.count >= 3) ||
               (stripped.allSatisfy({ $0 == "*" }) && stripped.count >= 3) ||
               (stripped.allSatisfy({ $0 == "_" }) && stripped.count >= 3)
    }

    /// Walk the parsed block list and collapse runs of `.paragraph`
    /// (with optional `.blank` separators between them) into a single
    /// `RenderableUnit.paragraphGroup`. Non-paragraph blocks stay as
    /// `.block(...)` units.
    ///
    /// Why: each block previously rendered to its own SwiftUI `Text`,
    /// and SwiftUI's `.textSelection(.enabled)` is per-Text. The user
    /// could drag-select within one paragraph but not across the
    /// `\n\n` gap to the next — issue #93. Joining a run into one
    /// AttributedString preserves the visual gap (real `\n\n` in the
    /// rendered text) while giving the cursor one continuous selection
    /// scope.
    ///
    /// Visible behavior change is scoped to consecutive paragraphs.
    /// Selection across a heading/list/code-block boundary still
    /// terminates — fixing that requires NSTextView and a much larger
    /// refactor; the common case (multi-paragraph agent reply) is what
    /// the user actually reported.
    ///
    /// `internal` so the scarfTests target can pin the coalescing
    /// invariants without touching the SwiftUI render path.
    static func coalesceParagraphs(_ blocks: [MarkdownBlock]) -> [RenderableUnit] {
        var units: [RenderableUnit] = []
        var currentRun: [String] = []
        var pendingBlank = false

        func flushRun() {
            if !currentRun.isEmpty {
                units.append(.paragraphGroup(currentRun))
                currentRun.removeAll()
            }
            if pendingBlank {
                units.append(.block(.blank))
                pendingBlank = false
            }
        }

        for block in blocks {
            switch block {
            case .paragraph(let text):
                // Trailing blank from earlier is absorbed by the
                // `\n\n` join in the rendered paragraphGroup.
                pendingBlank = false
                currentRun.append(text)
            case .blank:
                if currentRun.isEmpty {
                    // Not in a paragraph run — emit the vertical gap
                    // verbatim (matches pre-fix rendering between
                    // non-paragraph blocks).
                    units.append(.block(.blank))
                } else {
                    // In a run. Defer: if the next block is another
                    // paragraph the run continues; if it's a
                    // structural block we'll flush + render the
                    // blank as a real visual gap.
                    pendingBlank = true
                }
            default:
                flushRun()
                units.append(.block(block))
            }
        }
        flushRun()
        return units
    }
}

// MARK: - Block Model

/// Parsed markdown block. `internal` so scarfTests can pin the
/// `coalesceParagraphs` invariants without re-implementing the parser
/// from scratch.
enum MarkdownBlock: Equatable {
    case heading(Int, String)
    case paragraph(String)
    case codeBlock(String, language: String?)
    case bulletItem(String, indent: Int)
    case numberedItem(Int, String)
    case blockquote(String)
    case horizontalRule
    case blank
}

/// Output of `MarkdownContentView.coalesceParagraphs(_:)`. Either a
/// single non-paragraph block we render verbatim, or a coalesced run
/// of paragraphs that becomes one selectable `Text(AttributedString)`.
enum RenderableUnit: Equatable {
    case block(MarkdownBlock)
    case paragraphGroup([String])
}
