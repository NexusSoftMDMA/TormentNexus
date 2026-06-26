//! Integration tests for `mnem embedding get` and `mnem embedding ls`.
//!
//! These tests seed a repository with a node + embedding directly via the
//! Rust API (no HTTP, no external embedder) and then invoke the CLI binary
//! to verify the read path.

use std::path::Path;
use std::process::Command;

use assert_cmd::prelude::*;
use bytes::Bytes;
use mnem_core::id::NodeId;
use mnem_core::objects::{Dtype, Embedding, Node};
use mnem_core::repo::ReadonlyRepo;
use mnem_core::store::{Blockstore, OpHeadsStore};
use tempfile::TempDir;

/// Open a redb-backed repo in `dir/.mnem/repo.redb`, seed one node with an
/// embedding, and return (repo_dir, node_uuid_string, model_string).
fn seed_repo_with_embedding(dir: &TempDir) -> (String, String) {
    use std::sync::Arc;

    let data_dir = dir.path().join(".mnem");
    std::fs::create_dir_all(&data_dir).expect("create .mnem");
    let db_path = data_dir.join("repo.redb");

    let (bs, ohs, _db) = mnem_backend_redb::open_or_init(&db_path).expect("open redb");
    let bs_arc: Arc<dyn Blockstore> = bs;
    let ohs_arc: Arc<dyn OpHeadsStore> = ohs;

    let repo = ReadonlyRepo::init(bs_arc, ohs_arc).expect("init repo");

    let model = "onnx:all-MiniLM-L6-v2".to_string();
    let dim: usize = 4; // tiny dim so stdout is short
    let vector: Vec<f32> = (0..dim).map(|i| (i + 1) as f32 / dim as f32).collect();
    let mut vector_bytes = Vec::with_capacity(dim * 4);
    for v in &vector {
        vector_bytes.extend_from_slice(&v.to_le_bytes());
    }

    let node = Node::new(NodeId::new_v7(), "Memory").with_summary("embedding test node");
    let node_id_str = node.id.to_uuid_string();

    let emb = Embedding {
        model: model.clone(),
        dtype: Dtype::F32,
        dim: dim as u32,
        vector: Bytes::from(vector_bytes),
    };

    let mut tx = repo.start_transaction();
    let node_cid = tx.add_node(&node).expect("add node");
    tx.set_embedding(node_cid, model.clone(), emb)
        .expect("set embedding");
    tx.commit("tests", "seed embedding").expect("commit");

    (node_id_str, model)
}

fn mnem(repo: &Path, args: &[&str]) -> Command {
    let mut cmd = Command::cargo_bin("mnem").expect("built mnem binary");
    cmd.current_dir(repo);
    cmd.arg("-R").arg(repo);
    for a in args {
        cmd.arg(a);
    }
    cmd
}

#[test]
fn embedding_get_prints_vector_to_stdout() {
    let dir = TempDir::new().unwrap();
    let (node_id, model) = seed_repo_with_embedding(&dir);

    let out = mnem(dir.path(), &["embedding", "get", &node_id, &model])
        .assert()
        .success();

    let stdout = String::from_utf8_lossy(&out.get_output().stdout).to_string();
    // Should be space-separated floats, non-empty.
    assert!(!stdout.trim().is_empty(), "stdout must be non-empty");
    // Every token must parse as f32.
    for tok in stdout.split_whitespace() {
        tok.parse::<f32>()
            .unwrap_or_else(|_| panic!("expected f32 token, got: {tok:?}"));
    }
}

#[test]
fn embedding_ls_prints_model_name() {
    let dir = TempDir::new().unwrap();
    let (node_id, model) = seed_repo_with_embedding(&dir);

    let out = mnem(dir.path(), &["embedding", "ls", &node_id])
        .assert()
        .success();

    let stdout = String::from_utf8_lossy(&out.get_output().stdout).to_string();
    assert!(
        stdout.contains(&model),
        "ls output must mention the model: {stdout}"
    );
}

#[test]
fn embedding_get_nonexistent_model_fails() {
    let dir = TempDir::new().unwrap();
    let (node_id, _model) = seed_repo_with_embedding(&dir);

    mnem(
        dir.path(),
        &["embedding", "get", &node_id, "nonexistent-model"],
    )
    .assert()
    .failure();
}
