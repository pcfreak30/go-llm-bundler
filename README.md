# go-llm-bundler

A go tool to create a LLM friendly bundle to be able to understand a golang codebase.

Created by an LLM 🙃, with some fixes along the way 😉.

## Installation

You can install go-llm-bundler directly using Go's install command:

```bash
go install github.com/pcfreak30/go-llm-bundler@latest
```

This will download the source, compile it, and install the binary in your `$GOPATH/bin` directory. Make sure this directory is in your system's PATH.

## Usage

After installation, you can run the tool using:

```bash
go-llm-bundler [options]
```

### Command-line Options

- `-dir string`: The root directory of the Go project (default: current directory)
- `-out string`: The output file (default: `<project_name>_bundle.txt`)
- `-meta`: Include metadata such as package structure (default: false)
- `-minify int`: Minification level (1-3, default: 1)
- `-exclude string`: Comma-separated list of directories to exclude (default: "vendor,testdata")

### Examples

1. Bundle the current directory with default options:
   ```bash
   go-llm-bundler
   ```

2. Bundle a specific project with metadata and higher minification:
   ```bash
   go-llm-bundler -dir /path/to/my-project -meta -minify 2
   ```

3. Exclude additional directories and specify an output file:
   ```bash
   go-llm-bundler -dir /path/to/my-project -exclude "vendor,testdata,examples" -out my_custom_bundle.txt
   ```

## Output

The tool generates a file containing:
1. A JSON line with project metadata (dependencies, imports, and optionally structure)
2. The content of each Go file, prefixed with `###FILE:<filename>###`

This format is designed to be easily parsed by AI models while remaining human-readable.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Challenge to the Web

1. Try to come up with better way to represent the data to reduce tokens
2. Take this idea and port to other languages (preferably in go, 1 file binaries rule!)
3. Create a super LLM tool for multiple languages (framework?).

# Notes

1. A previous iteration used pure JSON, but the LLM thought a hybrid text might be better on tokens. Not actually tested but guess we will find out 🙃.
2. AI Code tools in IDE's I believe are doing things like this, but this isolated creates an agnostic way to tell any LLM about your project and ask questions.  I have been tinkering with this on/off for a while now.