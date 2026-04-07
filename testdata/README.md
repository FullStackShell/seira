# testdata

`go test` で使用されるテスト用シェルスクリプト群。各ディレクトリが1つのテストケースに対応する。

## バンドラー / 依存解決テスト

| ディレクトリ | 概要 |
|---|---|
| `simple/` | 依存なしの最小プロジェクト。`main()` が1つだけ。 |
| `deps/` | `source lib/helper.sh` で1つの依存を持つ基本ケース。 |
| `circular/` | `a.sh` ↔ `b.sh` の循環依存。`CycleError` の検出テストに使用。 |
| `varpath/` | `${SEIRA_ROOTDIR="."}/lib/dep.sh` のように変数展開を含むsourceパス。 |
| `sideeffect/` | sourceされるファイルにトップレベルの変数代入（副作用）がある場合のconcatモード処理。 |
| `directive/` | `# @seira:ignore` ディレクティブにより特定のsource行をバンドル対象から除外。 |
| `using/` | `# @seira:using str` ディレクティブによる名前空間エイリアス生成 (`str::upper` → `upper`)。 |
| `consumer/` | ライブラリ依存を持つプロジェクト。`deps/strutils/` がライブラリとして `.seirarc.json` で定義されている。ライブラリモード・エクスポート機能のテストにも使用。 |

## Lintテスト

| ディレクトリ | 概要 |
|---|---|
| `lint-bashsource/` | `BASH_SOURCE[0]` の使用を検出する `bash-source` ルール。concatモードでは警告、tarballモードでは問題なし。 |
| `lint-collision/` | 複数ファイルで同名の関数 (`helper()`) を定義。`func-collision` ルールの検出テスト。 |
| `lint-condsource/` | `if` 文内の条件付き `source`。`conditional-source` ルールの検出テスト。 |

## 使用箇所

- `internal/bundler/bundler_new_test.go` — バンドル生成・E2Eテスト
- `internal/depgraph/resolve_test.go` — 依存グラフ解決・トポロジカルソートテスト
- `internal/lint/lint_test.go` — Lintルールのテスト
