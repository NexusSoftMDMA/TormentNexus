//! Error types for the NER provider layer.

use thiserror::Error;

/// Error returned by NER provider operations.
#[derive(Debug, Error)]
pub enum NerError {
    /// The requested NER provider is not compiled in or cannot be loaded.
    #[error("NER provider not available: {0}")]
    NotAvailable(String),
    /// The NER provider configuration is invalid.
    #[error("NER provider configuration error: {0}")]
    Config(String),
}
