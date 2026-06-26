# LangChain RAG Baseline Benchmark

**Tokenizer:** `cl100k_base` (tiktoken)  
**Embeddings:** `sentence-transformers/all-MiniLM-L6-v2` (sentence-transformers)  
**Vector store:** FAISS (faiss-cpu, in-memory)  
**Retrieval:** similarity search, k=5, top 3 used  
**Chunk sizes tested:** 512, 1024, 2048 tokens, ~10% overlap  

## expressjs/express

| Metric | Value |
|--------|-------|
| Files indexed | **165** |
| Baseline tokens (all files) | **137,978** |
| Chunks (size 512) | 392 |
| Chunks (size 1024) | 254 |
| Chunks (size 2048) | 197 |

### Chunk size: 512 tokens

*392 chunks | embed 6.0s | total build 6.21s | FAISS 1,193 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 4,665 | 137,978 | 96.6% | 29.6x | 0/3 | 0/3 |
| `middleware` | 2,913 | 137,978 | 97.9% | 47.4x | 1/3 | 0/3 |
| `error exception` | 1,342 | 137,978 | 99.0% | 102.8x | 0/3 | 0/3 |
| `request response` | 2,898 | 137,978 | 97.9% | 47.6x | 0/3 | 0/3 |
| `context bind` | 2,615 | 137,978 | 98.1% | 52.8x | 0/3 | 1/3 |
| **Average** | — | — | **97.9%** | **56.0x** | **0.2** | **0.2** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 2,924 | 1,741 | 3/3 | 3/3 | 21.8 |
| `middleware` | 1,851 | 1,062 | 3/3 | 2/3 | 17.8 |
| `error exception` | 1,043 | 299 | 3/3 | 1/3 | 15.2 |
| `request response` | 1,768 | 1,130 | 3/3 | 1/3 | 13.4 |
| `context bind` | 1,921 | 694 | 0/3 | 2/3 | 12.0 |

</details>

### Chunk size: 1024 tokens

*254 chunks | embed 5.59s | total build 5.88s | FAISS 973 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 9,590 | 137,978 | 93.0% | 14.4x | 0/3 | 0/3 |
| `middleware` | 6,703 | 137,978 | 95.1% | 20.6x | 0/3 | 0/3 |
| `error exception` | 1,948 | 137,978 | 98.6% | 70.8x | 0/3 | 0/3 |
| `request response` | 5,640 | 137,978 | 95.9% | 24.5x | 0/3 | 1/3 |
| `context bind` | 6,232 | 137,978 | 95.5% | 22.1x | 0/3 | 0/3 |
| **Average** | — | — | **95.6%** | **30.5x** | **0.0** | **0.2** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 6,007 | 3,583 | 3/3 | 3/3 | 31.6 |
| `middleware` | 4,523 | 2,180 | 3/3 | 3/3 | 21.3 |
| `error exception` | 1,649 | 299 | 3/3 | 1/3 | 21.5 |
| `request response` | 3,522 | 2,118 | 3/3 | 3/3 | 18.2 |
| `context bind` | 3,811 | 2,421 | 0/3 | 2/3 | 20.9 |

</details>

### Chunk size: 2048 tokens

*197 chunks | embed 4.19s | total build 4.43s | FAISS 878 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 16,109 | 137,978 | 88.3% | 8.6x | 0/3 | 0/3 |
| `middleware` | 7,827 | 137,978 | 94.3% | 17.6x | 0/3 | 0/3 |
| `error exception` | 3,067 | 137,978 | 97.8% | 45.0x | 0/3 | 0/3 |
| `request response` | 6,464 | 137,978 | 95.3% | 21.3x | 0/3 | 0/3 |
| `context bind` | 1,816 | 137,978 | 98.7% | 76.0x | 0/3 | 0/3 |
| **Average** | — | — | **94.9%** | **33.7x** | **0.0** | **0.0** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 10,419 | 5,690 | 3/3 | 3/3 | 32.9 |
| `middleware` | 4,743 | 3,084 | 3/3 | 3/3 | 17.1 |
| `error exception` | 2,768 | 299 | 3/3 | 1/3 | 19.1 |
| `request response` | 3,797 | 2,667 | 3/3 | 1/3 | 18.1 |
| `context bind` | 1,116 | 700 | 0/3 | 1/3 | 15.2 |

