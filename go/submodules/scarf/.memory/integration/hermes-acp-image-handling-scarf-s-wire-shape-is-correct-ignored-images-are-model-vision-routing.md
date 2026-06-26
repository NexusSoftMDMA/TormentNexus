---
title: Hermes ACP image handling: Scarf's wire shape is correct; ignored images are model-vision routing
type: note
permalink: scarf/integration/hermes-acp-image-handling-scarf-s-wire-shape-is-correct-ignored-images-are-model-vision-routing
tags:
- acp
- images
- hermes
- vision
- gh113
- wire-format
---

Diagnosed gh#113 ("images attached to messages ignored") against the real Hermes v0.16.0 source at `~/.hermes/hermes-agent/` (build 2026.6.5). Scarf's image path is correct end-to-end; "ignored" is a Hermes model-vision routing decision, NOT a Scarf wire bug.

## Observations

- [wire-format] Scarf sends image attachments as ACP `session/prompt` content blocks: `{"type":"image","data":"<raw base64>","mimeType":"image/jpeg"}`. `ImageEncoder` (ScarfCore) produces RAW base64 via `Data.base64EncodedString()` — NO `data:` prefix (Hermes adds the prefix). `ACPClient.sendPrompt` builds `[{type:text}, {type:image,...}]`. This is the correct shape. #wire-format
- [hermes-parses-it] Hermes v0.16 `acp_adapter/server.py` parses that into `ImageContentBlock(type="image", data=…, mimeType=…)` and `_content_blocks_to_openai_user_content` converts it to OpenAI `{"type":"image_url","image_url":{"url":"data:<mime>;base64,<data>"}}`. Verified by Hermes's own passing test `tests/acp_adapter/test_acp_images.py`. `initialize` advertises `prompt_capabilities.image == True`. #hermes
- [text-vs-multimodal] In `prompt()`, text-only blocks collapse to a plain string (legacy/slash-command path); any image block makes `text_only_prompt=False`, so the multimodal list flows to `agent.run_conversation(user_message=user_content)`. Only the clean text is persisted to history (`persist_user_message`), not the base64. #hermes
- [root-cause] The deciding factor is `agent/image_routing.py::decide_image_input_mode(provider, model, cfg)`. Default `agent.image_input_mode: auto`: (1) if `auxiliary.vision.provider` is set → TEXT pipeline; (2) else if active model `supports_vision=True` (models.dev metadata) → NATIVE (model sees pixels); (3) else → TEXT pipeline (`vision_analyze` → lossy text summary, "model never sees the pixels"). A non-vision/unrecognized model + no vision backend ⇒ image effectively dropped = "ignored." #vision
- [user-fixes] Use a vision model; or `agent.image_input_mode: native` in config.yaml (force pixels); or configure `auxiliary.vision.provider` for the text pipeline. #vision
- [scarf-followup] Potential Scarf UX: warn in the composer when attaching an image to a session whose model isn't vision-capable. Needs a per-model vision signal — `HermesCapabilities` is version-scoped, not per-model, so this needs a models.dev lookup or heuristic. Tracked as t-31img. #scarf
- [gotcha] When debugging "X not reaching Hermes," check the real local install source first (`~/.hermes/hermes-agent/`, a git checkout for source installs) + its tests — faster and more authoritative than guessing the wire contract. #debugging

## Relations

- relates_to [[Model Presets Feature]]
- relates_to [[Scarf Architecture Rules]]
