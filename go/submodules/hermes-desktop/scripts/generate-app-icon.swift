#!/usr/bin/swift

import AppKit

enum IconGeneratorError: Error {
    case invalidArguments
    case cannotEncodePNG
}

func run() throws {
    let arguments = CommandLine.arguments
    guard arguments.count == 2 else {
        fputs("usage: generate-app-icon.swift /absolute/path/to/output.png\n", stderr)
        throw IconGeneratorError.invalidArguments
    }

    let outputURL = URL(fileURLWithPath: arguments[1])
    let size = NSSize(width: 1024, height: 1024)
    let image = NSImage(size: size)

    image.lockFocus()

    let canvas = NSRect(origin: .zero, size: size)
    NSColor.clear.setFill()
    canvas.fill()

    let outerShadow = NSShadow()
    outerShadow.shadowColor = NSColor(calibratedWhite: 0.0, alpha: 0.24)
    outerShadow.shadowBlurRadius = 42
    outerShadow.shadowOffset = NSSize(width: 0, height: -20)
    outerShadow.set()

    let iconRect = canvas.insetBy(dx: 76, dy: 76)
    let iconPath = NSBezierPath(roundedRect: iconRect, xRadius: 238, yRadius: 238)
    let backgroundGradient = NSGradient(
        colorsAndLocations:
            (NSColor(calibratedRed: 0.08, green: 0.16, blue: 0.24, alpha: 1.0), 0.0),
            (NSColor(calibratedRed: 0.12, green: 0.33, blue: 0.46, alpha: 1.0), 0.55),
            (NSColor(calibratedRed: 0.20, green: 0.48, blue: 0.61, alpha: 1.0), 1.0)
    )!
    backgroundGradient.draw(in: iconPath, angle: 130)

    let highlightPath = NSBezierPath(roundedRect: iconRect, xRadius: 238, yRadius: 238)
    NSGraphicsContext.saveGraphicsState()
    highlightPath.addClip()
    let highlightRect = NSRect(
        x: iconRect.minX,
        y: iconRect.midY,
        width: iconRect.width,
        height: iconRect.height / 1.5
    )
    let highlightGradient = NSGradient(
        colorsAndLocations:
            (NSColor(calibratedWhite: 1.0, alpha: 0.18), 0.0),
            (NSColor(calibratedWhite: 1.0, alpha: 0.0), 1.0)
    )!
    highlightGradient.draw(in: highlightRect, angle: 90)
    NSGraphicsContext.restoreGraphicsState()

    let panelRect = iconRect.insetBy(dx: 104, dy: 120)
    let panelPath = NSBezierPath(roundedRect: panelRect, xRadius: 120, yRadius: 120)
    NSColor(calibratedWhite: 1.0, alpha: 0.08).setFill()
    panelPath.fill()
    NSColor(calibratedWhite: 1.0, alpha: 0.18).setStroke()
    panelPath.lineWidth = 8
    panelPath.stroke()

    let monogramColor = NSColor(calibratedRed: 0.97, green: 0.94, blue: 0.88, alpha: 1.0)

    func roundedBar(x: CGFloat, y: CGFloat, width: CGFloat, height: CGFloat, radius: CGFloat) -> NSBezierPath {
        NSBezierPath(
            roundedRect: NSRect(x: x, y: y, width: width, height: height),
            xRadius: radius,
            yRadius: radius
        )
    }

    let barWidth: CGFloat = 92
    let barHeight: CGFloat = 382
    let barRadius: CGFloat = 46
    let leftBarX = canvas.midX - 176
    let rightBarX = canvas.midX + 84
    let barY = canvas.midY - 174

    monogramColor.setFill()
    roundedBar(x: leftBarX, y: barY, width: barWidth, height: barHeight, radius: barRadius).fill()
    roundedBar(x: rightBarX, y: barY, width: barWidth, height: barHeight, radius: barRadius).fill()

    let crossbar = roundedBar(
        x: leftBarX + barWidth - 14,
        y: canvas.midY - 42,
        width: (rightBarX - leftBarX) - barWidth + 28,
        height: 84,
        radius: 42
    )
    crossbar.fill()

    let promptRect = NSRect(x: canvas.midX - 32, y: canvas.midY - 18, width: 124, height: 36)
    let promptPath = NSBezierPath(roundedRect: promptRect, xRadius: 18, yRadius: 18)
    NSColor(calibratedRed: 0.31, green: 0.91, blue: 0.82, alpha: 1.0).setFill()
    promptPath.fill()

    let inkColor = NSColor(calibratedRed: 0.05, green: 0.18, blue: 0.22, alpha: 1.0)

    let promptArrow = NSBezierPath()
    promptArrow.move(to: NSPoint(x: promptRect.minX + 24, y: promptRect.midY))
    promptArrow.line(to: NSPoint(x: promptRect.minX + 44, y: promptRect.midY + 10))
    promptArrow.move(to: NSPoint(x: promptRect.minX + 24, y: promptRect.midY))
    promptArrow.line(to: NSPoint(x: promptRect.minX + 44, y: promptRect.midY - 10))
    promptArrow.lineCapStyle = .round
    promptArrow.lineJoinStyle = .round
    promptArrow.lineWidth = 8
    inkColor.setStroke()
    promptArrow.stroke()

    let promptCursor = NSBezierPath(
        roundedRect: NSRect(x: promptRect.minX + 62, y: promptRect.midY - 6, width: 26, height: 12),
        xRadius: 6,
        yRadius: 6
    )
    inkColor.setFill()
    promptCursor.fill()

    let statusDot = NSBezierPath(ovalIn: NSRect(x: panelRect.maxX - 90, y: panelRect.minY + 56, width: 34, height: 34))
    NSColor(calibratedRed: 1.0, green: 0.77, blue: 0.32, alpha: 1.0).setFill()
    statusDot.fill()

    image.unlockFocus()

    guard
        let tiffData = image.tiffRepresentation,
        let bitmap = NSBitmapImageRep(data: tiffData),
        let pngData = bitmap.representation(using: .png, properties: [:])
    else {
        throw IconGeneratorError.cannotEncodePNG
    }

    try FileManager.default.createDirectory(
        at: outputURL.deletingLastPathComponent(),
        withIntermediateDirectories: true,
        attributes: nil
    )
    try pngData.write(to: outputURL)
}

do {
    try run()
} catch {
    fputs("error: \(error)\n", stderr)
    exit(1)
}
