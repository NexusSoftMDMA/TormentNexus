import AppKit
import SwiftUI

private struct BackgroundImageActiveKey: EnvironmentKey {
    static let defaultValue = false
}

extension EnvironmentValues {
    var backgroundImageActive: Bool {
        get { self[BackgroundImageActiveKey.self] }
        set { self[BackgroundImageActiveKey.self] = newValue }
    }
}

enum HermesTheme {
    static let pageHorizontalPadding: CGFloat = 24
    static let pageVerticalPadding: CGFloat = 22
    static let panelCornerRadius: CGFloat = 14
    static let insetCornerRadius: CGFloat = 10
    static let rowCornerRadius: CGFloat = 12

    /// Window-level background. The sidebar List with `.sidebar` style also
    /// uses this token so the surface under the native chrome matches the app.
    static var appBackground: Color {
        Color(nsColor: .windowBackgroundColor)
    }

    /// Same as `appBackground` on macOS — kept as a dedicated token so views
    /// can opt into the sidebar surface without reading `windowBackgroundColor`
    /// directly.
    static var sidebarBackground: Color {
        Color(nsColor: .windowBackgroundColor)
    }

    /// Card / panel surface. Tied directly to `controlBackgroundColor` so the
    /// panel feels like a native macOS control, not a translucent overlay.
    static var panelFill: Color {
        Color(nsColor: .controlBackgroundColor)
    }

    /// Inset / prompt / editor surface. Slightly lighter than the panel using
    /// `textBackgroundColor`, which reads as a macOS text field on both
    /// Light and Dark.
    static var insetFill: Color {
        Color(nsColor: .textBackgroundColor).opacity(0.55)
    }

    /// Row inside a panel. Same family as `panelFill` but a touch lighter so
    /// rows separate from the panel background without needing heavy borders.
    static var rowFill: Color {
        Color(nsColor: .controlBackgroundColor).opacity(0.72)
    }

    /// Row hover highlight. Slightly more opaque than `rowFill` so a hovered
    /// row reads as "lifted" without a separate accent.
    static var hoverFill: Color {
        Color(nsColor: .controlBackgroundColor).opacity(0.88)
    }

    /// Subtle hairline used by every neutral card/row border. Anchored to
    /// `separatorColor` so it adapts to Light/Dark without hand-tuned opacities.
    static var subtleStroke: Color {
        Color(nsColor: .separatorColor).opacity(0.55)
    }

    /// Selection fill, mid of the recommended 0.12–0.14 range.
    static var selectedFill: Color {
        Color.accentColor.opacity(0.14)
    }

    /// Selection border, mid of the recommended 0.25–0.28 range.
    static var selectedStroke: Color {
        Color.accentColor.opacity(0.28)
    }

    /// Warning surface fill (missing skills, validation errors, etc).
    static var warningFill: Color {
        Color.orange.opacity(0.11)
    }

    /// Warning surface border. Pair with `warningFill` to keep dark mode from
    /// collapsing into a muddy brown.
    static var warningStroke: Color {
        Color.orange.opacity(0.28)
    }

    /// Warning foreground (titles, icons inside a warning surface).
    static var warningForeground: Color {
        Color.orange
    }
}

enum HermesPageWidth {
    case standard
    case dashboard
    case analytics

    var maxWidth: CGFloat {
        switch self {
        case .standard:
            return 1360
        case .dashboard:
            return 1480
        case .analytics:
            return 1560
        }
    }
}

enum HermesSplitMetrics {
    enum WorkspaceSidebar {
        static let minWidth: CGFloat = 120
        static let defaultWidth: CGFloat = 168
        static let maxWidth: CGFloat = 210
        static let detailFallbackMinWidth: CGFloat = 320
    }

    enum WorkbenchBrowser {
        static let minWidth: CGFloat = 260
        static let defaultWidth: CGFloat = 380
        static let maxWidth: CGFloat = 760
    }

    enum WorkbenchDetail {
        static let minWidth: CGFloat = 460
        static let sessionsIdealWidth: CGFloat = 520
        static let standardIdealWidth: CGFloat = 560
        static let formIdealWidth: CGFloat = 600
        static let editorIdealWidth: CGFloat = 640
    }

    static let minimumWindowWidth =
        WorkspaceSidebar.minWidth +
        WorkbenchBrowser.minWidth +
        WorkbenchDetail.minWidth +
        (HermesSplitLayout.resizeHandleWidth * 2)

    static var standardWorkbenchBrowserLayout: HermesSplitLayout {
        HermesSplitLayout(
            minPrimaryWidth: WorkbenchBrowser.minWidth,
            defaultPrimaryWidth: WorkbenchBrowser.defaultWidth,
            maxPrimaryWidth: WorkbenchBrowser.maxWidth
        )
    }
}

