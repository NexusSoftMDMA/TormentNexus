//! `NerProvider` trait and `NamedEntity` span type.

/// A single named entity span returned by a [`NerProvider`].
#[derive(Debug, Clone, PartialEq)]
pub struct NamedEntity {
    /// Verbatim text of the entity mention.
    pub text: String,
    /// Entity ntype label (one of [`crate::labels`] constants, or any
    /// free-form namespaced string for custom providers).
    pub label: String,
    /// Byte start offset within the input text passed to [`NerProvider::extract`].
    pub byte_start: usize,
    /// Byte end offset (exclusive).
    pub byte_end: usize,
    /// Heuristic confidence in `[0.0, 1.0]`.
    pub confidence: f32,
}

/// Pluggable named-entity classification provider.
///
/// Implementations must be `Send + Sync` to satisfy the `Extractor`
/// contract in `mnem-ingest`.
pub trait NerProvider: Send + Sync {
    /// Extract named entity spans from a text string.
    fn extract(&self, text: &str) -> Vec<NamedEntity>;

    /// Human-readable provider identifier (e.g. `"rule"`, `"null"`).
    fn provider_id(&self) -> &str;
}
