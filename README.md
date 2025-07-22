# Demo Simple Werewolf Agent with Tools

The personality of the Werewolf is based on the character sheet provided in `character_sheet.md`.
The `instructions.md` guides the Werewolf to answer questions based on this character sheet.

## Agents

This character is designed to be used with **2 agents**:
- Chat agent: `npcAgent`
- Tools agent (function calling): `toolsAgent`

## Model

```bash
# model
docker model pull ai/qwen2.5:latest
```
> If you use Docker Compose, this will be pulled automatically.

## Start the application

**With Docker Compose**:
```bash
docker compose up --build -d
docker attach $(docker compose ps -q werewolf-agent)
```
> `docker compose down` to stop the application.


**From a container**:
```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal/engines/llama.cpp/v1 \
MODEL_RUNNER_CHAT_MODEL=ai/qwen2.5:latest \
MODEL_RUNNER_TOOLS_MODEL=ai/qwen2.5:latest \
go run main.go
```


**From a local machine**:
```bash
MODEL_RUNNER_BASE_URL=http://localhost:12434/engines/llama.cpp/v1 \
MODEL_RUNNER_CHAT_MODEL=ai/qwen2.5:latest \
MODEL_RUNNER_TOOLS_MODEL=ai/qwen2.5:latest \
go run main.go
```

## Talk to the Werewolf

- What is your name?
- What is your occupation?
- What is your favorite food?  
- What is your background story?
- What is your main quote?

## Test the tools

- what is your health value?
- set your health value to 200
- increase your health by 10
- decrease your health by 5
- what is your intelligence value?
- set your intelligence value to 100
- decrease your intelligence by 5

## Information

### Devcontainer
This project is designed to run in a [Devcontainer](https://code.visualstudio.com/docs/devcontainers/containers) environment.
You can use the provided `.devcontainer` configuration in the `snippets` directory to set up the environment in Visual Studio Code or any compatible IDE.
