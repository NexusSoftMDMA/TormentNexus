.PHONY: install install-dev build test clean

# Install to ~/.cargo/bin/mnem (globally callable).
# Re-run after pulling changes to update the system-wide binary.
install:
	cargo install --path crates/mnem-cli --features bundled-embedder

# Install without bundled embedder (bring your own via config.toml).
install-bare:
	cargo install --path crates/mnem-cli

# Fast debug build (./target/debug/mnem). Does NOT update global binary.
build:
	cargo build -p mnem-cli

# Release build (./target/release/mnem). Does NOT update global binary.
build-release:
	cargo build --release -p mnem-cli --features bundled-embedder

# Run all tests.
test:
	cargo test -p mnem-cli

# Full clean rebuild then install.
reinstall: clean install

clean:
	cargo clean -p mnem-cli
