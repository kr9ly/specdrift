<!-- specdrift v1 -->

# AI コーディングエージェントへの specdrift 導入ガイド

specdrift のインストールから、AI コーディングエージェントを活用した開発プロセスへの組み込みまでを解説します。

プロジェクトに既存のドキュメント（設計ドキュメント、ルール定義、ガイドなど）がある場合、新しい仕様書ファイルを作る必要はありません。既存のドキュメントに直接 specdrift のアノテーションを追加できます。

## Part 1: ツールの導入

<!-- source: internal/annotation.go@352e344f, internal/config.go@06b50e71 -->

### インストール

```bash
go install github.com/kr9ly/specdrift@latest
```

### 初期化

プロジェクトのルートで実行します。ソースパスの起点となる `.specdrift` マーカーファイルが作成されます。

```bash
cd /path/to/your/project
specdrift init
```

### 既存ドキュメントにアノテーションを追加する

プロジェクトに既存のドキュメントがある場合、これが最も手軽な導入方法です。

1. ドキュメントの先頭に specdrift 宣言を追加します：

```markdown
<!-- specdrift v1 -->
```

2. 特定のソースファイルについて記述しているセクションを source アノテーションで囲みます：

```markdown
<!-- source: src/auth/handler.go@TODO -->
このハンドラは認証情報を検証し、JWT トークンを返す。
<!-- /source -->
```

> **重要:** `@TODO` または `@hash` の接尾辞は必須です。`src/auth/handler.go` のように `@` なしで記述するとエラーになります。

3. `specdrift update` を実行して、既存ファイルのハッシュを解決します：

```bash
specdrift update docs/design/auth.md
```

4. `specdrift check` で検証します：

```bash
specdrift check docs/design/auth.md
```

### 新しい仕様書を書く（仕様先行ワークフロー）

あるいは、仕様書ファイル（例: `docs/spec/auth.md`）を新規作成し、specdrift の宣言を追加します。

```markdown
<!-- specdrift v1 -->

# 認証

<!-- source: TODO -->
メールアドレスとパスワードで認証する。
ハンドラは入力を検証し、認証情報を確認してトークンを返す。
<!-- /source -->
```

これが **仕様先行ワークフロー** です。実装より先に仕様を書き、`TODO` プレースホルダでソースファイルが未作成であることを示します。

### コードの進行に合わせて仕様を更新する

ファイルパスが決まったら：

```markdown
<!-- source: src/auth/handler.go@TODO -->
```

ファイルが存在する状態で `update` を実行してハッシュを解決します：

```bash
specdrift update docs/spec/auth.md
```

アノテーションがこうなります：

```markdown
<!-- source: src/auth/handler.go@a1b2c3d4 -->
```

以降、`specdrift check` がソースファイルの変更を検出します。

### 検証

```bash
specdrift check docs/spec/auth.md
```

<!-- /source -->

## Part 2: 開発プロセスへの組み込み

<!-- source: internal/checker.go@913de33e, internal/updater.go@a22df0ea -->

### 基本原則

specdrift が強制するルールは一つ：**ソースコードが変わったら、仕様書をレビューすること。** ハッシュの不一致（DRIFT）は仕様書が間違っていることを意味するとは限りません。仕様がまだコードを正確に記述しているかどうかを、誰かが判断する必要があるということです。

AI コーディングエージェントにとって、これは特に有用です。エージェントはコードを自由に変更できますが、specdrift がチェックポイントを設けます。コミット前に、仕様との整合性を明示的に確認しなければなりません。

### コミットチェックに specdrift を追加する

`specdrift check` をコミット前のワークフローに組み込みます。具体的な仕組みはエージェントやプロジェクトによりますが、パターンは共通です。

**チェックを順番に実行し、最初の失敗で停止する。**

1. フォーマッタ / リンタ（`gofmt`, `eslint`, `ruff` など）
2. 静的解析（`go vet`, `tsc --noEmit` など）
3. テスト
4. `specdrift check 'docs/spec/*.md'`
5. コミット