</details>

---

## fastapi/fastapi

| Metric | Value |
|--------|-------|
| Files indexed | **951** |
| Baseline tokens (all files) | **699,425** |
| Chunks (size 512) | 2,256 |
| Chunks (size 1024) | 1,436 |
| Chunks (size 2048) | 1,114 |

### Chunk size: 512 tokens

*2,256 chunks | embed 46.8s | total build 48.67s | FAISS 7,556 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 2,785 | 699,425 | 99.6% | 251.1x | 0/3 | 2/3 |
| `middleware` | 2,464 | 699,425 | 99.6% | 283.9x | 1/3 | 0/3 |
| `error exception` | 3,402 | 699,425 | 99.5% | 205.6x | 0/3 | 1/3 |
| `request response` | 3,003 | 699,425 | 99.6% | 232.9x | 0/3 | 3/3 |
| `context bind` | 2,598 | 699,425 | 99.6% | 269.2x | 0/3 | 2/3 |
| **Average** | — | — | **99.6%** | **248.5x** | **0.2** | **1.6** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 2,045 | 740 | 3/3 | 2/3 | 16.6 |
| `middleware` | 1,292 | 1,172 | 3/3 | 2/3 | 14.1 |
| `error exception` | 2,325 | 1,077 | 3/3 | 3/3 | 12.3 |
| `request response` | 1,823 | 1,180 | 3/3 | 1/3 | 15.9 |
| `context bind` | 1,446 | 1,152 | 3/3 | 2/3 | 17.3 |

</details>

### Chunk size: 1024 tokens

*1,436 chunks | embed 28.85s | total build 30.39s | FAISS 6,226 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 6,692 | 699,425 | 99.0% | 104.5x | 0/3 | 2/3 |
| `middleware` | 680 | 699,425 | 99.9% | 1028.6x | 0/3 | 0/3 |
| `error exception` | 3,862 | 699,425 | 99.4% | 181.1x | 0/3 | 0/3 |
| `request response` | 4,701 | 699,425 | 99.3% | 148.8x | 0/3 | 2/3 |
| `context bind` | 5,461 | 699,425 | 99.2% | 128.1x | 0/3 | 2/3 |
| **Average** | — | — | **99.4%** | **318.2x** | **0.0** | **1.2** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 4,007 | 2,685 | 3/3 | 3/3 | 17.7 |
| `middleware` | 587 | 93 | 3/3 | 0/3 | 17.5 |
| `error exception` | 2,180 | 1,682 | 3/3 | 3/3 | 20.9 |
| `request response` | 3,485 | 1,216 | 3/3 | 1/3 | 15.4 |
| `context bind` | 3,629 | 1,832 | 3/3 | 3/3 | 19.6 |

</details>

### Chunk size: 2048 tokens

*1,114 chunks | embed 21.87s | total build 23.23s | FAISS 5,693 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 6,988 | 699,425 | 99.0% | 100.1x | 0/3 | 1/3 |
| `middleware` | 4,198 | 699,425 | 99.4% | 166.6x | 0/3 | 0/3 |
| `error exception` | 5,415 | 699,425 | 99.2% | 129.2x | 0/3 | 0/3 |
| `request response` | 4,995 | 699,425 | 99.3% | 140.0x | 0/3 | 2/3 |
| `context bind` | 5,963 | 699,425 | 99.1% | 117.3x | 0/3 | 2/3 |
| **Average** | — | — | **99.2%** | **130.6x** | **0.0** | **1.0** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 4,747 | 2,241 | 3/3 | 2/3 | 21.3 |
| `middleware` | 2,163 | 2,035 | 3/3 | 1/3 | 14.4 |
| `error exception` | 3,816 | 1,599 | 2/3 | 3/3 | 14.2 |
| `request response` | 2,667 | 2,328 | 3/3 | 1/3 | 14.7 |
| `context bind` | 3,651 | 2,312 | 3/3 | 3/3 | 14.8 |