struct HermesPageContainer<Content: View>: View {
    let width: HermesPageWidth
    let horizontalPadding: CGFloat
    let verticalPadding: CGFloat
    let content: Content

    init(
        width: HermesPageWidth = .standard,
        horizontalPadding: CGFloat = HermesTheme.pageHorizontalPadding,
        verticalPadding: CGFloat = HermesTheme.pageVerticalPadding,
        @ViewBuilder content: () -> Content
    ) {
        self.width = width
        self.horizontalPadding = horizontalPadding
        self.verticalPadding = verticalPadding
        self.content = content()
    }

    var body: some View {
        ScrollView {
            content
                .frame(maxWidth: width.maxWidth, alignment: .leading)
                .padding(.horizontal, horizontalPadding)
                .padding(.vertical, verticalPadding)
                .frame(maxWidth: .infinity, alignment: .top)
        }
    }
}

struct HermesPageHeader<Accessory: View>: View {
    let title: String
    let subtitle: String
    let accessory: Accessory

    init(
        title: String,
        subtitle: String,
        @ViewBuilder accessory: () -> Accessory
    ) {
        self.title = title
        self.subtitle = subtitle
        self.accessory = accessory()
    }

    init(title: String, subtitle: String) where Accessory == EmptyView {
        self.title = title
        self.subtitle = subtitle
        self.accessory = EmptyView()
    }

    var body: some View {
        HermesAdaptivePairLayout(
            horizontalSpacing: 20,
            verticalSpacing: 12,
            minimumPrimaryWidth: 260
        ) {
            titleBlock
            accessory
        }
        .frame(maxWidth: .infinity, alignment: .topLeading)
    }

    private var titleBlock: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(L10n.string(title))
                .font(.title)
                .fontWeight(.semibold)
                .lineLimit(1)
                .minimumScaleFactor(0.82)

            Text(L10n.string(subtitle))
                .font(.callout)
                .foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)
        }
    }
}

private struct HermesAdaptivePairLayout: Layout {
    let horizontalSpacing: CGFloat
    let verticalSpacing: CGFloat
    var minimumPrimaryWidth: CGFloat?

    func sizeThatFits(
        proposal: ProposedViewSize,
        subviews: Subviews,
        cache: inout Void
    ) -> CGSize {
        guard subviews.count >= 2 else {
            return subviews.first?.sizeThatFits(proposal) ?? .zero
        }

        let availableWidth = proposal.width ?? .greatestFiniteMagnitude
        let primaryIdealSize = subviews[0].sizeThatFits(.unspecified)
        let secondarySize = subviews[1].sizeThatFits(.unspecified)

        if usesHorizontalLayout(
            availableWidth: availableWidth,
            primaryIdealWidth: primaryIdealSize.width,
            secondaryWidth: secondarySize.width
        ) {
            let primaryWidth = horizontalPrimaryWidth(
                availableWidth: availableWidth,
                primaryIdealWidth: primaryIdealSize.width,
                secondaryWidth: secondarySize.width
            )
            let primarySize = subviews[0].sizeThatFits(ProposedViewSize(width: primaryWidth, height: nil))
            let width = proposal.width ?? primarySize.width + horizontalSpacing + secondarySize.width

            return CGSize(
                width: width,
                height: max(primarySize.height, secondarySize.height)
            )
        }

        let primarySize = subviews[0].sizeThatFits(ProposedViewSize(width: proposal.width, height: nil))
        let secondaryConstrainedSize = subviews[1].sizeThatFits(ProposedViewSize(width: proposal.width, height: nil))
        let width = proposal.width ?? max(primarySize.width, secondaryConstrainedSize.width)

        return CGSize(
            width: width,
            height: primarySize.height + verticalSpacing + secondaryConstrainedSize.height
        )
    }

    func placeSubviews(
        in bounds: CGRect,
        proposal: ProposedViewSize,
        subviews: Subviews,
        cache: inout Void
    ) {
        guard subviews.count >= 2 else {
            subviews.first?.place(
                at: bounds.origin,
                proposal: ProposedViewSize(width: bounds.width, height: bounds.height)
            )
            return
        }

        let primaryIdealSize = subviews[0].sizeThatFits(.unspecified)
        let secondarySize = subviews[1].sizeThatFits(.unspecified)

        if usesHorizontalLayout(
            availableWidth: bounds.width,
            primaryIdealWidth: primaryIdealSize.width,
            secondaryWidth: secondarySize.width
        ) {
            let primaryWidth = horizontalPrimaryWidth(
                availableWidth: bounds.width,
                primaryIdealWidth: primaryIdealSize.width,
                secondaryWidth: secondarySize.width
            )
            let primarySize = subviews[0].sizeThatFits(ProposedViewSize(width: primaryWidth, height: nil))

            subviews[0].place(
                at: CGPoint(x: bounds.minX, y: bounds.minY),
                proposal: ProposedViewSize(width: primaryWidth, height: primarySize.height)
            )
            subviews[1].place(
                at: CGPoint(x: bounds.maxX - secondarySize.width, y: bounds.minY),
                proposal: ProposedViewSize(width: secondarySize.width, height: secondarySize.height)
            )
        } else {
            let primarySize = subviews[0].sizeThatFits(ProposedViewSize(width: bounds.width, height: nil))
            let secondarySize = subviews[1].sizeThatFits(ProposedViewSize(width: bounds.width, height: nil))

            subviews[0].place(
                at: CGPoint(x: bounds.minX, y: bounds.minY),
                proposal: ProposedViewSize(width: bounds.width, height: primarySize.height)
            )
            subviews[1].place(
                at: CGPoint(x: bounds.minX, y: bounds.minY + primarySize.height + verticalSpacing),
                proposal: ProposedViewSize(width: bounds.width, height: secondarySize.height)
            )
        }
    }

