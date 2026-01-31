This document provides the detailed project plan, breaking down the development into distinct phases with clear milestones. This is the "blueprint" for your work.

# Project Phases & Milestones

This document outlines the strategic roadmap for the development of **cmdfy**. Each phase represents a significant step towards achieving a fully functional and scalable tool.

-----

## Phase 1: Minimal Viable Product (MVP) - (Completed)

**Objective**: Build a basic, functional tool that translates simple natural language requests into shell commands using an external LLM.

**Milestones**:

  - [x] **Core CLI**: Implement the `cmdfy` command with flag parsing (`--config`, `-y`).
  - [x] **LLM Integration**: Successfully integrate the Gemini API via the Go SDK to receive a generated command string.
  - [x] **Basic Command Generation**: The tool can interpret simple requests (e.g., "convert `input.mp4` to `output.avi`") and generate a valid command. (General purpose)
  - [x] **Configuration Management**: A secure method to store the Gemini API key.
  - [x] **Execution Logic**: The `-y` flag correctly executes the generated command.

-----

## Phase 2: Local LLM Support (Ollama) - (Completed)

**Objective**: Introduce local LLM support via Ollama, providing offline capability, privacy, and cost savings.

**Milestones**:

  - [x] **Ollama Provider**: Implement the `ollama` provider in `pkg/llm/ollama` using the Ollama API.
  - [x] **Configuration Update**: Update the CLI to support setting the usage of a local model and custom base URL (default: `http://localhost:11434`).
  - [x] **Model Selection**: Allow the user to specify which local model to use (e.g., `llama3`, `mistral`).
  - [x] **Verification**: Ensure seamless switching between Cloud (Gemini/OpenAI) and Local (Ollama) execution.

-----

## Phase 3: Structured Command Architecture - (Completed)

**Objective**: Expand the tool's capabilities by building a flexible architecture that supports multiple command-line tools and structured output.

**Milestones**:

  - [x] **Structured Models**: Define `GeneratedCommand` struct to hold tool, args, explanation, and danger status.
  - [x] **LLM Interface Update**: Update `Provider` interface to return structured data instead of raw strings.
  - [x] **JSON Prompting**: Update prompts to request JSON output from all providers (Gemini, OpenAI, Ollama).
  - [x] **CLI Updates**: Update `cmdfy` CLI to parse and display structured output nicely.

-----

## Phase 4: Command Merging & Piping - (Completed)

**Objective**: Complete the project by adding support for advanced command-line operations, allowing for multi-step tasks.

**Milestones**:

  - [x] **Multi-Stage Parser**: The parser (LLM) recognizes multi-stage commands and breaks them into steps.
  - [x] **Pipeline Data Structure**: Implemented `CommandResult` and `CommandStep` to represent sequences and operators.
  - [x] **Sequential Execution**: The CLI correctly assembles merged commands using operators (`&&`, `;`).
  - [x] **Pipe Operation Support**: The system supports pipe operations (`|`) for data flow between commands.


## Phase 5: Multi-Provider SDK & Smart Context - (Completed)
**Objective** Implement Provider interface; add OpenAI, Claude, and Ollama support with context awareness.

**Milestones**:
  - [x] **Expandable Interface**: Keep the interface open for future providers (Claude added).
  - [x] **Context aware**: Added  `--directory`local file availability to LLM context to prevent hallucinations.
  - [x] **Context from clipboard**: Allow the user to copy using `--clipboard` flag.
  - [x] **Anthropic Provider**: Implemented Claude API support.



## Phase 6,Benchmarking Mode
**Objective** Launch --compare flag with a side-by-side TUI for voting on results.

**Milestones**:
  - [x] **Config update**: update the config to allow multiple providers to be used in benchmarking mode.
  - [x] **Add Metrics**: need to add metrics like `token usage`, `time taken`, `cost` etc. 
  - [x] **TUI Implementation**: Implement a side-by-side TUI for voting on results.


## Phase 7: Local "Brain" - (Completed)
**Objective** Allow cmdfy to learn from your picks to improve the local NLP accuracy.

**Milestones**:
  - [x] **The "Aha" Moment**: Piping stderr directly into the CLI.
  - [x] **Local Brain Implementation**: Implement a local brain (`~/.cmdfy/brain.jsonl`) to learn from user choices and successful executions, enabling few-shot learning for all providers.

## Phase Futures will plan after completion of Phase 6
**Milestones**:

  - [ ] **Future Planning**: Plan for future features and improvements.