</details>

---

## gin-gonic/gin

| Metric | Value |
|--------|-------|
| Files indexed | **98** |
| Baseline tokens (all files) | **187,018** |
| Chunks (size 512) | 478 |
| Chunks (size 1024) | 271 |
| Chunks (size 2048) | 170 |

### Chunk size: 512 tokens

*478 chunks | embed 11.01s | total build 11.41s | FAISS 1,453 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 3,585 | 187,018 | 98.1% | 52.2x | 0/3 | 0/3 |
| `middleware` | 4,697 | 187,018 | 97.5% | 39.8x | 0/3 | 0/3 |
| `error exception` | 4,807 | 187,018 | 97.4% | 38.9x | 1/3 | 0/3 |
| `request response` | 4,559 | 187,018 | 97.6% | 41.0x | 1/3 | 0/3 |
| `context bind` | 4,110 | 187,018 | 97.8% | 45.5x | 0/3 | 1/3 |
| **Average** | — | — | **97.7%** | **43.5x** | **0.4** | **0.2** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 2,427 | 1,158 | 3/3 | 1/3 | 15.0 |
| `middleware` | 2,972 | 1,725 | 3/3 | 2/3 | 14.4 |
| `error exception` | 2,950 | 1,857 | 3/3 | 2/3 | 14.3 |
| `request response` | 2,736 | 1,823 | 3/3 | 3/3 | 14.5 |
| `context bind` | 2,692 | 1,418 | 3/3 | 2/3 | 16.2 |

</details>

### Chunk size: 1024 tokens

*271 chunks | embed 6.11s | total build 6.42s | FAISS 1,129 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 5,785 | 187,018 | 96.9% | 32.3x | 0/3 | 0/3 |
| `middleware` | 9,102 | 187,018 | 95.1% | 20.5x | 0/3 | 1/3 |
| `error exception` | 8,822 | 187,018 | 95.3% | 21.2x | 1/3 | 0/3 |
| `request response` | 7,188 | 187,018 | 96.2% | 26.0x | 0/3 | 0/3 |
| `context bind` | 6,796 | 187,018 | 96.4% | 27.5x | 0/3 | 0/3 |
| **Average** | — | — | **96.0%** | **25.5x** | **0.2** | **0.2** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 3,764 | 2,021 | 3/3 | 2/3 | 35.6 |
| `middleware` | 5,570 | 3,532 | 3/3 | 3/3 | 18.6 |
| `error exception` | 5,674 | 3,148 | 3/3 | 3/3 | 17.6 |
| `request response` | 4,718 | 2,470 | 3/3 | 3/3 | 15.7 |
| `context bind` | 4,105 | 2,691 | 3/3 | 3/3 | 16.4 |

</details>

### Chunk size: 2048 tokens

*170 chunks | embed 3.91s | total build 4.22s | FAISS 969 KB*

| Query | RAG tokens | Baseline | Reduction | Ratio | Complete/3 | Split/3 |
|-------|----------:|----------:|---------:|------:|:---------:|:-------:|
| `router route handler` | 15,365 | 187,018 | 91.8% | 12.2x | 0/3 | 0/3 |
| `middleware` | 15,191 | 187,018 | 91.9% | 12.3x | 0/3 | 1/3 |
| `error exception` | 13,344 | 187,018 | 92.9% | 14.0x | 0/3 | 0/3 |
| `request response` | 10,296 | 187,018 | 94.5% | 18.2x | 0/3 | 0/3 |
| `context bind` | 10,052 | 187,018 | 94.6% | 18.6x | 0/3 | 0/3 |
| **Average** | — | — | **93.1%** | **15.1x** | **0.0** | **0.2** |

