# auto-font-by-ppi

対象ディスプレイの PPI に応じて、GNOME のフォント設定や kitty のフォントサイズを自動調整する Go 製の CLI ツールです。

使用中のディスプレイを検出し、PPI を計算し、対応するプロファイルを選択して、GNOME の `text-scaling-factor` や各種フォントサイズ、および kitty のフォントサイズを更新します。

## 主な機能

- **マルチバックエンド検出**:
  - Xorg では `xrandr` を使用
  - GNOME Wayland では Mutter DisplayConfig D-Bus を使用
- **堅牢なディスプレイ選択**:
  - プライマリディスプレイの自動選択、または名前指定（`--display`）での選択
  - `--auto` オプションにより設定ファイルの優先設定（`preferred_display`）を無視して現在の接続環境から自動選択
- **柔軟な適用モード**:
  - テキストスケーリングのみ、またはフォント一式の適用に対応（GNOME）
  - 現在アクティブな kitty インスタンスへのフォントサイズ動的反映（kitty ターゲット）
- **異常値補正・フォールバック**:
  - EDID の物理サイズが壊れている環境向けに、手動での対角サイズ上書き（`--display-diagonal` または設定ファイル）をサポート
  - 手動対角インチ指定がある場合は、検出された物理サイズより常に優先して適用
  - GNOME Wayland でミリメートル値が取れない場合に `xrandr` から補完
  - 使える物理サイズが無い場合に `PPI=100` 仮定へフォールバック
- **プロファイルごとの個別設定**:
  - profile ごとに専用の `kitty_font_size` を設定可能

## 必要条件

- `gsettings` (GNOME の設定反映に必要)
- Xorg セッション：`xrandr`
- GNOME Wayland セッション：`gdbus`（自動で `xrandr` からの補完も行います）

## ビルド方法

リポジトリのルートで以下のコマンドを実行してビルドします。

```bash
# Build the binary
go build -o auto-font-by-ppi ./cmd/auto-font-by-ppi
```

## 使い方

基本の実行：

```bash
# Run with dry-run to preview changes
./auto-font-by-ppi --dry-run

# Apply settings (Apply mode is default)
./auto-font-by-ppi

# Select a specific display by name
./auto-font-by-ppi --display eDP-1

# Ignore preferred display in config and auto-detect instead
./auto-font-by-ppi --auto
```

一時的な上書き（特定のディスプレイ選択後に適用されます）：

```bash
# Assume the selected display is 27 inches
./auto-font-by-ppi --inch 27

# Assume the selected display has 110 PPI
./auto-font-by-ppi --ppi 110
```

特定のモニターの物理対角サイズを明示的に上書き：

```bash
# Set manual diagonal override for HDMI-1
./auto-font-by-ppi --display-diagonal HDMI-1=31.5

# Specify multiple overrides
./auto-font-by-ppi --display-diagonal HDMI-1=31.5 --display-diagonal eDP-1=14.0
```

## 設定ファイル

既定の設定ファイルのパスは以下のとおりです。

```text
~/.config/auto-font-by-ppi/config.toml
```

このファイルが存在しない場合、初回実行時に既定値ベースのサンプル設定がこの場所に自動生成されます。

別の設定ファイルを使う場合は `--config` を指定します。

```bash
# Load custom configuration path
./auto-font-by-ppi --config ./custom-config.toml
```

設定形式の詳細は [examples/config.toml](./examples/config.toml) を参照してください。

### 主要な設定項目
* `dry_run`: `true` に設定すると常にプレビューのみ行います。
* `preferred_display`: 常に優先して選択するディスプレイの名前（例: `"HDMI-1"`）。`--auto` フラグを指定することで一時的に無視できます。
* `target_names`: 適用するターゲットのリスト（例: `["gnome", "kitty"]`）。kitty は既定では無効（opt-in）です。
* `display.diagonal_overrides`: モニターごとの手動対角インチ上書きマップ。

## 検出とフォールバックの動作

次の条件に合致する場合、取得した物理サイズを異常値として扱います。

- ミリメートルの幅または高さが 0 以下
- ミリメートルの幅と高さがピクセル解像度と完全一致する
- 計算された PPI が設定された妥当範囲の外にある（既定: 50 〜 400）

異常値が検出された場合、以下の優先順で補正が行われます。

1. `display.diagonal_overrides` または `--display-diagonal` に上書き指定があれば、そのインチ値から物理サイズと PPI を計算します。
2. 上書きがない場合、`gnome-wayland` バックエンドからの取得結果を優先的に使用します。
3. `gnome-wayland` でミリメートル値が取れない場合、`xrandr` からの情報を補完します。
4. 最後まで使える物理サイズが取得できない場合は、`PPI=100` を仮定して物理サイズを導出して処理を継続します。

## オプション

```text
--dry-run
    設定の適用を行わずに、変更内容のプレビューのみを表示します。
--display NAME
    自動検出を行わず、指定した名前のディスプレイを選択します。
--auto
    設定ファイルに preferred_display が設定されていても無視し、接続中の最適なディスプレイを自動検出して選択します（--display との同時指定はできません）。
--inch INCHES
    選択されたディスプレイを指定した対角インチ数として扱って PPI を計算します（config の上書きより優先）。
--ppi VALUE
    選択されたディスプレイを指定した PPI として扱います（config の上書きより優先）。
--display-diagonal NAME=INCHES
    特定のディスプレイ名に対して物理対角サイズをインチ数で上書き定義します。
--config PATH
    読み込む TOML 設定ファイルのパスを指定します。
--verbose
    詳細な診断ログを出力します。
--help, -h
    ヘルプメッセージを表示して終了します。
```

## プロファイル

PPI に応じて、以下のプロファイルから設定が選択されます。

```text
110  -> scaling 1.00, UI 11, document 11, monospace 10, titlebar 11, kitty 10
140  -> scaling 1.00, UI 12, document 12, monospace 11, titlebar 12, kitty 11
180  -> scaling 1.00, UI 13, document 13, monospace 12, titlebar 13, kitty 12
240  -> scaling 1.00, UI 14, document 14, monospace 13, titlebar 14, kitty 13
9999 -> scaling 1.60, UI 16, document 16, monospace 14, titlebar 16, kitty 14
```

## ライセンス

MIT License。詳細はソースファイル内のライセンスヘッダを参照してください。
