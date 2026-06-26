//! No-op NER provider. Used when `[ner]\nprovider = "none"` is configured.

use crate::provider::{NamedEntity, NerProvider};

/// A [`NerProvider`] that never emits any entities.
///
/// Useful when the caller wants to suppress all entity extraction
/// (e.g. for code files, config files, or structured data where the
/// capitalized-phrase heuristic would produce noise).
#[derive(Debug, Default, Clone)]
pub struct NullNer;

impl NerProvider for NullNer {
    fn extract(&self, _text: &str) -> Vec<NamedEntity> {
        Vec::new()
    }

    fn provider_id(&self) -> &str {
        "null"
    }
}
