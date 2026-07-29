export type FingerprintSignalType = "header_exact" | "header_prefix" | "body_path";

export interface FingerprintSignalRow {
  type: FingerprintSignalType;
  match: string;
  required: boolean;
}

const VALID_TYPES: FingerprintSignalType[] = [
  "header_exact",
  "header_prefix",
  "body_path",
];

export function parseFingerprintSignalsToRows(raw: string): FingerprintSignalRow[] {
  if (!raw || !raw.trim()) return [];
  try {
    const entries = JSON.parse(raw);
    if (!Array.isArray(entries)) return [];
    return entries.map((entry) => ({
      type: VALID_TYPES.includes(entry?.type) ? entry.type : "header_exact",
      match: Array.isArray(entry?.match)
        ? entry.match
            .filter((value: unknown) => typeof value === "string")
            .join(" / ")
        : "",
      required: entry?.required === true,
    }));
  } catch {
    return [];
  }
}

export function serializeFingerprintRowsToJSON(rows: FingerprintSignalRow[]): string {
  const entries = rows
    .map((row) => ({
      type: row.type,
      match: row.match
        .split("/")
        .map((value) => value.trim())
        .filter((value) => value.length > 0),
      required: row.required === true,
    }))
    .filter((entry) => entry.match.length > 0);
  return JSON.stringify(entries);
}

export function defaultFingerprintSignalRows(): FingerprintSignalRow[] {
  return [
    { type: "header_prefix", match: "x-codex-", required: true },
    { type: "header_exact", match: "session-id / session_id", required: false },
    { type: "header_exact", match: "thread-id / thread_id", required: false },
    {
      type: "body_path",
      match:
        "client_metadata.x-codex-window-id / client_metadata.x-codex-installation-id",
      required: false,
    },
  ];
}
