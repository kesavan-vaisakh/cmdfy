# cmdfy: Turn Natural Language into Shell Commands

**cmdfy** is a command-line tool that translates natural language requests into executable shell commands. It leverages Large Language Models (LLMs) like Gemini, OpenAI, and local options via Ollama to generate accurate commands tailored to your operating system's context.

## Installation

### Install via Go Install

If you have Go installed, the easiest way to install **cmdfy** is using `go install`:

```bash
go install github.com/kesavan-vaisakh/cmdfy/app@latest
mv $(go env GOPATH)/bin/app $(go env GOPATH)/bin/cmdfy
```
*Note: The install path is being streamlined.*

### Build from Source

You can build the binary for your current system using `go build`:

```bash
git clone https://github.com/kesavan-vaisakh/cmdfy.git
cd cmdfy
go build -o cmdfy app/main.go
# Move to your PATH
mv cmdfy /usr/local/bin/
```

Or use the `Makefile` to make things easier.

**Build for current OS:**
```bash
make build
```

**Cross-compile for all platforms:**
```bash
make build-all
# Output binaries will be in the bin/ directory
```

## Usage

### 1. Configuration

Before using **cmdfy**, you need to configure your preferred LLM provider.

**Gemini:**
```bash
./cmdfy --config llm-gemini --api-key YOUR_GEMINI_API_KEY
```

**Local (Ollama):**
```bash
# To use the local NLP engine:
./cmdfy --config local
```

### 2. Basic Command Generation

The default behavior is to print the generated command to the terminal for review.

```sh
# Convert a video file
./cmdfy "convert input.mp4 to output.mov with h264 codec"

# Expected Output:
# ffmpeg -i input.mp4 -c:v h264 output.mov
```

### 3. Direct Execution

Use the `-y` flag to execute the command immediately after it's generated.

```sh
# Be careful! This runs the command directly.
./cmdfy -y "convert video.mov to a 720p version called video_720.mp4"
```

### 4. Benchmarking Mode (`--compare`)

Unsure which AI model is best? Run a benchmark!

```bash
./cmdfy "find all large files over 100MB" --compare
```
This opens an interactive TUI (Terminal User Interface) that runs your query against all configured providers (e.g., Ollama, Gemini, OpenAI) and shows the results side-by-side. 

**Bonus:** When you pick a winner, `cmdfy` **memorizes** it to its Local Brain (`~/.cmdfy/brain.jsonl`), teaching your local models to be smarter next time.

### 5. Error Fixing ("The Aha Moment")

If a command fails, pipe the error output to `cmdfy` to fix it automatically.

```bash
# Example: Forgot to install a tool
some_command_that_fails 2>&1 | cmdfy "fix this"
```
`cmdfy` reads the error from stdin, analyzes it, and suggests a corrected command.

## Project Roadmap

This project is being developed in a phased approach. For a detailed breakdown of each phase, its milestones, and a more in-depth architectural overview, please see the dedicated [Phases Document](Phases.md).

## Contributing

Contributions are welcome! Please refer to the [Phases Document](Phases.md) for information on current progress and how to get involved.

## License

This project is licensed under the MIT License - see the `LICENSE` file for details.