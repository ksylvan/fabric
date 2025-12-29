# Code Analysis Report

**Target Directory:** `/Users/kayvan/src/fabric`

## Summary

- **Total Files:** 270
- **Total Lines of Code:** 30322
- **Large Files (> 500 lines):** 8

## Files by Extension

| Extension | Count | Percentage |
|-----------|-------|------------|
| .go | 164 | 60.7% |
| .ts | 79 | 29.3% |
| .js | 27 | 10.0% |

## Potentially Large Files

The following files exceed 500 lines and may benefit from refactoring:

- `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/generator.go` - 805 lines
- `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/processing.go` - 530 lines
- `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/git/walker.go` - 574 lines
- `/Users/kayvan/src/fabric/docs/docs.go` - 536 lines
- `/Users/kayvan/src/fabric/internal/cli/flags.go` - 557 lines
- `/Users/kayvan/src/fabric/internal/core/plugin_registry.go` - 577 lines
- `/Users/kayvan/src/fabric/internal/plugins/ai/gemini/gemini.go` - 547 lines
- `/Users/kayvan/src/fabric/internal/tools/youtube/youtube.go` - 840 lines

## File Details

| File | Extension | Lines | Large |
|------|-----------|-------|-------|
| `/Users/kayvan/src/fabric/cmd/code_analysis/internal/analyzer.go` | .go | 176 | No |
| `/Users/kayvan/src/fabric/cmd/code_analysis/main.go` | .go | 53 | No |
| `/Users/kayvan/src/fabric/cmd/code_helper/code.go` | .go | 181 | No |
| `/Users/kayvan/src/fabric/cmd/code_helper/main.go` | .go | 65 | No |
| `/Users/kayvan/src/fabric/cmd/fabric/main.go` | .go | 18 | No |
| `/Users/kayvan/src/fabric/cmd/fabric/version.go` | .go | 3 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/cache/cache.go` | .go | 476 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/generator.go` | .go | 805 | Yes |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/generator_test.go` | .go | 115 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/merge_detection_test.go` | .go | 82 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/processing.go` | .go | 530 | Yes |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/processing_test.go` | .go | 262 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/changelog/summarize.go` | .go | 79 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/config/config.go` | .go | 21 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/git/types.go` | .go | 26 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/git/walker.go` | .go | 574 | Yes |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/github/client.go` | .go | 431 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/github/email_test.go` | .go | 59 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/github/types.go` | .go | 68 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/internal/release.go` | .go | 149 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/main.go` | .go | 117 | No |
| `/Users/kayvan/src/fabric/cmd/generate_changelog/util/token.go` | .go | 31 | No |
| `/Users/kayvan/src/fabric/cmd/to_pdf/main.go` | .go | 220 | No |
| `/Users/kayvan/src/fabric/docs/docs.go` | .go | 536 | Yes |
| `/Users/kayvan/src/fabric/internal/chat/chat.go` | .go | 132 | No |
| `/Users/kayvan/src/fabric/internal/cli/chat.go` | .go | 200 | No |
| `/Users/kayvan/src/fabric/internal/cli/chat_test.go` | .go | 166 | No |
| `/Users/kayvan/src/fabric/internal/cli/cli.go` | .go | 189 | No |
| `/Users/kayvan/src/fabric/internal/cli/cli_test.go` | .go | 21 | No |
| `/Users/kayvan/src/fabric/internal/cli/configuration.go` | .go | 28 | No |
| `/Users/kayvan/src/fabric/internal/cli/extensions.go` | .go | 26 | No |
| `/Users/kayvan/src/fabric/internal/cli/flags.go` | .go | 557 | Yes |
| `/Users/kayvan/src/fabric/internal/cli/flags_test.go` | .go | 484 | No |
| `/Users/kayvan/src/fabric/internal/cli/help.go` | .go | 286 | No |
| `/Users/kayvan/src/fabric/internal/cli/initialization.go` | .go | 57 | No |
| `/Users/kayvan/src/fabric/internal/cli/listing.go` | .go | 129 | No |
| `/Users/kayvan/src/fabric/internal/cli/management.go` | .go | 31 | No |
| `/Users/kayvan/src/fabric/internal/cli/output.go` | .go | 71 | No |
| `/Users/kayvan/src/fabric/internal/cli/output_test.go` | .go | 57 | No |
| `/Users/kayvan/src/fabric/internal/cli/setup_server.go` | .go | 30 | No |
| `/Users/kayvan/src/fabric/internal/cli/tools.go` | .go | 91 | No |
| `/Users/kayvan/src/fabric/internal/cli/transcribe.go` | .go | 37 | No |
| `/Users/kayvan/src/fabric/internal/core/chatter.go` | .go | 285 | No |
| `/Users/kayvan/src/fabric/internal/core/chatter_test.go` | .go | 230 | No |
| `/Users/kayvan/src/fabric/internal/core/plugin_registry.go` | .go | 577 | Yes |
| `/Users/kayvan/src/fabric/internal/core/plugin_registry_test.go` | .go | 98 | No |
| `/Users/kayvan/src/fabric/internal/domain/attachment.go` | .go | 171 | No |
| `/Users/kayvan/src/fabric/internal/domain/domain.go` | .go | 75 | No |
| `/Users/kayvan/src/fabric/internal/domain/domain_test.go` | .go | 27 | No |
| `/Users/kayvan/src/fabric/internal/domain/file_manager.go` | .go | 189 | No |
| `/Users/kayvan/src/fabric/internal/domain/file_manager_test.go` | .go | 185 | No |
| `/Users/kayvan/src/fabric/internal/domain/think.go` | .go | 32 | No |
| `/Users/kayvan/src/fabric/internal/domain/think_test.go` | .go | 19 | No |
| `/Users/kayvan/src/fabric/internal/domain/thinking.go` | .go | 34 | No |
| `/Users/kayvan/src/fabric/internal/i18n/i18n.go` | .go | 240 | No |
| `/Users/kayvan/src/fabric/internal/i18n/i18n_test.go` | .go | 40 | No |
| `/Users/kayvan/src/fabric/internal/i18n/i18n_variants_test.go` | .go | 175 | No |
| `/Users/kayvan/src/fabric/internal/i18n/locale.go` | .go | 94 | No |
| `/Users/kayvan/src/fabric/internal/i18n/locale_test.go` | .go | 288 | No |
| `/Users/kayvan/src/fabric/internal/log/log.go` | .go | 78 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/anthropic/anthropic.go` | .go | 420 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/anthropic/anthropic_test.go` | .go | 327 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/anthropic/oauth.go` | .go | 327 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/anthropic/oauth_test.go` | .go | 433 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/azure/azure.go` | .go | 79 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/azure/azure_test.go` | .go | 90 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/bedrock/bedrock.go` | .go | 274 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/dryrun/dryrun.go` | .go | 136 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/dryrun/dryrun_test.go` | .go | 56 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/exolab/exolab.go` | .go | 52 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/gemini/gemini.go` | .go | 547 | Yes |
| `/Users/kayvan/src/fabric/internal/plugins/ai/gemini/gemini_test.go` | .go | 259 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/gemini/voices.go` | .go | 220 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/gemini_openai/gemini.go` | .go | 15 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/lmstudio/lmstudio.go` | .go | 351 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/models.go` | .go | 84 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/models_test.go` | .go | 104 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/ollama/ollama.go` | .go | 260 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/chat_completions.go` | .go | 121 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/direct_models.go` | .go | 123 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/message_conversion.go` | .go | 21 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/openai.go` | .go | 354 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/openai_audio.go` | .go | 153 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/openai_image.go` | .go | 142 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/openai_image_test.go` | .go | 444 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/openai_models_test.go` | .go | 58 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai/openai_test.go` | .go | 177 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai_compatible/direct_models_call.go` | .go | 13 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai_compatible/providers_config.go` | .go | 250 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/openai_compatible/providers_config_test.go` | .go | 57 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/perplexity/perplexity.go` | .go | 249 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/vendor.go` | .go | 18 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/vendors.go` | .go | 162 | No |
| `/Users/kayvan/src/fabric/internal/plugins/ai/vendors_test.go` | .go | 66 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/api.go` | .go | 13 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/contexts.go` | .go | 32 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/contexts_test.go` | .go | 29 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/db.go` | .go | 106 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/db_test.go` | .go | 55 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/patterns.go` | .go | 275 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/patterns_test.go` | .go | 300 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/sessions.go` | .go | 99 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/sessions_test.go` | .go | 38 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/storage.go` | .go | 158 | No |
| `/Users/kayvan/src/fabric/internal/plugins/db/fsdb/storage_test.go` | .go | 52 | No |
| `/Users/kayvan/src/fabric/internal/plugins/plugin.go` | .go | 334 | No |
| `/Users/kayvan/src/fabric/internal/plugins/plugin_test.go` | .go | 261 | No |
| `/Users/kayvan/src/fabric/internal/plugins/strategy/strategy.go` | .go | 255 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/constants.go` | .go | 5 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/datetime.go` | .go | 144 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/datetime_test.go` | .go | 138 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/extension_executor.go` | .go | 200 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/extension_executor_test.go` | .go | 361 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/extension_manager.go` | .go | 135 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/extension_manager_test.go` | .go | 184 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/extension_registry.go` | .go | 333 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/extension_registry_test.go` | .go | 75 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/fetch.go` | .go | 134 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/fetch_test.go` | .go | 72 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/file.go` | .go | 197 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/file_test.go` | .go | 152 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/hash.go` | .go | 33 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/hash_test.go` | .go | 119 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/sys.go` | .go | 87 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/sys_test.go` | .go | 140 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/template.go` | .go | 151 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/template_extension_mixed_test.go` | .go | 77 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/template_extension_multiple_test.go` | .go | 71 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/template_sentinel_test.go` | .go | 275 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/template_test.go` | .go | 145 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/text.go` | .go | 64 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/text_test.go` | .go | 104 | No |
| `/Users/kayvan/src/fabric/internal/plugins/template/utils.go` | .go | 41 | No |
| `/Users/kayvan/src/fabric/internal/server/auth.go` | .go | 38 | No |
| `/Users/kayvan/src/fabric/internal/server/chat.go` | .go | 232 | No |
| `/Users/kayvan/src/fabric/internal/server/configuration.go` | .go | 138 | No |
| `/Users/kayvan/src/fabric/internal/server/contexts.go` | .go | 19 | No |
| `/Users/kayvan/src/fabric/internal/server/models.go` | .go | 54 | No |
| `/Users/kayvan/src/fabric/internal/server/ollama.go` | .go | 270 | No |
| `/Users/kayvan/src/fabric/internal/server/patterns.go` | .go | 108 | No |
| `/Users/kayvan/src/fabric/internal/server/serve.go` | .go | 86 | No |
| `/Users/kayvan/src/fabric/internal/server/sessions.go` | .go | 19 | No |
| `/Users/kayvan/src/fabric/internal/server/storage.go` | .go | 101 | No |
| `/Users/kayvan/src/fabric/internal/server/strategies.go` | .go | 61 | No |
| `/Users/kayvan/src/fabric/internal/server/youtube.go` | .go | 100 | No |
| `/Users/kayvan/src/fabric/internal/tools/converter/html_readability.go` | .go | 26 | No |
| `/Users/kayvan/src/fabric/internal/tools/converter/html_readability_test.go` | .go | 46 | No |
| `/Users/kayvan/src/fabric/internal/tools/custom_patterns/custom_patterns.go` | .go | 67 | No |
| `/Users/kayvan/src/fabric/internal/tools/custom_patterns/custom_patterns_test.go` | .go | 79 | No |
| `/Users/kayvan/src/fabric/internal/tools/defaults.go` | .go | 78 | No |
| `/Users/kayvan/src/fabric/internal/tools/githelper/githelper.go` | .go | 111 | No |
| `/Users/kayvan/src/fabric/internal/tools/jina/jina.go` | .go | 72 | No |
| `/Users/kayvan/src/fabric/internal/tools/lang/language.go` | .go | 43 | No |
| `/Users/kayvan/src/fabric/internal/tools/notifications/notifications.go` | .go | 128 | No |
| `/Users/kayvan/src/fabric/internal/tools/notifications/notifications_test.go` | .go | 168 | No |
| `/Users/kayvan/src/fabric/internal/tools/patterns_loader.go` | .go | 371 | No |
| `/Users/kayvan/src/fabric/internal/tools/youtube/timestamp_test.go` | .go | 61 | No |
| `/Users/kayvan/src/fabric/internal/tools/youtube/youtube.go` | .go | 840 | Yes |
| `/Users/kayvan/src/fabric/internal/tools/youtube/youtube_optional_test.go` | .go | 19 | No |
| `/Users/kayvan/src/fabric/internal/tools/youtube/youtube_test.go` | .go | 168 | No |
| `/Users/kayvan/src/fabric/internal/util/groups_items.go` | .go | 182 | No |
| `/Users/kayvan/src/fabric/internal/util/oauth_storage.go` | .go | 124 | No |
| `/Users/kayvan/src/fabric/internal/util/oauth_storage_test.go` | .go | 232 | No |
| `/Users/kayvan/src/fabric/internal/util/utils.go` | .go | 91 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/ambient.d.ts` | .ts | 253 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/app.js` | .js | 48 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/matchers.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/0.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/1.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/10.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/11.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/12.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/13.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/14.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/2.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/3.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/4.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/5.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/6.js` | .js | 1 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/7.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/8.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/client/nodes/9.js` | .js | 3 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/generated/server/internal.js` | .js | 53 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/non-ambient.d.ts` | .ts | 54 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/$types.d.ts` | .ts | 28 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/about/$types.d.ts` | .ts | 20 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/api/youtube/transcript/$types.d.ts` | .ts | 10 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/chat/$types.d.ts` | .ts | 28 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/contact/$types.d.ts` | .ts | 20 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/notes/$types.d.ts` | .ts | 10 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/obsidian/$types.d.ts` | .ts | 10 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/posts/$types.d.ts` | .ts | 20 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/posts/[slug]/$types.d.ts` | .ts | 21 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/posts/[slug]/proxy+page.ts` | .ts | 34 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/posts/proxy+page.ts` | .ts | 31 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/tags/$types.d.ts` | .ts | 20 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/tags/[tag]/$types.d.ts` | .ts | 21 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/tags/[tag]/proxy+page.ts` | .ts | 31 | No |
| `/Users/kayvan/src/fabric/web/.svelte-kit/types/src/routes/tags/proxy+page.ts` | .ts | 36 | No |
| `/Users/kayvan/src/fabric/web/eslint.config.js` | .js | 32 | No |
| `/Users/kayvan/src/fabric/web/my-custom-theme.ts` | .ts | 102 | No |
| `/Users/kayvan/src/fabric/web/postcss.config.js` | .js | 6 | No |
| `/Users/kayvan/src/fabric/web/rollup.config.js` | .js | 22 | No |
| `/Users/kayvan/src/fabric/web/src/app.d.ts` | .ts | 9 | No |
| `/Users/kayvan/src/fabric/web/src/index.test.ts` | .ts | 7 | No |
| `/Users/kayvan/src/fabric/web/src/lib/actions/clickOutside.ts` | .ts | 15 | No |
| `/Users/kayvan/src/fabric/web/src/lib/api/base.ts` | .ts | 98 | No |
| `/Users/kayvan/src/fabric/web/src/lib/api/config.ts` | .ts | 27 | No |
| `/Users/kayvan/src/fabric/web/src/lib/api/contexts.ts` | .ts | 13 | No |
| `/Users/kayvan/src/fabric/web/src/lib/api/models.ts` | .ts | 24 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/posts/post-interface.ts` | .ts | 13 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/button/index.js` | .js | 31 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/checkbox/index.ts` | .ts | 4 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/connections/ParticleSystem.ts` | .ts | 110 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/connections/canvas.ts` | .ts | 12 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/connections/colors.ts` | .ts | 4 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/connections/particle.ts` | .ts | 10 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/input/index.ts` | .ts | 4 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/label/index.js` | .js | 6 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/select/index.js` | .js | 6 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/slider/index.js` | .js | 6 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/textarea/index.js` | .js | 6 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/tooltip/Tooltip.test.ts` | .ts | 56 | No |
| `/Users/kayvan/src/fabric/web/src/lib/components/ui/tooltip/positioning.ts` | .ts | 27 | No |
| `/Users/kayvan/src/fabric/web/src/lib/config/environment.ts` | .ts | 62 | No |
| `/Users/kayvan/src/fabric/web/src/lib/config/features.ts` | .ts | 16 | No |
| `/Users/kayvan/src/fabric/web/src/lib/interfaces/chat-interface.ts` | .ts | 52 | No |
| `/Users/kayvan/src/fabric/web/src/lib/interfaces/context-interface.ts` | .ts | 4 | No |
| `/Users/kayvan/src/fabric/web/src/lib/interfaces/model-interface.ts` | .ts | 19 | No |
| `/Users/kayvan/src/fabric/web/src/lib/interfaces/pattern-interface.ts` | .ts | 16 | No |
| `/Users/kayvan/src/fabric/web/src/lib/interfaces/session-interface.ts` | .ts | 7 | No |
| `/Users/kayvan/src/fabric/web/src/lib/interfaces/storage-interface.ts` | .ts | 7 | No |
| `/Users/kayvan/src/fabric/web/src/lib/services/ChatService.ts` | .ts | 284 | No |
| `/Users/kayvan/src/fabric/web/src/lib/services/PdfConversionService.ts` | .ts | 74 | No |
| `/Users/kayvan/src/fabric/web/src/lib/services/pdf-config.ts` | .ts | 19 | No |
| `/Users/kayvan/src/fabric/web/src/lib/services/toast-service.ts` | .ts | 17 | No |
| `/Users/kayvan/src/fabric/web/src/lib/services/transcriptService.ts` | .ts | 80 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/chat-config.ts` | .ts | 24 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/chat-store.ts` | .ts | 170 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/favorites-store.ts` | .ts | 37 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/language-store.ts` | .ts | 13 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/model-store.ts` | .ts | 38 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/note-store.ts` | .ts | 109 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/obsidian-store.ts` | .ts | 68 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/pattern-store.ts` | .ts | 124 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/session-store.ts` | .ts | 93 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/strategy-store.ts` | .ts | 32 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/theme-store.ts` | .ts | 55 | No |
| `/Users/kayvan/src/fabric/web/src/lib/store/toast-store.ts` | .ts | 30 | No |
| `/Users/kayvan/src/fabric/web/src/lib/types/index.ts` | .ts | 4 | No |
| `/Users/kayvan/src/fabric/web/src/lib/utils/file-utils.ts` | .ts | 44 | No |
| `/Users/kayvan/src/fabric/web/src/lib/utils/markdown.ts` | .ts | 27 | No |
| `/Users/kayvan/src/fabric/web/src/lib/utils/utils.ts` | .ts | 6 | No |
| `/Users/kayvan/src/fabric/web/src/lib/utils/validators.ts` | .ts | 4 | No |
| `/Users/kayvan/src/fabric/web/src/routes/+layout.ts` | .ts | 3 | No |
| `/Users/kayvan/src/fabric/web/src/routes/+page.ts` | .ts | 1 | No |
| `/Users/kayvan/src/fabric/web/src/routes/about/+page.ts` | .ts | 5 | No |
| `/Users/kayvan/src/fabric/web/src/routes/api/youtube/transcript/+server.ts` | .ts | 46 | No |
| `/Users/kayvan/src/fabric/web/src/routes/chat/+page.ts` | .ts | 5 | No |
| `/Users/kayvan/src/fabric/web/src/routes/chat/+server.ts` | .ts | 163 | No |
| `/Users/kayvan/src/fabric/web/src/routes/contact/+page.ts` | .ts | 5 | No |
| `/Users/kayvan/src/fabric/web/src/routes/notes/+server.ts` | .ts | 35 | No |
| `/Users/kayvan/src/fabric/web/src/routes/obsidian/+server.ts` | .ts | 105 | No |
| `/Users/kayvan/src/fabric/web/src/routes/posts/+page.ts` | .ts | 29 | No |
| `/Users/kayvan/src/fabric/web/src/routes/posts/[slug]/+page.ts` | .ts | 33 | No |
| `/Users/kayvan/src/fabric/web/src/routes/tags/+page.ts` | .ts | 34 | No |
| `/Users/kayvan/src/fabric/web/src/routes/tags/[tag]/+page.ts` | .ts | 30 | No |
| `/Users/kayvan/src/fabric/web/svelte.config.js` | .js | 103 | No |
| `/Users/kayvan/src/fabric/web/tailwind.config.ts` | .ts | 142 | No |
| `/Users/kayvan/src/fabric/web/vite.config.ts` | .ts | 92 | No |

---
*Generated by fabric code_analysis tool*