    private func usesHorizontalLayout(
        availableWidth: CGFloat,
        primaryIdealWidth: CGFloat,
        secondaryWidth: CGFloat
    ) -> Bool {
        guard availableWidth.isFinite else { return true }

        let requiredPrimaryWidth = minimumPrimaryWidth ?? primaryIdealWidth
        return availableWidth >= requiredPrimaryWidth + horizontalSpacing + secondaryWidth
    }

    private func horizontalPrimaryWidth(
        availableWidth: CGFloat,
        primaryIdealWidth: CGFloat,
        secondaryWidth: CGFloat
    ) -> CGFloat {
        guard availableWidth.isFinite else { return primaryIdealWidth }
        return max(0, availableWidth - horizontalSpacing - secondaryWidth)
    }
}

struct HermesSurfacePanel<Content: View>: View {
    let title: String?
    let subtitle: String?
    let content: Content
    @Environment(\.backgroundImageActive) private var backgroundImageActive

    init(
        title: String? = nil,
        subtitle: String? = nil,
        @ViewBuilder content: () -> Content
    ) {
        self.title = title
        self.subtitle = subtitle
        self.content = content()
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            if title != nil || subtitle != nil {
                VStack(alignment: .leading, spacing: 6) {
                    if let title {
                        Text(L10n.string(title))
                            .font(.headline)
                    }

                    if let subtitle {
                        Text(L10n.string(subtitle))
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                            .fixedSize(horizontal: false, vertical: true)
                    }
                }
            }

            content
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background {
            RoundedRectangle(cornerRadius: HermesTheme.panelCornerRadius, style: .continuous)
                .fill(panelBackground)
        }
        .overlay {
            RoundedRectangle(cornerRadius: HermesTheme.panelCornerRadius, style: .continuous)
                .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
        }
    }

    private var panelBackground: AnyShapeStyle {
        backgroundImageActive
            ? AnyShapeStyle(.regularMaterial)
            : AnyShapeStyle(HermesTheme.panelFill)
    }
}

struct HermesInsetSurface<Content: View>: View {
    let content: Content
    @Environment(\.backgroundImageActive) private var backgroundImageActive

    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    var body: some View {
        content
            .padding(.horizontal, 14)
            .padding(.vertical, 12)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background {
                RoundedRectangle(cornerRadius: HermesTheme.insetCornerRadius, style: .continuous)
                    .fill(insetBackground)
            }
    }

    private var insetBackground: AnyShapeStyle {
        backgroundImageActive
            ? AnyShapeStyle(.thinMaterial)
            : AnyShapeStyle(HermesTheme.insetFill)
    }
}

struct HermesLoadingState: View {
    let label: String
    var minHeight: CGFloat = 300

    var body: some View {
        VStack(spacing: 12) {
            ProgressView()
                .controlSize(.regular)

            Text(L10n.string(label))
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, minHeight: minHeight)
    }
}

struct HermesLoadingOverlay: View {
    var body: some View {
        ProgressView()
            .controlSize(.small)
            .padding(10)
            .background(.ultraThinMaterial, in: RoundedRectangle(cornerRadius: HermesTheme.insetCornerRadius, style: .continuous))
            .overlay {
                RoundedRectangle(cornerRadius: HermesTheme.insetCornerRadius, style: .continuous)
                    .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
            }
    }
}

struct HermesValidationMessage: View {
    let text: String

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 6) {
            Image(systemName: "exclamationmark.circle.fill")
                .font(.caption)
                .foregroundStyle(.orange)

            Text(L10n.string(text))
                .font(.caption)
                .foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)
        }
        .accessibilityElement(children: .combine)
    }
}

struct HermesRefreshButton: View {
    let isRefreshing: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            if isRefreshing {
                HStack(spacing: 8) {
                    ProgressView()
                        .controlSize(.small)

                    Text(L10n.string("Refreshing…"))
                }
            } else {
                Label(L10n.string("Refresh"), systemImage: "arrow.clockwise")
            }
        }
        .buttonStyle(.bordered)
        .controlSize(.regular)
        .help(L10n.string("Refresh"))
        .disabled(isRefreshing)
    }
}

