#!/bin/bash
MODEL_RUNNER_BASE_URL=http://localhost:12434/engines/llama.cpp/v1 \
MODEL_RUNNER_CHAT_MODEL=ai/qwen2.5:1.5B-F16 \
MODEL_RUNNER_TOOLS_MODEL=hf.co/menlo/lucy-gguf:q8_0 \
go run main.go

# MODEL_RUNNER_TOOLS_MODEL=hf.co/salesforce/xlam-2-3b-fc-r-gguf:q3_k_l \
# MODEL_RUNNER_TOOLS_MODEL=hf.co/salesforce/llama-xlam-2-8b-fc-r-gguf:q4_k_m \
# MODEL_RUNNER_TOOLS_MODEL=hf.co/salesforce/xlam-2-3b-fc-r-gguf:q3_k_l \
# MODEL_RUNNER_TOOLS_MODEL=ai/qwen2.5:latest \