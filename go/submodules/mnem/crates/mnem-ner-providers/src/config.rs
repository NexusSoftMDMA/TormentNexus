//! NER provider configuration and `open()` factory.

use serde::{Deserialize, Serialize};

use crate::error::NerError;
use crate::null::NullNer;
use crate::provider::NerProvider;
use crate::rule::RuleNer;

/// NER provider selection.
///
/// Serialised under the `[ner]` section of `config.toml`:
/// ```toml
/// [ner]
/// provider = "rule"   # default
/// # or
/// provider = "none"   # disables NER entirely
/// ```
#[derive(Debug, Clone, Default, PartialEq, Eq, Serialize, Deserialize)]
#[serde(tag = "provider", rename_all = "lowercase")]
pub enum NerConfig {
    /// Capitalized-phrase heuristic. Zero dependencies. Default.
    #[default]
    Rule,
    /// Suppress all entity extraction. No entity nodes are emitted.
    None,
}

/// Open a boxed [`NerProvider`] from a [`NerConfig`].
///
/// # Errors
///
/// Returns [`NerError`] if the config requests an unavailable provider
/// (e.g. a compiled-out ONNX feature). Neither `Rule` nor `None`
/// can fail.
pub fn open(cfg: &NerConfig) -> Result<Box<dyn NerProvider>, NerError> {
    match cfg {
        NerConfig::Rule => Ok(Box::new(RuleNer)),
        NerConfig::None => Ok(Box::new(NullNer)),
    }
}
