---
name: pdf-processor
description: Process and extract text from PDF files
allowed_tools: [Read, Bash, Write]
---

## PDF Processing

When asked to process a PDF file:

1. Use `pdftotext` to extract text content
2. If tables are present, use `tabula` for extraction
3. For forms, use `pdfinfo` first to understand the structure
4. Save extracted content to a .txt or .csv file as appropriate

## Commands Reference

```bash
# Extract text
pdftotext -layout input.pdf output.txt

# Extract tables
tabula -o output.csv input.pdf

# Get PDF info
pdfinfo input.pdf
```
