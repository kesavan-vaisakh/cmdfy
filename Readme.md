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

### Build from Source

You can build the binary for your current system using `go build`:

```bash
git clone https://github.com/kesavan-vaisakh/cmdfy.git
cd cmdfy
go build -o cmdfy app/main.go
# Move to your PATH
mv cmdfy /usr/local/bin/
```

Or use the `Makefile` to cross-compile for all supported platforms:

```bash
make build-all
# Output binaries will be in the bin/ directory
```

-----

## Project Roadmap

This project is being developed in a phased approach. For a detailed breakdown of each phase, its milestones, and a more in-depth architectural overview, please see our dedicated **Phases Document**.

-----

## Contributing

We welcome contributions\! Please refer to the **Phases Document** for information on what we're currently working on and how you can get involved.

-----

## 📄 License

This project is licensed under the MIT License - see the `LICENSE` file for details.