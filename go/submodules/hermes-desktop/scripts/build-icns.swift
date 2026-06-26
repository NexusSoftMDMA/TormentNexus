#!/usr/bin/swift

import Foundation

enum ICNSBuilderError: Error {
    case invalidArguments
    case invalidChunkType(String)
}

func appendUInt32(_ value: UInt32, to data: inout Data) {
    var bigEndian = value.bigEndian
    withUnsafeBytes(of: &bigEndian) { bytes in
        data.append(contentsOf: bytes)
    }
}

func run() throws {
    let arguments = CommandLine.arguments
    guard arguments.count == 3 else {
        fputs("usage: build-icns.swift /path/to/AppIcon.iconset /path/to/AppIcon.icns\n", stderr)
        throw ICNSBuilderError.invalidArguments
    }

    let iconsetURL = URL(fileURLWithPath: arguments[1], isDirectory: true)
    let outputURL = URL(fileURLWithPath: arguments[2])

    let mappings: [(chunkType: String, fileName: String)] = [
        ("ic04", "icon_16x16.png"),
        ("ic11", "icon_16x16@2x.png"),
        ("ic05", "icon_32x32.png"),
        ("ic12", "icon_32x32@2x.png"),
        ("ic07", "icon_128x128.png"),
        ("ic13", "icon_128x128@2x.png"),
        ("ic08", "icon_256x256.png"),
        ("ic14", "icon_256x256@2x.png"),
        ("ic09", "icon_512x512.png"),
        ("ic10", "icon_512x512@2x.png")
    ]

    var chunks = Data()

    for mapping in mappings {
        guard mapping.chunkType.utf8.count == 4 else {
            throw ICNSBuilderError.invalidChunkType(mapping.chunkType)
        }

        let fileURL = iconsetURL.appendingPathComponent(mapping.fileName)
        let pngData = try Data(contentsOf: fileURL)

        chunks.append(contentsOf: mapping.chunkType.utf8)
        appendUInt32(UInt32(pngData.count + 8), to: &chunks)
        chunks.append(pngData)
    }

    var icnsData = Data()
    icnsData.append(contentsOf: "icns".utf8)
    appendUInt32(UInt32(chunks.count + 8), to: &icnsData)
    icnsData.append(chunks)

    try FileManager.default.createDirectory(
        at: outputURL.deletingLastPathComponent(),
        withIntermediateDirectories: true,
        attributes: nil
    )
    try icnsData.write(to: outputURL)
}

do {
    try run()
} catch {
    fputs("error: \(error)\n", stderr)
    exit(1)
}