ポイント：specdrift check は **テストの後**、**コミットの前** に実行します。変更の文脈が新鮮なうちに仕様のずれに対処できます。

プロジェクトの構成に合わせて glob パターンを選びます：

- `'docs/spec/*.md'` — 仕様書専用ディレクトリがある場合
- `'docs/**/*.md'` — 既存ドキュメント全体にアノテーションが散在する場合（`<!-- specdrift v1 -->` 宣言のないファイルは自動的にスキップされます）

#### 例: Claude Code カスタムコマンド

プロジェクトに `.claude/commands/commit.md` を作成します：

```markdown
チェックを実行してコミットする。順番に実行し、最初の失敗で停止。

1. **format**: フォーマッタを実行
2. **lint/vet**: 静的解析を実行
3. **test**: テストスイートを実行
4. **specdrift check**: `specdrift check 'docs/spec/*.md'`
   - DRIFT が検出された場合、すぐに update しない。まず差分のある仕様と変更されたソースを
     読み、仕様の文面を修正する必要があるか判断する。必要なら仕様を更新してから
     `specdrift update` でハッシュを同期する。
5. **commit**: ステージングしてコミット
```

#### 例: Git pre-commit フック

```bash
#!/bin/sh
specdrift check 'docs/spec/*.md'
```

### DRIFT の対応：レビュールール

`specdrift check` が DRIFT を報告したときの正しい対応手順：

1. **差分のある仕様セクションを読む** — 仕様が何を言っているか理解する
2. **変更されたソースコードを読む** — 実際に何が変わったか理解する
3. **判断する**: 仕様はまだコードの振る舞いを正しく記述しているか？
   - はい → `specdrift update` でハッシュを同期
   - いいえ → 仕様の文面を修正してから `specdrift update`

**仕様を確認せずに `specdrift update` を実行しないこと。** 無言の更新はツールの意味を無にします。これがエージェントに伝えるべき最も重要なルールです。

Claude Code の場合、CLAUDE.md で明示できます：

```markdown
specdrift check で DRIFT が検出された場合、update する前に仕様と変更されたソースを読むこと。
仕様の文面の修正が必要かどうかを確認せずに `specdrift update` を実行しないこと。
```

### プロジェクト指示書への記載

エージェントがプロジェクトの文脈として読むファイル（`CLAUDE.md`, `AGENTS.md`, `CONVENTIONS.md` など）に specdrift を記載します。含めるべき情報：

1. **ビルドとチェックの方法**: 実行するコマンド
2. **レビュールール**: 無言の update 禁止
3. **仕様書の場所**: 仕様ファイルの配置先（例: `docs/spec/`）

CLAUDE.md の記載例：

```markdown
## 仕様ドリフト検出

このプロジェクトでは specdrift を使って仕様書とソースコードの同期を管理しています。

- チェック: `specdrift check 'docs/spec/*.md'`
- 更新（レビュー後）: `specdrift update 'docs/spec/*.md'`
- DRIFT が検出された場合、update する前に仕様と変更されたソースの両方を読むこと。
```

### 仕様先行ワークフロー

specdrift は TODO プレースホルダにより、コードより先に仕様を書くワークフローをサポートします：

1. **仕様を書く** — `<!-- source: TODO -->` でコードが何をすべきか記述する
2. **コードを実装する** — エージェントがソースファイルを作成する
3. **パスを記入する** — `<!-- source: path/to/file.go@TODO -->` に変更
4. **update を実行する** — `specdrift update` でハッシュを解決

このワークフローは AI コーディングエージェントと自然に噛み合います。仕様が構造化されたプロンプトとして機能し、望ましい振る舞いを記述します。エージェントがそれを実装し、ハッシュアノテーションが仕様と実装を紐付けて、以降のドリフトを検出します。

### 仕様書を管理しやすく保つ

実践的なヒント：