struct HermesCreateActionButton: View {
    let title: String
    let help: String?
    let action: () -> Void

    init(_ title: String, help: String? = nil, action: @escaping () -> Void) {
        self.title = title
        self.help = help
        self.action = action
    }

    var body: some View {
        Button(action: action) {
            Label(L10n.string(title), systemImage: "plus")
        }
        .buttonStyle(.bordered)
        .controlSize(.regular)
        .fixedSize(horizontal: true, vertical: false)
        .help(help.map { L10n.string($0) } ?? L10n.string(title))
    }
}

struct HermesBadge: View {
    let text: String
    let tint: Color
    var systemImage: String?
    var prominence: BadgeProminence = .subtle
    var isMonospaced = false

    var body: some View {
        HStack(spacing: 4) {
            if let systemImage {
                Image(systemName: systemImage)
                    .font(.caption2.weight(.bold))
            }

            Text(L10n.string(text))
                .font(isMonospaced ? .system(.caption2, design: .monospaced).weight(.semibold) : .caption2.weight(.semibold))
        }
        .foregroundStyle(foregroundStyle)
        .lineLimit(1)
        .fixedSize(horizontal: true, vertical: true)
        .padding(.horizontal, 8)
        .padding(.vertical, 3)
        .background(backgroundStyle, in: Capsule())
        .overlay {
            Capsule()
                .strokeBorder(borderStyle, lineWidth: prominence.borderWidth)
        }
    }

    enum BadgeProminence {
        case subtle
        case strong

        var borderWidth: CGFloat {
            switch self {
            case .subtle:
                return 0
            case .strong:
                return 1
            }
        }
    }

    private var foregroundStyle: Color {
        switch prominence {
        case .subtle:
            return tint
        case .strong:
            return .white
        }
    }

    private var backgroundStyle: Color {
        switch prominence {
        case .subtle:
            return tint.opacity(0.10)
        case .strong:
            return tint.opacity(0.86)
        }
    }

    private var borderStyle: Color {
        switch prominence {
        case .subtle:
            return .clear
        case .strong:
            return Color.white.opacity(0.18)
        }
    }
}

struct HermesLabeledValue: View {
    let label: String
    let value: String
    var isMonospaced = false
    var emphasizeValue = false

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(L10n.string(label))
                .font(.caption2.weight(.semibold))
                .foregroundStyle(.secondary)

            Text(value)
                .font(valueFont)
                .foregroundStyle(emphasizeValue ? .primary : .secondary)
                .textSelection(.enabled)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    private var valueFont: Font {
        if isMonospaced {
            return .system(.callout, design: .monospaced)
        }

        return emphasizeValue ? .callout.weight(.semibold) : .callout
    }
}

struct HermesInspectorField: Identifiable, Hashable {
    let id: String
    let label: String
    let value: String
    var isMonospaced = false
    var emphasizeValue = false

    init(
        id: String? = nil,
        label: String,
        value: String,
        isMonospaced: Bool = false,
        emphasizeValue: Bool = false
    ) {
        self.id = id ?? "\(label)|\(value)"
        self.label = label
        self.value = value
        self.isMonospaced = isMonospaced
        self.emphasizeValue = emphasizeValue
    }
}

struct HermesInspectorFieldList: View {
    let fields: [HermesInspectorField]
    var labelWidth: CGFloat = 108

    var body: some View {
        VStack(spacing: 0) {
            ForEach(Array(fields.enumerated()), id: \.element.id) { index, field in
                HermesInspectorFieldRow(field: field, labelWidth: labelWidth)

                if index < fields.count - 1 {
                    Divider()
                        .padding(.leading, labelWidth + 10)
                        .opacity(0.58)
                }
            }
        }
        .background(
            RoundedRectangle(cornerRadius: HermesTheme.insetCornerRadius, style: .continuous)
                .fill(HermesTheme.insetFill)
        )
        .overlay {
            RoundedRectangle(cornerRadius: HermesTheme.insetCornerRadius, style: .continuous)
                .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
        }
    }
}

private struct HermesInspectorFieldRow: View {
    let field: HermesInspectorField
    let labelWidth: CGFloat

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 10) {
            Text(L10n.string(field.label))
                .font(.caption2.weight(.semibold))
                .foregroundStyle(.secondary)
                .lineLimit(1)
                .frame(width: labelWidth, alignment: .leading)

            Text(field.value)
                .font(valueFont)
                .foregroundStyle(field.emphasizeValue ? .primary : .secondary)
                .textSelection(.enabled)
                .fixedSize(horizontal: false, vertical: true)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(.horizontal, 10)
        .padding(.vertical, 8)
    }

    private var valueFont: Font {
        if field.isMonospaced {
            return .system(.callout, design: .monospaced)
        }

        return field.emphasizeValue ? .callout.weight(.semibold) : .callout
    }
}

