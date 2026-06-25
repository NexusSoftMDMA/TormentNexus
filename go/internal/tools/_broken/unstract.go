package tools

import (
	"context"
)

// HandleGetAdapterExamples returns usage examples for adapters
func HandleGetAdapterExamples(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	provider, _ := getString(args, "provider")
	exampleType, _ := getString(args, "example_type")

	examples := map[string]string{
		"basic": `from unstract.sdk1.adapters.adapterkit import Adapterkit

kit = Adapterkit()
adapters = kit.get_adapters_list("LLM")

for adapter in adapters:
    print(f"{adapter['name']}: {adapter['provider']}")`,

		"initialize": `from unstract.sdk1.adapters.adapterkit import Adapterkit

kit = Adapterkit()
adapter = kit.get_adapter("openai|502ecf49-e47c-445c-9907-6d4b90c5cd17")

if adapter:
    metadata = adapter.get_metadata()
    print(f"Adapter: {metadata['name']}")`,

		"validate": `from unstract.sdk1.adapters.base1 import OpenAILLMParameters

config = {
    "model": "gpt-4",
    "api_key": "sk-...",
    "max_tokens": 1000
}

validated = OpenAILLMParameters.validate(config)
print(f"Validated config: {validated}")`,

		"custom_adapter": `from unstract.sdk1.adapters.base1 import BaseChatCompletionParameters, BaseAdapter
from unstract.sdk1.adapters.enums import AdapterTypes

class MyCustomLLMParameters(BaseChatCompletionParameters):
    api_key: str`,
	}

	_ = provider

	if exampleType != "" {
		if val, exists := examples[exampleType]; exists {
			return ok(val)
		}
	}

	return ok(examples["basic"])
}