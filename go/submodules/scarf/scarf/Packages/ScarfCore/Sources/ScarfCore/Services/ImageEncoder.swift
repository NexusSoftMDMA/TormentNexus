import Foundation
#if canImport(AppKit)
import AppKit
#endif
#if canImport(UIKit)
import UIKit
#endif
#if canImport(CoreImage)
import CoreImage
#endif

/// Downsamples + base64-encodes user-supplied images for ACP transport.
///
/// **Why downsample on the producer side.** Hermes happily forwards the
/// bytes to a vision model, but a 12 MP screenshot at 4 MB is wasteful
/// — it eats 5–6× more tokens than a 1024×1024 thumbnail and gives the
/// model no extra signal. Cap the long edge at 1568 px (Anthropic's
/// recommended max for Claude vision) and drop quality to JPEG 0.85,
/// which keeps screenshot text crisp while landing under ~300 KB per
/// image. The 5-image-per-message limit (chosen on the producer side)
/// keeps the total prompt payload below ~2 MB.
///
/// **Why detached.** Image loading + downsampling is CPU-bound. Run only
/// from a `Task.detached` context (the encoder type is `Sendable` and
/// every method is `nonisolated`). The companion `ChatImageAttachment`
/// is a Sendable value type so the result hops back to MainActor cleanly.
public struct ImageEncoder: Sendable {
    /// Long-edge pixel cap. 1568 is Anthropic's recommended ceiling for
    /// Claude vision input — past it, the provider downsamples server-side
    /// and we just paid for the extra bytes. Tweak only with vision-model
    /// guidance from Hermes side.
    public static let maxLongEdge: CGFloat = 1568
    /// JPEG quality factor. 0.85 is the inflection point above which
    /// file size jumps quickly without obvious visual gain on screenshots
    /// or photographs.
    public static let jpegQuality: CGFloat = 0.85
    /// Long-edge cap for the inline thumbnail rendered in the composer
    /// chip. Kept under the system thumbnail size so `Image(data:)`
    /// renders without extra resampling.
    public static let thumbnailLongEdge: CGFloat = 256

    public init() {}

    public enum EncoderError: Error, LocalizedError {
        case unsupportedFormat
        case decodeFailed
        case encodeFailed
        case empty

        public var errorDescription: String? {
            switch self {
            case .unsupportedFormat: return "Image format not recognized"
            case .decodeFailed: return "Couldn't decode image data"
            case .encodeFailed: return "Couldn't encode image as JPEG"
            case .empty: return "Image data was empty"
            }
        }
    }

    /// Encode raw bytes (from a paste/drop/picker) into a wire-ready
    /// attachment. Detached-only — never call from MainActor. The
    /// originating bytes are not retained beyond this call.
    public nonisolated func encode(
        rawBytes: Data,
        sourceFilename: String? = nil
    ) throws -> ChatImageAttachment {
        guard !rawBytes.isEmpty else { throw EncoderError.empty }
        ScarfMon.event(.render, "imageEncoder.input.bytes", count: 1, bytes: rawBytes.count)
        return try ScarfMon.measure(.render, "imageEncoder.downsample") {
        #if canImport(AppKit)
        guard let nsImage = NSImage(data: rawBytes) else { throw EncoderError.decodeFailed }
        let targetSize = Self.fittedSize(for: nsImage.size, maxLongEdge: Self.maxLongEdge)
        let mainData = try Self.jpegBytes(from: nsImage, size: targetSize)
        let thumbSize = Self.fittedSize(for: nsImage.size, maxLongEdge: Self.thumbnailLongEdge)
        let thumbData = try? Self.jpegBytes(from: nsImage, size: thumbSize)
        ScarfMon.event(.render, "imageEncoder.bytes", count: 1, bytes: mainData.count)
        return ChatImageAttachment(
            mimeType: "image/jpeg",
            base64Data: mainData.base64EncodedString(),
            thumbnailBase64: thumbData?.base64EncodedString(),
            filename: sourceFilename,
            approximateByteCount: mainData.count
        )

        #elseif canImport(UIKit)
        guard let uiImage = UIImage(data: rawBytes) else { throw EncoderError.decodeFailed }
        let targetSize = Self.fittedSize(for: uiImage.size, maxLongEdge: Self.maxLongEdge)
        let mainData = try Self.jpegBytes(from: uiImage, size: targetSize)
        let thumbSize = Self.fittedSize(for: uiImage.size, maxLongEdge: Self.thumbnailLongEdge)
        let thumbData = try? Self.jpegBytes(from: uiImage, size: thumbSize)
        ScarfMon.event(.render, "imageEncoder.bytes", count: 1, bytes: mainData.count)
        return ChatImageAttachment(
            mimeType: "image/jpeg",
            base64Data: mainData.base64EncodedString(),
            thumbnailBase64: thumbData?.base64EncodedString(),
            filename: sourceFilename,
            approximateByteCount: mainData.count
        )

        #else
        // Linux CI / unknown platforms: pass through raw bytes if the
        // input already looks like a JPEG, else refuse. Keeps the
        // package compiling without a hard AppKit/UIKit dep.
        if rawBytes.starts(with: [0xFF, 0xD8]) {
            ScarfMon.event(.render, "imageEncoder.bytes", count: 1, bytes: rawBytes.count)
            return ChatImageAttachment(
                mimeType: "image/jpeg",
                base64Data: rawBytes.base64EncodedString(),
                thumbnailBase64: nil,
                filename: sourceFilename,
                approximateByteCount: rawBytes.count
            )
        }
        throw EncoderError.unsupportedFormat
        #endif
        }
    }

    nonisolated private static func fittedSize(for source: CGSize, maxLongEdge: CGFloat) -> CGSize {
        let longest = max(source.width, source.height)
        if longest <= maxLongEdge { return source }
        let scale = maxLongEdge / longest
        return CGSize(
            width: floor(source.width * scale),
            height: floor(source.height * scale)
        )
    }

    #if canImport(AppKit)
    nonisolated private static func jpegBytes(from image: NSImage, size: CGSize) throws -> Data {
        let resized = NSImage(size: size)
        resized.lockFocus()
        NSGraphicsContext.current?.imageInterpolation = .high
        image.draw(
            in: CGRect(origin: .zero, size: size),
            from: .zero,
            operation: .copy,
            fraction: 1.0
        )
        resized.unlockFocus()
        guard let tiff = resized.tiffRepresentation,
              let rep = NSBitmapImageRep(data: tiff),
              let data = rep.representation(
                using: .jpeg,
                properties: [.compressionFactor: jpegQuality]
              )
        else {
            throw EncoderError.encodeFailed
        }
        return data
    }
    #elseif canImport(UIKit)
    nonisolated private static func jpegBytes(from image: UIImage, size: CGSize) throws -> Data {
        let format = UIGraphicsImageRendererFormat()
        format.scale = 1
        format.opaque = true
        let renderer = UIGraphicsImageRenderer(size: size, format: format)
        let resized = renderer.image { _ in
            image.draw(in: CGRect(origin: .zero, size: size))
        }
        guard let data = resized.jpegData(compressionQuality: jpegQuality) else {
            throw EncoderError.encodeFailed
        }
        return data
    }
    #endif
}