struct HermesExpandableSearchField: View {
    @Binding var text: String

    var prompt = "Search"
    var collapsedWidth: CGFloat = 34
    var expandedWidth: CGFloat = 240
    var focusRequestID: UUID?

    @FocusState private var isFocused: Bool
    @State private var isExpanded = false

    private var localizedPrompt: String {
        L10n.string(prompt)
    }

    private var shouldShowExpandedField: Bool {
        isExpanded || !text.isEmpty
    }

    var body: some View {
        HStack(spacing: 8) {
            Button {
                withAnimation(.spring(response: 0.24, dampingFraction: 0.88)) {
                    isExpanded = true
                }
                DispatchQueue.main.async {
                    isFocused = true
                }
            } label: {
                Image(systemName: "magnifyingglass")
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(shouldShowExpandedField ? .secondary : .primary)
                    .frame(width: 14, height: 14)
            }
            .buttonStyle(.plain)
            .accessibilityLabel(localizedPrompt)

            if shouldShowExpandedField {
                TextField(localizedPrompt, text: $text)
                    .textFieldStyle(.plain)
                    .font(.subheadline)
                    .focused($isFocused)
                    .submitLabel(.search)
                    .frame(maxWidth: .infinity, alignment: .leading)

                Button {
                    text = ""
                    isFocused = false
                    isExpanded = false
                } label: {
                    Image(systemName: "xmark.circle.fill")
                        .font(.system(size: 12))
                        .foregroundStyle(.tertiary)
                }
                .buttonStyle(.plain)
                .accessibilityLabel(L10n.string("Close search"))
            }
        }
        .padding(.horizontal, 10)
        .frame(width: shouldShowExpandedField ? expandedWidth : collapsedWidth, height: 30, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .fill(HermesTheme.panelFill)
        )
        .overlay {
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .strokeBorder(HermesTheme.subtleStroke, lineWidth: 1)
        }
        .animation(.spring(response: 0.24, dampingFraction: 0.88), value: shouldShowExpandedField)
        .onAppear {
            isExpanded = !text.isEmpty
        }
        .onChange(of: focusRequestID) { _, requestID in
            guard requestID != nil else { return }
            withAnimation(.spring(response: 0.24, dampingFraction: 0.88)) {
                isExpanded = true
            }
            DispatchQueue.main.async {
                isFocused = true
            }
        }
        .onChange(of: isFocused) { _, focused in
            if !focused && text.isEmpty {
                isExpanded = false
            }
        }
    }
}

struct HermesSearchActionBar<LeadingContent: View>: View {
    @Binding var text: String

    var prompt = "Search"
    var collapsedWidth: CGFloat = 34
    var expandedWidth: CGFloat = 240
    var focusRequestID: UUID?
    let leadingContent: LeadingContent

    init(
        text: Binding<String>,
        prompt: String = "Search",
        collapsedWidth: CGFloat = 34,
        expandedWidth: CGFloat = 240,
        focusRequestID: UUID? = nil,
        @ViewBuilder leadingContent: () -> LeadingContent
    ) {
        self._text = text
        self.prompt = prompt
        self.collapsedWidth = collapsedWidth
        self.expandedWidth = expandedWidth
        self.focusRequestID = focusRequestID
        self.leadingContent = leadingContent()
    }

