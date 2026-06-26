import Foundation

/// One image attached to an outgoing chat prompt.
///
/// Hermes v0.12 ACP advertises `prompt_capabilities.image = true` and
/// accepts content-block arrays in `session/prompt`. Scarf produces these
/// blocks from drag-dropped / pasted / picker-selected images. We
/// downsample + JPEG-encode at the producer side so the wire payload
/// stays under a few hundred kilobytes per image even when the user
/// drops a 12 MP screenshot.
///
/// Constructed via `ImageEncoder.encode(...)`. The store-the-bytes-once
/// shape means `RichChatViewModel` can keep the array between turns
/// (e.g. while the agent is responding) without holding `NSImage` /
/// `UIImage` references that would pin the originals in memory.
public struct ChatImageAttachment: Sendable, Equatable, Identifiable {
    public let id: String
    /// IANA MIME type — matches the `mimeType` field on ACP `ImageContentBlock`.
    /// Currently always `image/jpeg` after re-encoding; PNG-only originals
    /// keep their type when small enough to skip the JPEG step.
    public let mimeType: String
    /// Base64-encoded payload. NOT prefixed with `data:` — Hermes wraps it
    /// when forwarding to OpenAI multimodal payloads (see
    /// `_image_block_to_openai_part` in `acp_adapter/server.py`).
    public let base64Data: String
    /// Small inline thumbnail for the composer's preview strip. Same MIME
    /// type as `base64Data`. Nil when the source was already small enough
    /// to use directly.
    public let thumbnailBase64: String?
    /// Original filename, when known (drag-drop carries it; paste doesn't).
    /// Surfaced as a tooltip on the preview chip.
    public let filename: String?
    /// Approximate decoded byte count, kept for the composer's
    /// "X images, Y KB" status pill.
    public let approximateByteCount: Int

    public init(
        id: String = UUID().uuidString,
        mimeType: String,
        base64Data: String,
        thumbnailBase64: String?,
        filename: String?,
        approximateByteCount: Int
    ) {
        self.id = id
        self.mimeType = mimeType
        self.base64Data = base64Data
        self.thumbnailBase64 = thumbnailBase64
        self.filename = filename
        self.approximateByteCount = approximateByteCount
    }
}
