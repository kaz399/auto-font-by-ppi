# Go リライト進捗

## 目的

この文書は、`auto-font-by-ppi` の Go リライトについて、設計方針、これまでに完了した作業、残っている項目をまとめたものです。

現在のリライト作業は `go-rewrite-skeleton` ブランチで進めています。既存の Bash 実装は、移行期間中の参照実装およびフォールバックとして引き続きリポジトリ内に残しています。

## リライトの目標

- 単一の Bash スクリプトから、保守しやすい Go 製 CLI へ移行する
- ディスプレイ検出、PPI/profile 選択、target への適用を明確に分離する
- GNOME 以外の複数 target を扱える構造にする
- dry-run と apply の挙動を明示的かつテスト可能にする
- backend や target adapter を追加しやすい形にする

## 設計・実装方針

### 基本方針

- 実行形式は単一の CLI バイナリとする
- レイヤー構成を維持する
  - `display`: 接続中ディスプレイと物理サイズ情報の取得
  - `profile`: 値の正規化、異常値判定、対角インチ補正、profile 選択
  - `target`: 解決済み設定を各プログラム向けの具体的な action に変換
  - `app`: 検出、選択、plan 生成、dry-run 表示、apply 実行を統括
- 実行フローは plan ベースにする
  - 検出と解決
  - action 生成
  - plan 表示
  - dry-run でなければ apply
- main の分岐を target 固有処理で汚さず、adapter 追加で拡張する
- 設定は `config.toml` ベースにする

### 現在のアーキテクチャ

```text
cmd/auto-font-by-ppi/
internal/app/
internal/cli/
internal/config/
internal/display/
internal/execx/
internal/model/
internal/profile/
internal/target/
examples/config.toml
```

### 主なデータフロー

1. CLI フラグを解析する
2. `config.toml` を読み込む
3. 設定された backend 順でディスプレイを検出する
4. 異常なディスプレイ物理値を正規化し、必要なら対角インチ補正を適用する
5. 対象ディスプレイを選択する
6. 検出した PPI から profile を選択する
7. 有効な target から plan を組み立てる
8. plan を表示する
9. `dry_run = false` のときだけ action を適用する

## これまでに実装した内容

### Go CLI の雛形

- `go.mod` を追加
- `cmd/auto-font-by-ppi/main.go` に最小のエントリポイントを追加
- app、config、display backend、profile、execution、target adapter の internal package を追加

### 設定まわり

- `examples/config.toml` を追加
- `github.com/BurntSushi/toml` を使った TOML パースを追加
- Go 側にデフォルト設定値を追加
- 部分的な config でも必要な値だけ上書きできるよう、デフォルト値とのマージを維持
- decode 時に未対応の config key を拒否
- 以下の設定をサポート
  - `dry_run`
  - `preferred_display`
  - `display_backend_priority`
  - `target_names`
  - display sanity thresholds
  - diagonal overrides
  - font families
  - profiles
  - `target.gnome`
  - `target.kitty`

### Display backend

- X11/Xorg 向けに `xrandr` backend を実装
- GNOME Wayland 向けに `gdbus` と Mutter `DisplayConfig` を使う `gnome-wayland` backend を実装
- デフォルトの backend 優先順を `gnome-wayland` → `xrandr` に設定

### Profile ロジック

- PPI 計算を実装
- 異常値判定を実装
- 対角インチ override の適用を実装
- 対象ディスプレイ選択を実装
  - 明示指定された display
  - primary display
  - 先頭 display へのフォールバック
- `max_ppi` に基づく profile 選択を実装

### Target adapter

- GNOME target adapter を実装
  - `scaling_only`
  - `full_fonts`
- kitty target adapter を実装
  - `remote` strategy のみ
  - 任意の `--to` socket
  - font size field の切り替え

### Plan と実行

- `buildPlan` を実装
- `applyPlan` を実装
- selected display、selected profile、生成 action を dry-run 表示する処理を実装

## 追加済みテスト

### Config テスト

- サンプル config のパース確認
- config 値がデフォルト値を正しく上書きすることの確認
- 未知の root key を拒否することの確認
- 未知のネストした key を拒否することの確認

### Display テスト

- GNOME Wayland の `gdbus` 出力パース確認
- 重複ディスプレイの除去確認
- primary display 判定の確認

### Profile テスト

- 異常な物理値に対する diagonal override 適用確認
- preferred display と primary display の選択確認
- 指定 PPI に対する profile 選択確認

### Target テスト

- GNOME の `full_fonts` / `scaling_only` action 生成確認
- GNOME の不正 mode 拒否確認
- kitty remote command 生成確認
- kitty の不正 strategy / 不正 font size field 拒否確認
- apply 時に adapter が runner に正しく委譲することの確認

### App テスト

- `buildPlan` が target 順序を維持することの確認
- 無効 target を skip することの確認
- 不明 target でエラーになることの確認
- adapter の build error が伝播することの確認
- `applyPlan` が action を順番に実行することの確認
- `applyPlan` が最初の adapter error で停止することの確認

### 検証状況

現在の Go 実装は以下でコンパイルおよびテスト通過を確認しています。

```bash
go test ./...
```

開発時の sandbox 環境では、Go のキャッシュ先を `/tmp` 配下へ向けて実行しています。

## 現在の制約

- Wayland 対応は現在 GNOME Mutter 専用です
- Sway や Hyprland など、非 GNOME の Wayland compositor には未対応です
- kitty 対応は、実行中の kitty を remote control で変更する方式のみです
- kitty の永続設定ファイル更新は未実装です
- Go 実装が十分な機能互換に達するまでは、既存 Bash 実装が実質的な本命実装です
- Go CLI のユーザー向けドキュメントはまだ十分ではありません

## 残項目

### 優先度高

- 既存 Bash スクリプトと必要な範囲で機能互換を取る
- Go CLI を既存スクリプト名へ置き換えるか、並行運用するかを決める
- Go CLI の利用方法と移行手順を文書化する

### Display/backend 関連

- 必要に応じて追加の Wayland compositor 対応を入れる
  - Sway 向けの `swaymsg`
  - Hyprland 向けの `hyprctl`
- `testdata/` に backend ごとの fixture を追加する
- Mutter 出力形式の変化に備えて、display metadata のパースをさらに堅牢化する

### Target 関連

- 任意の CLI 変更可能プログラム向けに generic command target を追加する
- kitty の永続設定をどの形でサポートするか決める
- 必要に応じて target adapter を追加する
  - terminal emulator
  - editor
  - compositor 固有設定

### App/CLI 関連

- 現在の最小実装から CLI フラグを拡張する
- backend のフォールバック判断を verbose でより詳しく出せるようにする
- runner 全体を通す integration に近いテストを追加する

### ドキュメント・配布

- `README.md` と `README_ja_JP.md` を Go 実装に合わせて更新する
- Bash 版からの移行ノートを追加する
- Go バイナリの build/install 手順を書く
- 自動実行向けに systemd user service の例を検討する

## 推奨する次の作業

1. 任意プログラム向けの generic command target adapter を追加する
2. `testdata/` に `xrandr` と GNOME Wayland の fixture を追加する
3. Go ツール単体で利用評価できる程度まで CLI とドキュメントを拡充する
4. decode 時の未知 key 拒否に加えて、追加の config validation が必要か検討する
5. 日常利用に耐える機能が揃ってから Bash スクリプトの引退を判断する
