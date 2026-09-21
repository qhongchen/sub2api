import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import ts from 'typescript'
import { parse } from 'vue/compiler-sfc'

const root = new URL('./', import.meta.url)
const printer = ts.createPrinter({ removeComments: true })

function readSource(path) {
  let text = readFileSync(new URL(path, root), 'utf8')
  if (path.endsWith('.vue')) {
    const { descriptor, errors } = parse(text, { filename: path })
    assert.equal(errors.length, 0, `${path}: invalid Vue syntax`)
    text = [descriptor.script?.content, descriptor.scriptSetup?.content].filter(Boolean).join('\n')
  }
  const source = ts.createSourceFile(path, text, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
  assert.equal(source.parseDiagnostics.length, 0, `${path}: invalid TypeScript syntax`)
  return source
}

function readSyntax(path, name) {
  const source = readSource(path)
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
  ['src/api/admin/ops.ts', 'OpsRuntimeLogConfig'],
  ['src/api/admin/proxies.ts', 'assertProxyArray'],
  ['src/api/admin/proxies.ts', 'list'],
  ['src/api/admin/proxies.ts', 'getAll'],
  ['src/api/admin/proxies.ts', 'getAllWithCount'],
  ['src/api/admin/accounts.ts', 'getGrokMediaEligibility'],
  ['src/api/admin/accounts.ts', 'updateGrokMediaEligibility'],
  ['src/types/index.ts', 'GrokMediaEligibilityMode'],
  ['src/types/index.ts', 'GrokMediaEligibilityState'],
  ['src/utils/featureFlags.ts', 'isChannelMonitorUserRankingHidden'],
  ['src/features/channel-monitor-v2/monitorFormat.ts', 'healthScoreClass'],
  ['src/features/channel-monitor-v2/monitorFormat.ts', 'isTtftUnavailable'],
  ['src/features/channel-monitor-v2/monitorFormat.ts', 'ttftDisplayState'],
  ['src/features/channel-monitor-v2/MetricCell.vue', 'missingValue'],
  ['src/features/channel-monitor-v2/MetricCell.vue', 'resolvedState'],
  ['src/features/channel-monitor-v2/MetricCell.vue', 'stateClass'],
  ['src/features/channel-monitor-v2/MetricCell.vue', 'dotClass'],
  ['src/views/user/ChannelStatusV2View.vue', 'showUserRanking'],
  ['src/views/user/ChannelStatusV2View.vue', 'tabs'],
  ['src/views/user/ChannelStatusV2View.vue', 'activeTab'],
  ['src/views/user/ChannelStatusV2View.vue', 'parseTab'],
  ['src/views/user/ChannelStatusV2View.vue', 'loadTab'],
  ['src/views/user/ChannelStatusV2View.vue', 'ttftCellState'],
  ['src/components/account/EditAccountModal.vue', 'modeFromGrokMediaExtra'],
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

const opsPath = 'src/views/admin/ops/components/OpsSystemLogTable.vue'
assert.match(readSyntax(opsPath, 'runtimeConfig'), /persist_access_logs: false/)
for (const name of ['loadRuntimeConfig', 'saveRuntimeConfig', 'resetRuntimeConfig']) {
  assert.match(readSyntax(opsPath, name), /runtimeConfig\.persist_access_logs = (?:cfg|saved)\.persist_access_logs/)
}
assert.match(readSyntax(opsPath, 'saveRuntimeConfig'), /updateRuntimeLogConfig\(\{ \.\.\.runtimeConfig \}\)/)
console.log('PASS Ops access log defaults, loading, saving and reset')

for (const name of ['SystemSettings', 'UpdateSettingsRequest']) {
  assert.match(readSyntax('src/api/admin/settings.ts', name), /channel_monitor_hide_user_ranking\??: boolean/)
}
assert.match(readSyntax('src/types/index.ts', 'PublicSettings'), /channel_monitor_hide_user_ranking\??: boolean/)
const settingsPath = 'src/views/admin/SettingsView.vue'
assert.match(readSyntax(settingsPath, 'form'), /channel_monitor_hide_user_ranking: false/)
assert.match(readSyntax(settingsPath, 'loadSettings'), /form\.channel_monitor_hide_user_ranking = Boolean\(settings\.channel_monitor_hide_user_ranking\)/)
assert.match(readSyntax(settingsPath, 'saveSettings'), /channel_monitor_hide_user_ranking: Boolean\(form\.channel_monitor_hide_user_ranking\)/)
assert.match(readSyntax('src/views/user/ChannelStatusV2View.vue'), /watch\(showUserRanking, \(allowed\) => \{\s+if \(!allowed && activeTab\.value === 'users'\) \{\s+activeTab\.value = 'models'/)
console.log('PASS ranking settings contract and hidden-tab fallback')

const editPath = 'src/components/account/EditAccountModal.vue'
const loadEligibility = readSyntax(editPath, 'loadGrokMediaEligibility')
assert.match(loadEligibility, /requestVersion !== grokMediaEligibilityRequestVersion/)
assert.match(loadEligibility, /requestVersion === grokMediaEligibilityRequestVersion/)
assert.match(readSyntax(editPath, 'handleClose'), /grokMediaEligibilityRequestVersion\+\+/)
assert.match(readSyntax(editPath, 'syncFormFromAccount'), /loadGrokMediaEligibility\(newAccount\.id\)/)
const persistEligibility = readSyntax(editPath, 'persistGrokMediaEligibility')
assert.match(persistEligibility, /updateGrokMediaEligibility\(accountID, mode\)/)
assert.match(persistEligibility, /getGrokMediaEligibility\(accountID\)/)
assert.match(persistEligibility, /grokMediaEligibility\.partialSave/)
assert.match(persistEligibility, /delete nextExtra\.grok_media_eligible/)
assert.match(readSyntax(editPath, 'submitUpdateAccount'), /const mediaEligibilityMode =[\s\S]+await adminAPI\.accounts\.update[\s\S]+persistGrokMediaEligibility\(accountID, updatedAccount, mediaEligibilityMode\)/)
for (const name of ['getGrokMediaEligibility', 'updateGrokMediaEligibility']) {
  assert.ok(readSyntax('src/api/admin/accounts.ts', 'accountsAPI').includes(name), `accountsAPI: missing ${name}`)
}
assert.match(readSyntax('src/composables/useModelWhitelist.ts', 'openaiModels'), /'gpt-image-2\.5-flare', 'gpt-image-2\.5-sunburst'/)
console.log('PASS Grok request isolation, partial saves and Image 2.5 presets')

for (const [path, binding] of [
  [opsPath, 'v-model="runtimeConfig.persist_access_logs"'],
  [settingsPath, 'v-model="form.channel_monitor_hide_user_ranking"'],
  [editPath, ':disabled="grokMediaEligibilityLoading || submitting"'],
  ['src/features/channel-monitor-v2/MetricCell.vue', 'v-if="resolvedState"'],
  ['src/views/user/ChannelStatusV2View.vue', ':state="ttftCellState(snapshot.health.ttft, snapshot.metrics.ttft)"'],
]) {
  const { descriptor, errors } = parse(readFileSync(new URL(path, root), 'utf8'), { filename: path })
  assert.equal(errors.length, 0, `${path}: invalid Vue syntax`)
  assert.ok(descriptor.template?.content.includes(binding), `${path}: missing ${binding}`)
}

function translations(path) {
  const source = readSource(path)
  const entries = new Map()
  function visit(node, prefix = '') {
    assert.ok(node && ts.isObjectLiteralExpression(node), `${path}: expected translation object`)
    for (const property of node.properties) {
      assert.ok(ts.isPropertyAssignment(property), `${path}: expected translation property`)
      const key = prefix + property.name.text
      if (ts.isObjectLiteralExpression(property.initializer)) {
        visit(property.initializer, `${key}.`)
      } else {
        assert.ok(ts.isStringLiteralLike(property.initializer), `${path}: expected string for ${key}`)
        assert.ok(!entries.has(key), `${path}: duplicate ${key}`)
        entries.set(key, property.initializer.text)
      }
    }
  }
  visit(source.statements.find(ts.isExportAssignment)?.expression)
  return entries
}

for (const locale of ['en', 'zh']) {
  const local = translations(`src/i18n/locales/${locale}.ts`)
  for (const [section, keys] of [
    ['accounts', ['accounts.grokMediaEligibility.']],
    ['ops', ['ops.systemLogs.persistAccessLogs', 'ops.systemLogs.persistAccessLogsHint', 'ops.systemLogs.retentionDaysHint']],
    ['settings', ['settings.features.channelMonitor.hideUserRanking', 'settings.features.channelMonitor.hideUserRankingHint']],
  ]) {
    const upstream = translations(`../frontend/src/i18n/locales/${locale}/admin/${section}.ts`)
    for (const key of keys) {
      const required = key.endsWith('.') ? [...upstream.keys()].filter(candidate => candidate.startsWith(key)) : [key]
      assert.ok(required.length > 0, `${locale}: missing upstream ${key}`)
      for (const candidate of required) {
        assert.ok(upstream.get(candidate)?.trim(), `${locale}: missing upstream ${candidate}`)
        assert.ok(local.get(`admin.${candidate}`)?.trim(), `${locale}: missing admin.${candidate}`)
      }
    }
  }
}
console.log('PASS Vue bindings and English/Chinese migration translations')
