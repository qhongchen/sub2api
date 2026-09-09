import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import ts from "typescript";

const root = new URL("./", import.meta.url);
const printer = ts.createPrinter({ removeComments: true });

function readSyntax(relativePath) {
  const url = new URL(relativePath, root);
  const source = ts.createSourceFile(
    url.pathname,
    readFileSync(url, "utf8"),
    ts.ScriptTarget.Latest,
    true,
    ts.ScriptKind.TS,
  );
  assert.equal(source.parseDiagnostics.length, 0, `${relativePath}: invalid syntax`);
  return printer.printFile(source);
}

for (const name of ["groupModelAllowlist", "modelAllowlistCandidates"]) {
  const path = `src/views/admin/${name}.ts`;
  assert.equal(
    readSyntax(path),
    readSyntax(`../frontend/${path}`),
    `${name}: implementation differs from upstream`,
  );
  console.log(`PASS ${name}`);
}
