import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import { parse } from 'vue/compiler-sfc'

const root = new URL('./', import.meta.url)
const printer = ts.createPrinter({ removeComments: true })

function readSyntax(path, name) {
  let text = readFileSync(new URL(path, root), 'utf8')
  if (path.endsWith('.vue')) {
    const { descriptor, errors } = parse(text, { filename: path })
    assert.equal(errors.length, 0, `${path}: invalid Vue syntax`)
    text = [descriptor.script?.content, descriptor.scriptSetup?.content].filter(Boolean).join('\n')
  }
  const source = ts.createSourceFile(path, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  assert.equal(source.parseDiagnostics.length, 0, `${path}: invalid TypeScript syntax`)
  if (!name) return printer.printFile(source).trim()

  let declaration
  function visit(node) {
    if ((ts.isFunctionDeclaration(node) || ts.isVariableDeclaration(node) ||
      ts.isInterfaceDeclaration(node) || ts.isTypeAliasDeclaration(node)) && node.name?.getText(source) === name) {
      declaration = node
    }
    ts.forEachChild(node, visit)
  }
  visit(source)
  assert.ok(declaration, `${path}: missing ${name}`)
  return printer.printNode(ts.EmitHint.Unspecified, declaration, source).trim()
}

// Only parse source and compare declarations; do not compile or execute application code.
for (const [path, name] of [
  ['src/views/admin/groupsReasoningEffort.ts'],
  ['src/components/account/accountExpiry.ts'],
  ['src/utils/pricing.ts', 'resolveIntervalPrices'],
  ['src/components/admin/channel/types.ts', 'isValidPositiveMultiplier'],
  ['src/components/admin/channel/types.ts', 'apiIntervalsToForm'],
  ['src/components/admin/channel/types.ts', 'formIntervalsToAPI'],
  ['src/components/account/AccountUsageCell.vue', 'openAISevenDayEstimatedTotalCost'],
  ['src/types/index.ts', 'AccountListItem'],
  ['src/types/index.ts', 'CodexModelsManifestConfig'],
]) {
  assert.equal(readSyntax(path, name), readSyntax(`../frontend/${path}`, name), `${path}: ${name ?? 'source'} differs from upstream`)
  console.log(`PASS upstream parity: ${name ?? path}`)
}

for (const name of ['AdminGroup', 'CreateGroupRequest', 'UpdateGroupRequest']) {
  assert.match(readSyntax('src/types/index.ts', name), /codex_models_manifest_config\??: CodexModelsManifestConfig/)
}

for (const name of ['handleEdit', 'handleTest', 'handleViewStats']) {
  assert.match(readSyntax('src/views/admin/AccountsView.vue', name), /await loadAccountDetails\(a\)/)
}

for (const name of ['handleSubmit', 'createAccountAndFinish', 'handleGrokValidateRT', 'handleGrokImportSSO',
  'handleOpenAIExchange', 'handleOpenAIImportCodexSession', 'handleOpenAIImportCodexPAT',
  'handleOpenAIBatchRT', 'handleAntigravityValidateRT', 'handleCookieAuth']) {
  assert.match(readSyntax('src/components/account/CreateAccountModal.vue', name), /withUpstreamRequestIdHeader\(/)
}
console.log('PASS group DTOs, account detail loading and upstream ID creation paths')
