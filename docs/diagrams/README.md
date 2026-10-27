# Diagrams

Mermaid sources for the figures in `docs/Proposal-ADR.docx` and `docs/Proposal.md`.
Edit the `.mmd` file, then re-render the PNG next to it:

```bash
cd docs/diagrams
for f in *.mmd; do npx -p @mermaid-js/mermaid-cli mmdc -i "$f" -o "${f%.mmd}.png" -b white -s 2; done
```

After re-rendering, replace the matching image in the `.docx` as well.