    var body: some View {
        HermesAdaptivePairLayout(
            horizontalSpacing: 12,
            verticalSpacing: 10
        ) {
            leadingContent
            searchField
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var searchField: some View {
        HermesExpandableSearchField(
            text: $text,
            prompt: prompt,
            collapsedWidth: collapsedWidth,
            expandedWidth: expandedWidth,
            focusRequestID: focusRequestID
        )
    }
}

struct HermesSplitLayout: Equatable {
    static let resizeHandleWidth: CGFloat = 1

    let minPrimaryWidth: CGFloat
    let defaultPrimaryWidth: CGFloat
    let maxPrimaryWidth: CGFloat
    var primaryWidth: CGFloat?
    var isPrimaryCollapsed: Bool

    init(
        minPrimaryWidth: CGFloat,
        defaultPrimaryWidth: CGFloat,
        maxPrimaryWidth: CGFloat = 760,
        isPrimaryCollapsed: Bool = false
    ) {
        self.minPrimaryWidth = minPrimaryWidth
        self.defaultPrimaryWidth = defaultPrimaryWidth
        self.maxPrimaryWidth = max(maxPrimaryWidth, minPrimaryWidth)
        self.isPrimaryCollapsed = isPrimaryCollapsed
    }

    mutating func togglePrimaryCollapsed() {
        isPrimaryCollapsed.toggle()
    }

    func expandedWidthRequirement(detailMinWidth: CGFloat) -> CGFloat {
        minPrimaryWidth + detailMinWidth + Self.resizeHandleWidth
    }

    var preferredPrimaryWidth: CGFloat {
        clamped(primaryWidth ?? defaultPrimaryWidth)
    }

    mutating func rememberPrimaryWidth(_ width: CGFloat) {
        guard width.isFinite, width > 0 else { return }

        let clampedWidth = clamped(width)
        if let primaryWidth, abs(primaryWidth - clampedWidth) < 1 {
            return
        }

        primaryWidth = clampedWidth
    }

    mutating func rememberPrimaryWidth(_ width: CGFloat, availableWidth: CGFloat, detailMinWidth: CGFloat) {
        guard width.isFinite, width > 0 else { return }

        let clampedWidth = clampedPrimaryWidth(
            width,
            availableWidth: availableWidth,
            detailMinWidth: detailMinWidth
        )
        if let primaryWidth, abs(primaryWidth - clampedWidth) < 1 {
            return
        }

        primaryWidth = clampedWidth
    }

    private func clamped(_ width: CGFloat) -> CGFloat {
        min(max(width, minPrimaryWidth), maxPrimaryWidth)
    }

    func clampedPrimaryWidth(_ width: CGFloat, availableWidth: CGFloat, detailMinWidth: CGFloat) -> CGFloat {
        guard availableWidth.isFinite, detailMinWidth.isFinite else {
            return clamped(width)
        }

        let availableBeforeDetail = availableWidth - detailMinWidth - Self.resizeHandleWidth
        let upperBound = max(minPrimaryWidth, min(maxPrimaryWidth, availableBeforeDetail))
        return min(max(width, minPrimaryWidth), upperBound)
    }
}

extension View {
    func hermesSplitDetailColumn(minWidth: CGFloat, idealWidth: CGFloat) -> some View {
        frame(
            minWidth: minWidth,
            idealWidth: idealWidth,
            maxWidth: .infinity,
            maxHeight: .infinity,
            alignment: .topLeading
        )
    }
}

struct HermesCollapseToolbarButton: View {
    let systemImage: String
    let isActive: Bool
    let isEnabled: Bool
    let help: String
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            Image(systemName: systemImage)
                .font(.system(size: 14, weight: .medium))
                .symbolVariant(isActive ? .fill : .none)
                .frame(width: 28, height: 28)
                .contentShape(Rectangle())
                .background {
                    RoundedRectangle(cornerRadius: 7, style: .continuous)
                        .fill(isActive ? HermesTheme.selectedFill : .clear)
                }
                .overlay {
                    RoundedRectangle(cornerRadius: 7, style: .continuous)
                        .strokeBorder(isActive ? HermesTheme.selectedStroke : .clear, lineWidth: 1)
                }
        }
        .buttonStyle(.borderless)
        .foregroundStyle(isActive ? Color.accentColor : Color.primary.opacity(isEnabled ? 0.82 : 0.32))
        .disabled(!isEnabled)
        .help(help)
        .accessibilityLabel(Text(help))
    }
}

struct HermesToolbarControlCluster<Content: View>: View {
    let content: Content

    init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    var body: some View {
        HStack(spacing: 6) {
            content
        }
    }
}

struct HermesToolbarPrincipalTitle: View {
    let title: String

    var body: some View {
        Text(L10n.string(title))
            .font(.title3.weight(.semibold))
            .foregroundStyle(.primary.opacity(0.96))
            .lineLimit(1)
            .minimumScaleFactor(0.85)
    }
}

final class HermesTitleBarConfiguratorView: NSView {
    var backgroundImageActive = false {
        didSet {
            configureWindow()
        }
    }
    var windowOpacity = AppWindowOpacityPreference.defaultValue {
        didSet {
            configureWindow()
        }
    }
    var windowMaterial: AppWindowMaterialPreference = .solid {
        didSet {
            configureWindow()
        }
    }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        configureWindow()
    }

    private func configureWindow() {
        guard let window else { return }
        let clampedOpacity = CGFloat(AppWindowOpacityPreference.clamped(windowOpacity))
        let needsTransparentWindow = backgroundImageActive || windowMaterial == .translucent || clampedOpacity < 1.0
        window.titleVisibility = .hidden
        window.alphaValue = clampedOpacity
        window.titlebarAppearsTransparent = backgroundImageActive || windowMaterial == .translucent
        window.isOpaque = !needsTransparentWindow
        window.backgroundColor = needsTransparentWindow ? .clear : .windowBackgroundColor
    }
}

struct HermesWindowTitleBarConfigurator: NSViewRepresentable {
    let backgroundImageActive: Bool
    let windowOpacity: Double
    let windowMaterial: AppWindowMaterialPreference

    func makeNSView(context: Context) -> HermesTitleBarConfiguratorView {
        HermesTitleBarConfiguratorView(frame: .zero)
    }

