# arXiv preprint

`main.tex` is a self-contained LaTeX preprint (article class, inline
bibliography) describing gpu-mcp-server.

## Build

A single pass is enough (the bibliography is inline):

- Overleaf: upload `main.tex`, compile. (Easiest, no local install.)
- Tectonic: `tectonic main.tex`
- TeX Live: `pdflatex main.tex`

## Submit to arXiv

1. Category: `cs.DC` (Distributed, Parallel, and Cluster Computing), optionally
   cross-list `cs.AI`.
2. Upload the LaTeX source (`main.tex`); arXiv compiles it server-side.
3. First-time submitters may need an endorsement.
4. Keep author name and ORCID consistent with `CITATION.cff`.

The preprint cites the software DOI (`10.5281/zenodo.22866670`).
