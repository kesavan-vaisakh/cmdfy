# 🚀 cmdfy: The Command Agnostic CLI

**cmdfy** is a powerful command-line interface that translates natural language requests into executable shell commands. Built with Go, it aims to streamline your workflow by abstracting away complex command-line syntax, allowing you to focus on your task, not the tool.

-----

## ✨ Features

  - **Natural Language to Command Translation**: Generate complex commands by simply describing your intent in plain English.
  - **Configurable Engine**: Choose between a powerful LLM-based engine (like Gemini) for maximum flexibility or a fast, local NLP engine for offline, deterministic behavior.
  - **Tool Agnostic**: Extensible architecture designed to support multiple command-line tools beyond just `ffmpeg`, including `git`, `imagemagick`, and more.
  - **Command Chaining & Piping**: Seamlessly merge and pipe commands to execute complex multi-step operations.

-----

## 🛠️ Installation

**Prerequisites**:

  - Go 1.22 or higher
  - A Gemini API key (optional, for LLM mode)

**From Source**:

1.  Clone the repository:
    ```sh
    git clone https://github.com/your-username/cmdfy.git
    cd cmdfy
    ```
2.  Build the executable:
    ```sh
    go build -o cmdfy ./cmd/cmdfy
    ```
3.  Add the executable to your system's `PATH` for global access.

-----

## 🚀 Usage

### 1\. Configuration

Before first use, configure your preferred command generation engine.

```sh
# To use the LLM-based engine:
./cmdfy --config llm-gemini --api-key YOUR_GEMINI_API_KEY

# To use the local NLP engine:
./cmdfy --config local
```

### 2\. Basic Command Generation

The default behavior is to print the generated command to the terminal for review.

```sh
# Convert a video file
./cmdfy "convert input.mp4 to output.mov with h264 codec"

# Expected Output:
# ffmpeg -i input.mp4 -c:v h264 output.mov
```

### 3\. Direct Execution

Use the `-y` flag to execute the command immediately after it's generated.

```sh
# Be careful! This runs the command directly.
./cmdfy -y "convert video.mov to a 720p version called video_720.mp4"
```

-----

## 🗺️ Project Roadmap

This project is being developed in a phased approach. For a detailed breakdown of each phase, its milestones, and a more in-depth architectural overview, please see our dedicated **Phases Document**.

-----

## 🤝 Contributing

We welcome contributions\! Please refer to the **Phases Document** for information on what we're currently working on and how you can get involved.

-----

## 📄 License

This project is licensed under the MIT License - see the `LICENSE` file for details.