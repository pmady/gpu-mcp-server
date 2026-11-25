# Evidence for "From a direct MCP client to goose"

These two archives hold the records from the goose test runs described in the AAIF article "From a direct MCP client to goose". I ran the tests in Google Colab with one Tesla T4 per run on November 20 and November 24, 2026 (UTC dates), using gpu-mcp-server at commit `0579aeadba62e782fc5f7c6831b6da74963dfc87` and goose 1.59.0.

| Archive | Contents |
| --- | --- |
| `aaif-article3-evidence-updated.zip` | November 20. Session exports for cases 01 to 07, the case 04 permission file, the case 08 workload check, and the setup, build and test logs. |
| `aaif-article3-evidence-nov24.zip` | November 24. Session exports and captured MCP stdio traffic for cases 09 and 10, the case 10 workload script and nvidia-smi output, the stdio wrapper, and the setup and build logs. |

Paths in the article are relative to `aaif-article3/` inside each archive. Case numbers follow the folders in the archives.

## Checksums

```
1f987f6be8286196d5624fbb72d8edb4e8664b045497792870304e682ca5e971  aaif-article3-evidence-updated.zip
22c690f6f6fdcdefbb4ccfec92e8f67a1490fd6ed8fefaec3645cb9a9a18c403  aaif-article3-evidence-nov24.zip
```

To verify after downloading both files and `SHA256SUMS`:

```bash
sha256sum -c SHA256SUMS
```

## Notes

- The November 24 archive has a `traces/` directory with copies of the two case captures and two preflight checks of the wrapper.
- `cases/permission-comparison.json` in the November 20 archive is a summary derived from cases 02 and 03.
- Session `20261120_4` is not in the archives and is not part of the case count.
- The archives were checked for credentials before publishing and contain none. They do contain the prompts, the model replies, token counts, the Colab container hostname and the GPU UUIDs of the two Colab runtimes.
