# specdrift アノテーションフォーマット (v1)

仕様書（Markdown）にソースコード参照を埋め込み、乖離を検出可能にする。

## 宣言

ファイル先頭付近に以下を記述する。宣言がないファイルは specdrift の対象外として扱われる。

```markdown
<!-- specdrift v1 -->
```

バージョン番号は省略可能（省略時は v1）。

## ソース参照

仕様の記述範囲を `source` タグで囲み、対応するソースファイルを紐づける。

### 基本の流れ

仕様を書く時点ではまず TODO で宣言し、実装後に `update` でハッシュを解決する。

```markdown
<!-- source: TODO -->
この機能の仕様。実装先は未定。
<!-- /source -->
```

パスが決まったら記入する。ハッシュは `TODO` のままでよい。

```markdown
<!-- source: path/to/handler.go@TODO -->
handler の仕様。
<!-- /source -->
```

`specdrift update` を実行すると、ファイルの SHA-256 先頭 8 文字でハッシュが埋まる。

```markdown
<!-- source: path/to/handler.go@a1b2c3d4 -->
handler の仕様。
<!-- /source -->
```

以降、ソースファイルが変更されると `specdrift check` が DRIFT を検出する。

### 複数ファイル参照

カンマ区切りで 1 つのタグに複数ファイルを指定できる。

```markdown
<!-- source: handler.go@a1b2c3d4, handler_test.go@e5f6a7b8 -->
```

### ネスト

source タグはネストできる。外側のスコープと内側のスコープを分けて記述可能。

```markdown
<!-- source: api/router.go@a1b2c3d4 -->
ルーティング全体の仕様。

  <!-- source: api/middleware.go@e5f6a7b8 -->
  ミドルウェアに関する仕様。
  <!-- /source -->

<!-- /source -->
```

## セットアップ

プロジェクトルートで `init` を実行する。`.specdrift` ファイルが生成され、ソースパスの起点になる。

```bash
specdrift init
```

## コマンド

```bash
# 乖離チェック
specdrift check docs/spec.md

# ハッシュを現在のファイル内容で更新（path@TODO も解決される）
specdrift update docs/spec.md

# ソースパスの起点を明示的に指定（.specdrift より優先）
specdrift check --base /path/to/repo docs/spec.md
```

## チェック結果

| 状態    | 意味                                 |
| ------- | ------------------------------------ |
| OK      | ハッシュ一致                         |
| DRIFT   | ハッシュ不一致（ソースが変更された） |
| MISSING | ソースファイルが存在しない           |
| TODO    | 未実装のプレースホルダ               |

DRIFT, MISSING, TODO のいずれかがあると exit 1 を返す。
