# auto-font-by-ppi

対象ディスプレイの PPI に応じて、GNOME のフォント設定を自動調整するスクリプトです。

`gnome-auto-font-by-ppi.sh` は使用中のディスプレイを検出し、PPI を計算し、対応するプロファイルを選択して、GNOME の `text-scaling-factor` や各種フォントサイズを更新します。

## 主な機能

- Xorg では `xrandr` を使用
- GNOME Wayland では Mutter DisplayConfig D-Bus を使用
- プライマリディスプレイの自動選択、または名前指定での選択
- テキストスケーリングのみ、またはフォント一式の適用に対応
- EDID の物理サイズが壊れている環境向けに手動サイズ上書きをサポート
- コマンドライン、環境変数、設定ファイルに対応

## 必要条件

- Bash
- `gsettings`
- Xorg セッションでは `xrandr`
- GNOME Wayland セッションでは `gdbus` と `python3`

## 使い方

```bash
./gnome-auto-font-by-ppi.sh --dry-run
./gnome-auto-font-by-ppi.sh --apply
./gnome-auto-font-by-ppi.sh --apply --display eDP-1
./gnome-auto-font-by-ppi.sh --apply --mode scaling_only
```

## Go CLI の現状

Go リライト版では、Bash スクリプトと並行して実験的な `auto-font-by-ppi` CLI を提供しています。

利用例:

```bash
go run ./cmd/auto-font-by-ppi --dry-run
go run ./cmd/auto-font-by-ppi --apply --display eDP-1
go run ./cmd/auto-font-by-ppi --dry-run --display-diagonal HDMI-1=31.5
```

Go CLI が読む TOML 設定ファイルの既定パスは次のとおりです。

```text
~/.config/auto-font-by-ppi/config.toml
```

現在の設定形式は [examples/config.toml](./examples/config.toml) を参照してください。

## モニタサイズの手動上書き

モニタによっては、EDID から取得した物理サイズが不正な場合があります。典型的には、ミリメートル値がピクセル解像度と同じになってしまい、明らかに不正な PPI が計算されます。

その場合は、対角インチを手動指定してください。

### コマンドライン

```bash
./gnome-auto-font-by-ppi.sh --dry-run --display-diagonal HDMI-1=31.5
./gnome-auto-font-by-ppi.sh --dry-run --display-diagonal HDMI-1=31.5 --display-diagonal eDP-1=14.0
```

### 環境変数

```bash
MONITOR_DIAGONAL_OVERRIDES="HDMI-1=31.5,eDP-1=14.0" ./gnome-auto-font-by-ppi.sh --dry-run
```

## 設定ファイル

デフォルトでは、次のファイルを読み込みます。

```text
~/.config/gnome-auto-font-by-ppi.conf
```

別の設定ファイルを使う場合は `--config` を指定します。

```bash
./gnome-auto-font-by-ppi.sh --config ./gnome-auto-font-by-ppi.conf.example --dry-run
```

設定例:

```bash
PREFERRED_DISPLAY="HDMI-1"
APPLY_MODE="full_fonts"
MONITOR_DIAGONAL_OVERRIDES="HDMI-1=31.5,eDP-1=14.0"
MIN_REASONABLE_PPI=50
MAX_REASONABLE_PPI=400
```

詳細は [gnome-auto-font-by-ppi.conf.example](./gnome-auto-font-by-ppi.conf.example) を参照してください。

## 検出とフォールバックの動作

次の条件では、取得した物理サイズを怪しい値として扱います。

- ミリメートルの幅または高さが 0 以下
- ミリメートルの幅と高さがピクセル解像度と完全一致する
- 計算された PPI が設定された妥当範囲の外にある

怪しい値が検出され、かつそのディスプレイに対する手動対角インチ指定がある場合は、そのインチ値から物理サイズと PPI を再計算します。

## オプション

```text
--dry-run
--apply
--display NAME
--display-diagonal NAME=INCHES
--config PATH
--mode scaling_only|full_fonts
--verbose
--help
```

現在の Go CLI でサポートしているオプション:

```text
--dry-run
--apply
--display NAME
--display-diagonal NAME=INCHES
--config PATH
--verbose
--help
```

## プロファイル

現在は次の PPI プロファイルを使用します。

```text
110  -> scaling 1.00, UI 11, document 11, monospace 10, titlebar 11
140  -> scaling 1.00, UI 12, document 12, monospace 11, titlebar 12
180  -> scaling 1.00, UI 13, document 13, monospace 12, titlebar 13
240  -> scaling 1.00, UI 14, document 14, monospace 13, titlebar 14
9999 -> scaling 1.60, UI 16, document 16, monospace 14, titlebar 16
```

## ライセンス

MIT License。詳細はソースファイル内のライセンスヘッダを参照してください。
