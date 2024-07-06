# go-llm-bundler

A go tool to create a LLM friendly bundle to be able to understand a golang codebase.

Created by an LLM 🙃, with some fixes along the way 😉.

---

## Challenge to the Web

1. Try to come up with better way to represent the data to reduce tokens
2. Take this idea and port to other languages (preferably in go, 1 file binaries rule!)
3. Create a super LLM tool for multiple languages (framework?).

# Notes

1. A previous iteration used pure JSON, but the LLM thought a hybrid text might be better on tokens. Not actually tested but guess we will find out 🙃.
2. AI Code tools in IDE's I believe are doing things like this, but this isolated creates an agnostic way to tell any LLM about your project and ask questions.  I have been tinkering with this on/off for a while now.