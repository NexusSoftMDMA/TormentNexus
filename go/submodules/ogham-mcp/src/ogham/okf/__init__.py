"""Open Knowledge Format (OKF) v0.1 support for Ogham.

OKF spec: docs/snapshots/okf/v0.1/SPEC.md
"""

from ogham.okf.bundle import export_okf_bundle, import_okf_bundle

SUPPORTED_OKF_VERSION = "0.1"

__all__ = ["SUPPORTED_OKF_VERSION", "export_okf_bundle", "import_okf_bundle"]
