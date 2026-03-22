# specdrift

仕様書とソースコードの乖離を検出する。

Markdown の仕様書にソースファイル参照をハッシュアノテーションとして埋め込み、ソースが変更されると `specdrift check` が検知する。

## インストール

```bash
go install github.com/kr9ly/specdrift@latest
```

## クイックスタート

プロジェクトルートで初期化:

```bash
specdrift init
```

仕様書に宣言とソースアノテーションを追加:

```markdown
<!-- specdrift v1 -->

<!-- source: path/to/handler.go@TODO -->
Handler の仕様をここに書く。
<!-- /source -->
```

プロジェクト全体のハッシュ解決と乖離チェック:

```bash
specdrift update '**/*.md'
specdrift check '**/*.md'
```

個別ファイルを指定することもできる:

```bash
specdrift check docs/spec.md
```

## アノテーションフォーマット

詳細は [docs/format/annotation-format.ja.md](docs/format/annotation-format.ja.md) を参照。

English version: [docs/format/annotation-format.md](docs/format/annotation-format.md)
