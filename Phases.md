This document provides the detailed project plan, breaking down the development into distinct phases with clear milestones. This is the "blueprint" for your work.

# Project Phases & Milestones

This document outlines the strategic roadmap for the development of **cmdfy**. Each phase represents a significant step towards achieving a fully functional and scalable tool.

-----

## Phase 1: Minimal Viable Product (MVP) - (In Progress)

**Objective**: Build a basic, functional tool that translates simple natural language requests into `ffmpeg` commands using an external LLM.

**Milestones**:

  - [ ] **Core CLI**: Implement the `cmdfy` command with flag parsing (`--config`, `-y`).
  - [ ] **LLM Integration**: Successfully integrate the Gemini API via the Go SDK to receive a generated command string.
  - [ ] **Basic FFmpeg Generation**: The tool can interpret simple requests (e.g., "convert `input.mp4` to `output.avi`") and generate a valid command.
  - [ ] **Configuration Management**: A secure method to store the Gemini API key.
  - [ ] **Execution Logic**: The `-y` flag correctly executes the generated command.

-----

## Phase 2: Localized NLP System

**Objective**: Introduce a local, rule-based NLP engine as an alternative to the LLM, providing offline capability and predictable behavior.

**Milestones**:

  - [ ] **Parser Redesign**: Re-architect the parser to handle a rule-based system instead of relying on the LLM's output.
  - [ ] **Keyword Mapping**: Create a Go-based mapping system that links keywords (e.g., "resize," "fps") to `ffmpeg` flags and filters.
  - [ ] **Engine Selection**: The `--config` flag allows a user to switch between the LLM and the new local NLP engine.
  - [ ] **Feature Parity (MVP)**: The local engine can handle all commands supported by the Phase 1 MVP.
  - [ ] **Robust Error Handling**: Provide clear, user-friendly error messages when the local engine cannot parse a request.

-----

## Phase 3: Command Agnostic Architecture

**Objective**: Expand the tool's capabilities by building a flexible architecture that supports multiple command-line tools.

**Milestones**:

  - [ ] **Tool Abstraction Layer**: Implement a new system (`tool_registry`) that decouples parsing from command generation.
  - [ ] **Generic Request Object**: The parser now generates a tool-agnostic request object (e.g., `CommandRequest`).
  - [ ] **New Tool Rules**: Add rule sets for at least two new tools (e.g., `git`, `imagemagick`).
  - [ ] **Tool Inference**: The system can infer the correct tool based on context (e.g., file extension for `ffmpeg` or `imagemagick`).
  - [ ] **Configurable Rules**: The rule sets for each tool are easily extensible (e.g., by using a separate file or a new Go package).

-----

## Phase 4: Command Merging & Piping

**Objective**: Complete the project by adding support for advanced command-line operations, allowing for multi-step tasks.

**Milestones**:

  - [ ] **Multi-Stage Parser**: The parser can recognize multi-stage commands using keywords like "and," "then," and the `|` symbol.
  - [ ] **Pipeline Data Structure**: A new data structure is created to represent a sequence of commands and their relationships (`Pipeline`).
  - [ ] **Sequential Execution**: The generator correctly assembles merged commands using appropriate separators (e.g., `;`).
  - [ ] **Pipe Operation Support**: The generator can create command strings with pipe operations (`|`) where the output of one command is directed to the input of another.
  - [ ] **Final Polish & Documentation**: Comprehensive documentation on all supported commands, syntax, and features.