    func updateNSView(_ nsView: HermesTitleBarConfiguratorView, context: Context) {
        nsView.backgroundImageActive = backgroundImageActive
        nsView.windowOpacity = windowOpacity
        nsView.windowMaterial = windowMaterial
    }
}

struct HermesCollapsibleHSplitView<Primary: View, Detail: View>: View {
    @Binding var layout: HermesSplitLayout
    let detailMinWidth: CGFloat
    let usesTransition: Bool
    let keepsSplitViewWhenCollapsed: Bool
    let primary: Primary
    let detail: Detail
    private let collapseAnimation = Animation.snappy(duration: 0.16, extraBounce: 0)
    private let resizeUpdateStep: CGFloat = 1
    @State private var activeResizeStartingWidth: CGFloat?
    @State private var activeResizeLiveWidth: CGFloat?

    init(
        layout: Binding<HermesSplitLayout>,
        detailMinWidth: CGFloat,
        usesTransition: Bool = true,
        keepsSplitViewWhenCollapsed: Bool = false,
        @ViewBuilder primary: () -> Primary,
        @ViewBuilder detail: () -> Detail
    ) {
        self._layout = layout
        self.detailMinWidth = detailMinWidth
        self.usesTransition = usesTransition
        self.keepsSplitViewWhenCollapsed = keepsSplitViewWhenCollapsed
        self.primary = primary()
        self.detail = detail()
    }

    var body: some View {
        GeometryReader { geometry in
            let availableWidth = geometry.size.width
            let showsPrimary = !layout.isPrimaryCollapsed
            let primaryWidth = activeResizeLiveWidth ?? resolvedPrimaryWidth(in: availableWidth)

            HStack(spacing: 0) {
                if showsPrimary || keepsSplitViewWhenCollapsed {
                    primary
                        .frame(width: showsPrimary ? primaryWidth : 0)
                        .opacity(showsPrimary ? 1 : 0)
                        .allowsHitTesting(showsPrimary)
                        .accessibilityHidden(!showsPrimary)
                        .clipped()
                        .transition(usesTransition ? .move(edge: .leading).combined(with: .opacity) : .identity)
                }

                if showsPrimary {
                    HermesSplitResizeHandle(
                        onDragStart: {
                            activeResizeStartingWidth = layout.primaryWidth ?? layout.preferredPrimaryWidth
                        },
                        onDrag: { translation in
                            let startWidth = activeResizeStartingWidth ?? layout.preferredPrimaryWidth
                            let targetWidth = layout.clampedPrimaryWidth(
                                startWidth + translation,
                                availableWidth: availableWidth,
                                detailMinWidth: detailMinWidth
                            )
                            let nextWidth = roundedResizeWidth(targetWidth)
                            let currentWidth = activeResizeLiveWidth ?? startWidth
                            guard abs(currentWidth - nextWidth) >= resizeUpdateStep else { return }

                            var transaction = Transaction()
                            transaction.disablesAnimations = true
                            withTransaction(transaction) {
                                activeResizeLiveWidth = nextWidth
                            }
                        },
                        onDragEnd: {
                            let finalWidth = activeResizeLiveWidth ?? layout.primaryWidth ?? layout.preferredPrimaryWidth
                            activeResizeStartingWidth = nil
                            activeResizeLiveWidth = nil
                            layout.rememberPrimaryWidth(
                                finalWidth,
                                availableWidth: availableWidth,
                                detailMinWidth: detailMinWidth
                            )
                        }
                    )
                    .transition(usesTransition ? .opacity : .identity)
                }

                detail
                    .frame(minWidth: 0, maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                    .layoutPriority(1)
                    .clipped()
            }
            .frame(width: availableWidth, height: geometry.size.height, alignment: .leading)
            .animation(usesTransition && activeResizeStartingWidth == nil ? collapseAnimation : nil, value: showsPrimary)
        }
        .clipped()
    }

    private func resolvedPrimaryWidth(in availableWidth: CGFloat) -> CGFloat {
        layout.clampedPrimaryWidth(
            layout.preferredPrimaryWidth,
            availableWidth: availableWidth,
            detailMinWidth: detailMinWidth
        )
    }

    private func roundedResizeWidth(_ width: CGFloat) -> CGFloat {
        (width / resizeUpdateStep).rounded() * resizeUpdateStep
    }
}

private struct HermesSplitResizeHandle: View {
    static let width = HermesSplitLayout.resizeHandleWidth
    private static let hitWidth: CGFloat = 18
    private static let activeVisibleWidth: CGFloat = 3

    let onDragStart: () -> Void
    let onDrag: (CGFloat) -> Void
    let onDragEnd: () -> Void

    @State private var isHovering = false
    @State private var isDragging = false