- **モジュールや機能ごとに1つの仕様ファイル** — 巨大な仕様書は避ける
- **アノテーションのスコープを絞る** — ドキュメント全体ではなく、特定のファイルについて述べているセクションにアノテーションを付ける
- **階層的な仕様にはネストを使う** — 外側のスコープでモジュール全体、内側で個別ファイル
- **すべてのファイルにアノテーションを付けない** — 意図が重要なファイル（ハンドラ、コアロジック）に集中し、ボイラープレートは対象外にする
- **複数ドキュメントから同一ソースを参照してよい** — これは正常な状態（例: 設計ドキュメントとルールドキュメントが同じファイルを追跡する）

#### アノテーション対象の選び方

- ドキュメントが特定のソースファイルの **使い方** を記述している → アノテーションする
- ドキュメントが特定の実装に依存しない **一般的な概念** を説明している → アノテーション不要
- 間接的に関連するファイルにはアノテーションしない — ノイズと誤検出の原因になる

<!-- /source -->

## Part 3: 発展的な使い方

<!-- source: internal/coverage.go@f9c6194b, internal/graph.go@a6f63d82, internal/ignore.go@56954b8e -->

### ドキュメンテーションカバレッジ

`specdrift coverage` は、ソースコードのうちどれだけが仕様書で追跡されているかを計測します。

```bash
specdrift coverage --src 'src/**/*.go' 'docs/spec/*.md'
```

出力例：

```
Coverage: 9/12 (75.0%)

Covered:
  src/auth/handler.go  <- docs/spec/auth.md
  src/db/client.go     <- docs/spec/db.md
  ...

Not covered:
  src/util/hash.go
  src/util/strings.go
  src/middleware/cors.go
```

`--src` フラグ（複数指定可）で計測対象のソースファイルを指定します：

```bash
# 複数のソースディレクトリ
specdrift coverage --src 'src/**/*.go' --src 'pkg/**/*.go' 'docs/**/*.md'
```

#### .specdriftignore によるファイル除外

プロジェクトルート（`.specdrift` と同じ場所）に `.specdriftignore` ファイルを作成すると、カバレッジ計測からファイルを除外できます：

```
# テストファイル
*_test.go

# 自動生成コード
*.gen.go
generated/*
```

パターンはフルの相対パスとベースファイル名の両方に対してマッチします。このファイルは `graph --reverse` の出力にも影響します。

#### コミットワークフローへのカバレッジチェックの追加

`specdrift check` の後にカバレッジチェックのステップを追加できます。閾値を強制するのではなく、"Not covered" のリストを表示して開発者（またはエージェント）にドキュメントの要否を判断させるのが実用的です：

```markdown
5. **coverage**: `specdrift coverage --src 'src/**/*.go' 'docs/spec/*.md'`
   - "Not covered" のファイルがある場合、リストを確認し、コミット前に
     ドキュメントが必要かどうかを判断する。
6. **commit**: ステージングしてコミット
```

すべてのファイルにドキュメントが必要なわけではありません。ユーティリティ関数、シンプルなラッパー、ボイラープレートは仕様追跡の恩恵を受けることが稀です。カバレッジレポートは、見落としではなく意識的な判断でその決定を行う助けになります。

### 依存グラフ

`specdrift graph` は仕様書とソースファイルの関係を可視化します。

#### 正引き: 各仕様書は何をカバーしているか？

```bash
specdrift graph 'docs/spec/*.md'
```

```
docs/spec/auth.md
  -> src/auth/handler.go
  -> src/auth/token.go
docs/spec/db.md
  -> src/db/client.go
```

#### 逆引き: このファイルを追跡している仕様書は？

```bash
specdrift graph --reverse 'docs/spec/*.md'
```

```
src/auth/handler.go
  -> docs/spec/auth.md
src/db/client.go
  -> docs/spec/db.md
  -> docs/spec/migration.md
```

逆引きグラフは **「このファイルを変更したら、どの仕様書を更新する必要があるか？」** に答えます。変更前の影響分析に有用です — ソースファイルを修正する前に、どの仕様書がそれを参照しているかを確認することで、ドキュメントへの影響を把握できます。

<!-- /source -->