<details><summary>Retrieval quality detail (search + fetch tokens, latency)</summary>

| Query | Search&nbsp;tokens | Fetch&nbsp;tokens | With&nbsp;terms/3 | With&nbsp;logic/3 | Query&nbsp;ms |
|-------|-----------------:|------------------:|:-:|:-:|------:|
| `router route handler` | 9,096 | 6,269 | 3/3 | 3/3 | 30.4 |
| `middleware` | 8,410 | 6,781 | 3/3 | 3/3 | 19.3 |
| `error exception` | 8,276 | 5,068 | 3/3 | 3/3 | 17.3 |
| `request response` | 7,703 | 2,593 | 3/3 | 3/3 | 17.3 |
| `context bind` | 6,134 | 3,918 | 3/3 | 3/3 | 19.7 |

</details>

---

## Combined Comparison

Average RAG tokens per query (mean of 5 queries), compared to jCodemunch.
jCodemunch per-repo figures are from a back-to-back `run_benchmark.py` run on 2026-03-28
against the same index state (express 165 files, fastapi 951 files, gin 98 files).
jCodemunch wins on all three repos.

| Repo | Baseline | RAG-512 | RAG-1024 | RAG-2048 | jCodemunch | Best-RAG-ratio | jCodemunch-ratio | Winner |
|------|--------:|---------:|---------:|---------:|-----------:|--------------:|-----------------:|--------|
| expressjs/express | 137,978 | 2,887 | 6,023 | 7,057 | 924 | 47.8x | 149.3x | jCodemunch (3.1×) |
| fastapi/fastapi | 699,425 | 2,850 | 4,279 | 5,512 | 1,834 | 245.4x | 381.4x | jCodemunch (1.6×) |
| gin-gonic/gin | 187,018 | 4,352 | 7,539 | 12,850 | 1,124 | 43.0x | 166.4x | jCodemunch (3.9×) |

## Infrastructure Overhead

| Repo | Chunk size | Chunks | Embed time | FAISS size |
|------|:----------:|-------:|-----------:|-----------:|
| expressjs/express | 512 | 392 | 6.0s | 1,193 KB |
| expressjs/express | 1024 | 254 | 5.59s | 973 KB |
| expressjs/express | 2048 | 197 | 4.19s | 878 KB |
| fastapi/fastapi | 512 | 2,256 | 46.8s | 7,556 KB |
| fastapi/fastapi | 1024 | 1,436 | 28.85s | 6,226 KB |
| fastapi/fastapi | 2048 | 1,114 | 21.87s | 5,693 KB |
| gin-gonic/gin | 512 | 478 | 11.01s | 1,453 KB |
| gin-gonic/gin | 1024 | 271 | 6.11s | 1,129 KB |
| gin-gonic/gin | 2048 | 170 | 3.91s | 969 KB |

## Chunk Integrity Summary

Percentage of retrieved top-3 chunks that are complete code units vs. split mid-function.
(Heuristic: complete = starts with def/class/function/func and braces balance; split = brace imbalance > 2 or ends mid-indented-block.)

| Repo | Chunk size | Complete % | Split % |
|------|:----------:|-----------:|--------:|
| expressjs/express | 512 | 7% | 7% |
| expressjs/express | 1024 | 0% | 7% |
| expressjs/express | 2048 | 0% | 0% |
| fastapi/fastapi | 512 | 7% | 53% |
| fastapi/fastapi | 1024 | 0% | 40% |
| fastapi/fastapi | 2048 | 0% | 33% |
| gin-gonic/gin | 512 | 13% | 7% |
| gin-gonic/gin | 1024 | 7% | 7% |
| gin-gonic/gin | 2048 | 0% | 7% |