    var body: some View {
        let hitOutset = max(0, (Self.hitWidth - Self.width) / 2)

        ZStack {
            Capsule(style: .continuous)
                .fill(handleFill)
                .frame(width: isDragging || isHovering ? Self.activeVisibleWidth : Self.width)
        }
            .frame(minWidth: Self.width, idealWidth: Self.width, maxWidth: Self.width, maxHeight: .infinity)
            .padding(.horizontal, hitOutset)
            .contentShape(Rectangle())
            .background {
                HermesResizeCursorRegion()
            }
            .gesture(
                DragGesture(minimumDistance: 1)
                    .onChanged { value in
                        if !isDragging {
                            isDragging = true
                            onDragStart()
                        }
                        onDrag(value.translation.width)
                    }
                    .onEnded { _ in
                        isDragging = false
                        onDragEnd()
                    }
            )
            .onHover { hovering in
                guard hovering != isHovering else { return }
                isHovering = hovering
            }
            .padding(.horizontal, -hitOutset)
            .animation(.easeOut(duration: 0.12), value: isHovering)
            .animation(.easeOut(duration: 0.12), value: isDragging)
            .accessibilityLabel(Text("Resize pane"))
            .accessibilityHint(Text("Drag horizontally to resize this column"))
    }

    private var handleFill: Color {
        return Color.secondary.opacity(isDragging ? 0.28 : (isHovering ? 0.18 : 0.08))
    }
}

private struct HermesResizeCursorRegion: NSViewRepresentable {
    func makeNSView(context: Context) -> ResizeCursorView {
        ResizeCursorView()
    }

    func updateNSView(_ nsView: ResizeCursorView, context: Context) {
        nsView.window?.invalidateCursorRects(for: nsView)
    }

    final class ResizeCursorView: NSView {
        override func resetCursorRects() {
            addCursorRect(bounds, cursor: .resizeLeftRight)
        }
    }
}

struct HermesWrappingFlowLayout: Layout {
    let horizontalSpacing: CGFloat
    let verticalSpacing: CGFloat

    init(horizontalSpacing: CGFloat = 8, verticalSpacing: CGFloat = 8) {
        self.horizontalSpacing = horizontalSpacing
        self.verticalSpacing = verticalSpacing
    }

    func sizeThatFits(
        proposal: ProposedViewSize,
        subviews: Subviews,
        cache: inout Void
    ) -> CGSize {
        let sizes = subviews.map { $0.sizeThatFits(.unspecified) }
        let lines = computeLines(for: sizes, maxWidth: proposal.width)
        let height = lines.reduce(CGFloat.zero) { partial, line in
            partial + line.height
        } + verticalSpacing * CGFloat(max(0, lines.count - 1))
        let width = proposal.width ?? lines.map(\.width).max() ?? 0
        return CGSize(width: width, height: height)
    }

    func placeSubviews(
        in bounds: CGRect,
        proposal: ProposedViewSize,
        subviews: Subviews,
        cache: inout Void
    ) {
        let sizes = subviews.map { $0.sizeThatFits(.unspecified) }
        let lines = computeLines(for: sizes, maxWidth: bounds.width)
        var currentY = bounds.minY

        for line in lines {
            var currentX = bounds.minX
            for item in line.items {
                let size = sizes[item.index]
                subviews[item.index].place(
                    at: CGPoint(x: currentX, y: currentY),
                    proposal: ProposedViewSize(width: size.width, height: size.height)
                )
                currentX += size.width + horizontalSpacing
            }
            currentY += line.height + verticalSpacing
        }
    }

    private func computeLines(for sizes: [CGSize], maxWidth: CGFloat?) -> [HermesFlowLine] {
        let availableWidth = maxWidth ?? .greatestFiniteMagnitude
        guard !sizes.isEmpty else { return [] }

        var lines: [HermesFlowLine] = []
        var currentItems: [HermesFlowLineItem] = []
        var currentWidth: CGFloat = 0
        var currentHeight: CGFloat = 0

        for (index, size) in sizes.enumerated() {
            let proposedWidth = currentItems.isEmpty ? size.width : currentWidth + horizontalSpacing + size.width

            if !currentItems.isEmpty && proposedWidth > availableWidth {
                lines.append(
                    HermesFlowLine(
                        items: currentItems,
                        width: currentWidth,
                        height: currentHeight
                    )
                )
                currentItems = [HermesFlowLineItem(index: index)]
                currentWidth = size.width
                currentHeight = size.height
            } else {
                currentItems.append(HermesFlowLineItem(index: index))
                currentWidth = proposedWidth
                currentHeight = max(currentHeight, size.height)
            }
        }

        if !currentItems.isEmpty {
            lines.append(
                HermesFlowLine(
                    items: currentItems,
                    width: currentWidth,
                    height: currentHeight
                )
            )
        }

        return lines
    }
}

private struct HermesFlowLine {
    let items: [HermesFlowLineItem]
    let width: CGFloat
    let height: CGFloat
}

private struct HermesFlowLineItem {
    let index: Int
}
