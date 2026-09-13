package main

import (
	"bufio"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sonooverlay "github.com/Nexints/Sono-Overlay/pkg/sono-overlay"
	"github.com/Nexints/Sono-Overlay/pkg/sonolus"
	"github.com/fatih/color"
	"github.com/google/go-github/v57/github"
	"github.com/srinathh/gokilo/rawmode"
	"golang.org/x/sys/windows"
)

func checkUpdate() (string, string) {
	githubClient := github.NewClient(nil)
	release, _, err := githubClient.Repositories.GetLatestRelease(context.Background(), "Nexints", "Sono-Overlay")
	if err != nil {
		return "", ""
	}

	latestVersion := strings.TrimPrefix(release.GetTagName(), "v")
	if latestVersion == sonooverlay.Version || sonooverlay.Version == "0.0.0" {
		return "", ""
	}
	return latestVersion, release.GetHTMLURL()
}

func checkSubstrings(str []string, subs ...string) string {
	for _, s := range str {
		for _, sub := range subs {
			if strings.Contains(s, sub) {
				return sub
			}
		}
	}
	return ""
}

func BanList(name string) (bool, error) {

	return false, nil
	/*
		Due to the AGPL terms and conditions, I cannot have a ban list in my fork of pjsekai-overlay-APPEND.
	*/
}

func locale() (string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-WinSystemLocale | Select-Object -ExpandProperty Name")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func langPackCheck() (string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-InstalledLanguage")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func isAdminPerm(path string) bool {
	created := false
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return true
		}
		created = true
	}

	testFile := filepath.Join(path, ".test_access")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return true
	}

	// cleanup test file
	_ = os.Remove(testFile)

	if created {
		_ = os.Remove(path)
	}

	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func origMain(isOptionSpecified bool) {
	Title()

	var aviutlType int
	flag.IntVar(&aviutlType, "aviutl-type", 0, "AviUtlインスタンスを指定します。(Specify AviUtl instance.)\n'1': AviUtl\n'2': AviUtl ExEdit2")

	var skipAviutlModConfig bool
	flag.BoolVar(&skipAviutlModConfig, "skip-mod-config", false, "AviUtlの設定変更はスキップされます。(Skip modifying AviUtl configurations.)")

	var skipAviutlInstall bool
	flag.BoolVar(&skipAviutlInstall, "skip-obj-install", false, "AviUtlオブジェクトのインストールをスキップします。(Skip installing AviUtl objects.)")

	var skipAviutlScriptInstall bool
	flag.BoolVar(&skipAviutlScriptInstall, "skip-script-install", false, "AviUtlスクリプトのインストールをスキップします。(Skip installing AviUtl scripts.)")

	var noExplorerAutoOpen bool
	flag.BoolVar(&noExplorerAutoOpen, "no-explorer-auto-open", false, "出力先ディレクトリを自動で開くのを無効にします。(Disable auto-opening output directory in Explorer.)")

	var outDir string
	flag.StringVar(&outDir, "out-dir", "./dist/_chartId_", "出力先ディレクトリを指定します。_chartId_ は譜面IDに置き換えられます。\nEnter the output path. _chartId_ will be replaced with the chart ID.")

	var chartInstance string
	flag.StringVar(&chartInstance, "instance", "", "サーバーインスタンスを指定します。(Specify the server instance.)")

	var customBG bool
	flag.BoolVar(&customBG, "custom-bg", false, "UntitledChartsでカスタム背景を使用する。(Use custom background in UntitledCharts.)")

	var scoreModeInt int
	flag.IntVar(&scoreModeInt, "score-mode", 1, "採点モードを指定します。(Specify scoring mode.)\n'1': デフォルト/Default\n'2': 大会モード/Tournament Mode (PERFECT = +3)")

	var teamPower float64
	flag.Float64Var(&teamPower, "power", 250000, "総合力を指定します。(Specify the team's power.)")

	var enUI bool
	flag.BoolVar(&enUI, "en-ui", false, "英語版を使う(イントロ + v3 UI) - Use English version (Intro + v3 UI)")

	var allFlick bool
	flag.BoolVar(&allFlick, "all-flick", false, "すべてのノーツをフリックとして扱います。(Treat all notes as flicks.)")

	flag.Usage = func() {
		fmt.Println("Usage: Sono-Overlay [オプション (Options)] [譜面ID (Chart ID)]")
		flag.PrintDefaults()
	}

	flag.Parse()

	cwd, err := os.Getwd()

	// Version Checking
	latestVer, releaseURL := checkUpdate()
	if latestVer != "" {
		fmt.Printf(color.HiCyanString("新しいバージョンがリリースされています\nNew version released: v%s -> v%s\n"), sonooverlay.Version, latestVer)
		fmt.Printf(color.HiCyanString("ダウンロード (Download Here) -> %s\n"), releaseURL)
		fmt.Println(color.RedString("\nFAIL: Sono-Overlayを最新バージョンに更新してください。\nFAIL: Please update Sono-Overlay to the latest version."))
		fmt.Println(color.RedString("This program will run, but I will not provide support for this version of Sono-Overlay.\n"))
	}

	// Select screen
	fmt.Print("\n起動モードを選択してください (Select Startup Mode):\n'1': 通常起動 [譜面解析 + ビデオ生成] (Default (Overlay))\n'2': メディアダウンローダー単体起動 (YT Download / Experimental)\n> ")
	beforeStart, _ := rawmode.Enable()
	tmpStartByte, _ := bufio.NewReader(os.Stdin).ReadByte()
	tmpStart := string(tmpStartByte)
	rawmode.Restore(beforeStart)

	if tmpStart == "2" {
		fmt.Printf("\n\033[A\033[2K\r> %s\n", color.HiCyanString("2"))
		fmt.Println(color.GreenString("Mode: Standalone Downloader Mode"))

		dlTargetDir := filepath.Join(cwd, "dist", "downloads")
		if flag.Arg(0) != "" {
			dlTargetDir = filepath.Join(cwd, "dist", flag.Arg(0))
		} else {
			fmt.Print("保存先のフォルダ名を入力してください（空欄で「downloads」）\nEnter destination folder name (Leave blank for 'downloads'):\n> ")
			var customFolderName string
			fmt.Scanln(&customFolderName)
			customFolderName = strings.TrimSpace(customFolderName)
			if customFolderName != "" {
				dlTargetDir = filepath.Join(cwd, "dist", customFolderName)
			}
		}

		if err := os.MkdirAll(dlTargetDir, 0755); err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: Failed to create target directory: %s", err.Error())))
			return
		}

		// ADDED: Friendly warning if they pick the downloader without installing files
		ytdlpCheckPath := filepath.Join(cwd, "addons", "yt-dlp.exe")
		if _, err := os.Stat(ytdlpCheckPath); os.IsNotExist(err) {
			fmt.Println(color.RedString("\nFAIL: yt-dlp.exe was not found in your 'addons' folder!"))
			fmt.Println(color.HiYellowString("Please download yt-dlp.exe and place it inside: " + filepath.Join(cwd, "addons")))

			// Let them press a key so the console window doesn't instantly vanish
			fmt.Print(color.CyanString("\nPress any key to exit..."))
			beforeErr, _ := rawmode.Enable()
			bufio.NewReader(os.Stdin).ReadByte()
			rawmode.Restore(beforeErr)
			return
		}

		tryRunYtdlpAddon(dlTargetDir, true)

		fmt.Println(color.GreenString("\n処理が終了しました。(Operation complete.)"))

		if !noExplorerAutoOpen {
			explorerCmd := exec.Command(`explorer`, `/select,`, dlTargetDir)
			explorerCmd.Run()
		}
		return // EARLY INTERCEPT EXIT
	} else if tmpStart != "1" {
		fmt.Println(color.RedString("\nFAIL: Invalid option. Valid options are: 1, 2"))

		// Let them press a key so the console window doesn't instantly vanish
		fmt.Print(color.CyanString("\nPress any key to exit..."))
		beforeErr, _ := rawmode.Enable()
		bufio.NewReader(os.Stdin).ReadByte()
		rawmode.Restore(beforeErr)
		return

	}

	// yes

	// Normal Mode Selection Fallback
	fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString("1"))
	fmt.Println(color.GreenString("Mode: Default (Overlay)"))

	// removed forced updates lol

	fmt.Printf("- 前提条件を確認中 (Checking prerequisites)... ")

	locale, err := locale()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	} else if locale != "ja-JP" {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: お使いのシステムロケールが「日本語（日本）」に設定されていません。変更方法についてはWikiを参照してください。\nYour system locale is not set to \"Japanese (Japan)\". Refer to the wiki for how to change it.\n- System locale: %v", locale)))
		fmt.Println(color.RedString(fmt.Sprintf("\nThis program will run, but the output will be unusable till you install the language locale.")))
	}

	langPackCheck, err := langPackCheck()
	if err != nil {
		fmt.Println(color.HiYellowString(fmt.Sprintf("WARN: 言語パックを確認できません。(Unable to check language pack.)\n%s", err.Error())))
	} else if !strings.Contains(langPackCheck, "ja-JP") {
		fmt.Println(color.RedString("\nFAIL: 日本語言語パックがインストールされていません。変更方法についてはWikiを参照してください。\nJapanese language pack is not installed. Refer to the wiki for how to install it."))
		fmt.Println(color.RedString(fmt.Sprintf("\nThis program will run, but the output will be unusable till you install the language pack.")))
	}

	// it still checks for JP language pack, but this is irrelevant to an EN user, and so i removed the forced JP settings

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}
	if isAdminPerm(cwd) {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: ディレクトリには管理者権限が必要です。Sono-Overlayを「C:\\」または別の場所に移動してください。\nYour directory requires administrative permissions. Please move Sono-Overlay to \"C:\\\" or somewhere else.\n\n出力先ディレクトリ (Output path): %s", cwd)))
		return
	}
	if !isASCII(cwd) {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: ディレクトリに非ASCII文字が含まれています。Sono-Overlayを「C:\\」または別の場所に移動してください。\nYour directory contains non-ASCII characters. Please move Sono-Overlay to \"C:\\\" or somewhere else.\n\n出力先ディレクトリ (Output path): %s", cwd)))
		return
	}

	mappingName, mapping := sonooverlay.SetOverlayDefault()

	if len(mapping) != 23 {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL:「default.ini」ファイルのデータに異常があります。「default.ini」ファイルを削除し、プログラムを再起動して再生成してください。\nAbnormal \"default.ini\" data. Please regenerate by deleting the \"default.ini\" file and reopen the program.\n- Mapping count: %v != 23", len(mapping))))
		return
	}

	// what is this bro :sob:
	var mappingFloat64 []float64
	for _, v := range mapping {
		v = strings.TrimRightFunc(v, func(r rune) bool {
			return strings.HasSuffix(string(r), "+") || strings.HasSuffix(string(r), "-") || strings.HasSuffix(string(r), ".") || (r < '0' || r > '9')
		})
		mappingFloat64 = append(mappingFloat64, func() float64 {
			f, _ := strconv.ParseFloat(v, 64)
			return f
		}())
	}

	var float64Pointer = func(f float64) *float64 {
		return &f
	}

	var inRange = map[string]bool{
		// Root
		"offset":      mappingFloat64[0] >= -99999.99 && mappingFloat64[0] <= 99999.99,
		"cache":       mappingFloat64[1] == 0 || mappingFloat64[1] == 1,
		"text_lang":   mappingFloat64[2] == 0 || mappingFloat64[2] == 1,
		"watermark":   mappingFloat64[3] == 0 || mappingFloat64[3] == 1,
		"detail_stat": mappingFloat64[4] == 0 || mappingFloat64[4] == 1,
		// Life
		"life":       mappingFloat64[5] >= -9999 && mappingFloat64[5] <= 9999 && math.Mod(mappingFloat64[5], 1.0) == 0,
		"life_skill": mappingFloat64[6] == 0 || mappingFloat64[6] == 1,
		"overflow":   mappingFloat64[7] == 0 || mappingFloat64[7] == 1,
		"lead_zero":  mappingFloat64[8] == 0 || mappingFloat64[8] == 1,
		// Score
		"min_digit":   mappingFloat64[9] >= 1 && mappingFloat64[9] <= 99 && math.Mod(mappingFloat64[9], 1.0) == 0,
		"score_skill": mappingFloat64[10] >= 0 && mappingFloat64[10] <= 2 && math.Mod(mappingFloat64[10], 1.0) == 0,
		"score_speed": mappingFloat64[11] >= 0,
		"anim_score":  mappingFloat64[12] == 0 || mappingFloat64[12] == 1,
		"wds_anim":    mappingFloat64[13] == 0 || mappingFloat64[13] == 1,
		// Combo
		"ap":               mappingFloat64[14] == 0 || mappingFloat64[14] == 1,
		"tag":              mappingFloat64[15] == 0 || mappingFloat64[15] == 1,
		"last_digit":       mappingFloat64[16] >= 0 && math.Mod(mappingFloat64[16], 1.0) == 0,
		"combo_speed":      mappingFloat64[17] >= 0,
		"combo_burst":      mappingFloat64[18] == 0 || mappingFloat64[18] == 1,
		"achievement_rate": float64Pointer(mappingFloat64[19]) != nil,
		// Judgement
		"judge":       mappingFloat64[20] >= 1 && mappingFloat64[20] <= 10 && math.Mod(mappingFloat64[20], 1.0) == 0,
		"judge_speed": mappingFloat64[21] >= 0,
		// Add this line to validate text strings (always true since it's a string, not a bounding float)
		"custom_watermark": true,
	}

	var mappingErr []string
	for i := range mapping {
		inRangeBool := inRange[mappingName[i]]
		if !inRangeBool {
			mappingErr = append(mappingErr, mappingName[i], fmt.Sprintf("%v", mapping[i]))
		}
	}

	if mappingErr != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL:「default.ini」ファイルのデータに異常があります。「default.ini」ファイルを削除し、プログラムを再起動して再生成してください。\nAbnormal \"default.ini\" data. Please regenerate by deleting the \"default.ini\" file and reopen the program.\n- Mapping out of range: %s", mappingErr)))
		return
	}

	var mappingStr []string
	for _, v := range mapping {
		mappingStr = append(mappingStr, fmt.Sprintf("%v", v))
	}

	fmt.Println(color.GreenString("OK"))

	var aviutlPath, aviutlProcess, aviutlName string

	/*
		This code is archived.
		AviUtl ExEdit2 is the way to go.

			switch aviutlType {
			case 1:
				aviutlProcess = "aviutl.exe"
				aviutlName = "AviUtl"
				aviutlPath, _, _ = sonooverlay.DetectAviUtl()
			case 2:
				aviutlProcess = "aviutl2.exe"
				aviutlName = "AviUtl ExEdit2"
				aviutlPath, _ = filepath.Abs("C:\\ProgramData\\aviutl2")
			default:
				aviutlPath, aviutlProcess, aviutlName = sonooverlay.DetectAviUtl()
				if aviutlProcess != "" {
					fmt.Printf("Instance (auto-detected): %s\n", color.GreenString(aviutlName))
				}

				if aviutlProcess == "" {
					fmt.Print("ファイルを生成するAviUtlインスタンスを選択してください。\nChoose AviUtl instance to generate files.\n\n'1': AviUtl\n'2': AviUtl ExEdit2\n> ")
					before, _ := rawmode.Enable()
					tmpAviutlByte, _ := bufio.NewReader(os.Stdin).ReadByte()
					tmpAviutl := string(tmpAviutlByte)
					rawmode.Restore(before)
					switch tmpAviutl {
					default:
						aviutlProcess = ""
						fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpAviutl))
						fmt.Println(color.RedString("FAIL: AviUtlインスタンスが選択されていません。\nAviUtl instance not selected."))
						return
					case "1":
						aviutlProcess = "aviutl.exe"
						aviutlName = "AviUtl"
						aviutlPath, _, _ = sonooverlay.DetectAviUtl()
						fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpAviutl))
						fmt.Println(color.GreenString("Instance: AviUtl"))
					case "2":
						aviutlProcess = "aviutl2.exe"
						aviutlName = "AviUtl ExEdit2"
						aviutlPath, _ = filepath.Abs("C:\\ProgramData\\aviutl2")
						fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpAviutl))
						fmt.Println(color.GreenString("Instance: AviUtl ExEdit2"))
					}
				}
			}
	*/

	aviutlProcess = "aviutl2.exe"
	aviutlName = "AviUtl ExEdit2"
	aviutlPath, _ = filepath.Abs("C:\\ProgramData\\aviutl2")
	fmt.Println(color.GreenString("Instance: AviUtl ExEdit2 in C:\\ProgramData\\aviutl2 (Aviutl v1 is deprecated)"))

	var successInstall = false
	if !skipAviutlModConfig {
		success := sonooverlay.ModifyAviUtlConfig(aviutlPath, aviutlProcess)
		if success {
			fmt.Println(color.GreenString(aviutlName + "の設定変更が正常に完了しました。(" + aviutlName + " configurations successfully modified.)"))
			successInstall = true
		}
	}
	if !skipAviutlInstall {
		success := sonooverlay.TryInstallObject(aviutlPath, aviutlProcess, mappingStr)
		if success {
			fmt.Println(color.GreenString(aviutlName + "オブジェクトのインストールに成功しました。(" + aviutlName + " object successfully installed.)"))
			successInstall = true
		}
	}
	if !skipAviutlScriptInstall {
		success := sonooverlay.TryInstallScript(aviutlPath, aviutlProcess)
		if success {
			fmt.Println(color.GreenString(aviutlName + "依存関係スクリプトのインストールに成功しました。(" + aviutlName + " dependency scripts successfully installed.)"))
			successInstall = true
		}
	}
	if successInstall {
		fmt.Println(color.HiYellowString("変更を適用するには、" + aviutlName + "を再起動してください。(Please restart " + aviutlName + " to apply changes.)\n"))
	}

	// randomize if a tip even appears, 50/50 chance
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(2) == 0 {
		Tips()
	}

	var chartId string
	if flag.Arg(0) != "" {
		chartId = flag.Arg(0)
		fmt.Printf("譜面ID (Chart ID): %s\n", color.GreenString(chartId))
	} else {
		var sb strings.Builder

		sb.WriteString("譜面IDを接頭辞込みで入力して下さい。")
		sb.WriteString("\nEnter the chart ID including the prefix.")
		sb.WriteString("\n\n'sss-': Sbuga's Sonolus Server (sonolus.sbuga.com)")
		sb.WriteString("\n'sekai-best-': Also try SSS (sonolus.sekai.best)")
		sb.WriteString("\n'chcy-': Chart Cyanvas (cc.milkbun.org & offshoots)")
		sb.WriteString("\n'ptlv-': Potato Leaves (ptlv.milkbun.org)")
		sb.WriteString("\n'UnCh-': UntitledCharts (untitledcharts.com)")
		sb.WriteString("\n'sync-': Local Server (ScoreSync + ScoreSync Modern)")
		sb.WriteString("\n'coconut-next-sekai-': Next SEKAI (coconut.sonolus.com/next-sekai)")
		sb.WriteString("\n'coconut-horizon-': Sonolus Horizon (coconut.sonolus.com/horizon) <-- Original Sonolus Rhythm Game")
		sb.WriteString("\n* I have obtained permission to implement Sonolus Horizon into this overlay.\n")
		sb.WriteString("\n> ")

		// Convert back to a single string when done
		result := sb.String()
		fmt.Print(result)
		fmt.Scanln(&chartId)
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartId))
	}

	// Instance section
	if chartInstance == "" && strings.HasPrefix(chartId, "chcy-") {
		fmt.Printf("\nChart Cyanvasインスタンスを選択してください。(Please choose Chart Cyanvas instance.)\n%s\n\n[インスタンス一覧 - List of instance(s)]\n'0': アーカイブ/Archive - cc.milkbun.org\n'1': 分岐サーバー/Offshoot server - chart-cyanvas.com\n> ", color.HiYellowString("(!) 別のインスタンスを持っていますか？URLドメインを入力してください。(Do you have a different instance? Input the URL domain.)"))
		var chartInput string
		fmt.Scanln(&chartInput)
		chartInput = strings.TrimPrefix(chartInput, "http://")
		chartInput = strings.TrimPrefix(chartInput, "https://")
		chartInstance = strings.Split(chartInput, "/")[0]
		fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartInput))
	}

	var chartSource sonooverlay.Source
	if strings.HasPrefix(chartId, "sync") {
		chartSource, err = sonooverlay.DetectLocalChartSource()
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
		if strings.Contains(chartId, "-") {
			parts := strings.SplitN(chartId, "-", 2)
			if len(parts) == 2 {
				chartId = parts[1]
			}
		} else {
			fmt.Print("ローカルサーバーの譜面を入力してください。(Enter chart ID for the local server.)\n> ")
			fmt.Scanln(&chartId)
		}
	} else {
		chartSource, err = sonooverlay.DetectChartSource(chartId, chartInstance)
		if err != nil {
			fmt.Println(color.RedString("FAIL: 譜面が見つかりません。接頭辞も込め、正しい譜面IDを入力して下さい。\nChart not found. Please enter the correct chart ID including the prefix."))
			return
		}
		if chartSource.Status == 1 {
			fmt.Printf(color.RedString("FAIL: %sは対応されなくなりました。ご利用ありがとうございました。\n%s is no longer supported. Thank you for using the service.\n"), chartSource.Name, chartSource.Name)
			return
		}
		if chartSource.Status == 2 {
			fmt.Printf(color.HiYellowString("WARN: %sは現在開発中であり、正常に動作しない可能性があります。\n%s is currently in development and may not work.\n"), chartSource.Name, chartSource.Name)
		}
	}

	fmt.Printf("- 譜面を取得中 (Getting chart): %s%s%s ", RgbColorEscape(chartSource.Color), chartSource.Name, ResetEscape())

	var chart sonolus.LevelInfo
	prefixTrim := checkSubstrings([]string{chartId}, "lalo-", "skyra-")
	chart, err = sonooverlay.FetchChart(chartSource, strings.TrimPrefix(chartId, prefixTrim))

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	// Additional BG
	chartCCv1, _ := sonooverlay.FetchChart(chartSource, chartId+"?c_background=v1")
	chartUNv3, _ := sonooverlay.FetchChart(chartSource, chartId+"?levelbg=v3")
	chartUNv1, _ := sonooverlay.FetchChart(chartSource, chartId+"?levelbg=v1")
	chartUNv1def, _ := sonooverlay.FetchChart(chartSource, chartId+"?levelbg=default_or_v1")

	// I will update this when it breaks.
	if chart.Engine.Version != 13 {
		fmt.Println(color.RedString(fmt.Sprintf("\nFAIL (ver.%d): エンジンのバージョンが古い。Sono-Overlayを最新版に更新してください。\nUnsupported engine version. Please update Sono-Overlay to latest version.", chart.Engine.Version)))
		return
	}

	banList, err := BanList(chart.Author)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	} else if banList {
		fmt.Println(color.RedString("\nFAIL: 申し訳ありませんが、この譜面作者／組織はこのツールの使用が禁止されています。\nSorry, this charter/organization is banned from using this tool."))
		return
	}

	fmt.Println(color.GreenString("OK"))
	fmt.Printf("  %s / %s - %s (Lv. %s)\n",
		color.CyanString(chart.Title),
		color.CyanString(chart.Artists),
		color.CyanString(chart.Author),
		color.MagentaString(strconv.Itoa(chart.Rating)),
	)

	fmt.Printf("- exeのパスを取得中 (Getting executable path)... ")
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	formattedOutDir := filepath.Join(cwd, strings.ReplaceAll(outDir, "_chartId_", chartId))
	resultDir := filepath.Dir(formattedOutDir) + "\\" + chartId

	fmt.Println(color.GreenString("OK"))
	fmt.Printf("- 出力先ディレクトリ (Output path): %s\n", color.CyanString(resultDir))

	fmt.Print("- ジャケットをダウンロード中 (Downloading jacket)... ")
	err = sonooverlay.DownloadJacket(chartSource, chart, formattedOutDir)
	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	// fmt.Print("- 音声のプレビューをダウンロード中 (Downloading preview audio)... ")
	// err = sonooverlay.DownloadPreview(chartSource, chart, formattedOutDir)
	// if err != nil {
	// 	fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
	// 	return
	// }

	// fmt.Println(color.GreenString("OK"))

	if !isOptionSpecified && (chartSource.Id == "untitledcharts" || chartSource.Id == "skyra") {
		fmt.Print("\nカスタム背景を使用しますか？（デフォルトを使用するには「n」）[y/n]\nUse custom background? ('n' to use default) [y/n]\n> ")
		before, _ := rawmode.Enable()
		tmpCustomBGByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpCustomBG := string(tmpCustomBGByte)
		rawmode.Restore(before)
		if tmpCustomBG == "Y" || tmpCustomBG == "y" {
			customBG = true
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpCustomBG))
			fmt.Println(color.GreenString("TOGGLE: ON"))
		} else {
			customBG = false
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpCustomBG))
			fmt.Println(color.RedString("TOGGLE: OFF"))
		}
	}

	if customBG {
		fmt.Print("- 背景をダウンロード中 (Downloading background)... ")

		err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId, "", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}

		if chartSource.Id == "untitledcharts" {
			err = sonooverlay.DownloadBackground(chartSource, chartUNv1def, formattedOutDir, chartId+"?levelbg=default_or_v1", "", customBG)
			if err != nil {
				fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
				return
			}
		} else {
			err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId+"/", "", customBG)
			if err != nil {
				fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
				return
			}
		}
	} else if chartSource.Id == "untitledcharts" {
		fmt.Print("- 背景をダウンロード中 (Downloading background)... ")

		err = sonooverlay.DownloadBackground(chartSource, chartUNv3, formattedOutDir, chartId+"?levelbg=v3", "", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}

		err = sonooverlay.DownloadBackground(chartSource, chartUNv1, formattedOutDir, chartId+"?levelbg=v1", "", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
	} else if chartSource.Id == "chart_cyanvas" && chartSource.Name != "Chart Cyanvas Archive" {
		fmt.Print("- 背景をダウンロード中 (Downloading background)... ")

		err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId, "", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}

		err = sonooverlay.DownloadBackground(chartSource, chartCCv1, formattedOutDir, chartId+"?c_background=v1", "", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
	} else {
		fmt.Print("- ローカルで背景を生成中 - お待ちください (Generating background locally - please wait)... ")

		err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId, "-v 1", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}

		err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId, "-v 3", customBG)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
	}

	fmt.Println(color.GreenString("OK"))

	fmt.Print("- 譜面を解析中 (Analyzing chart)... ")
	levelData, err := sonooverlay.FetchLevelData(chartSource, chart)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	var scoreMode string
	switch scoreModeInt {
	default:
		scoreMode = "default"
	case 2:
		scoreMode = "tournament"
	}
	if !isOptionSpecified {
		fmt.Print("\n採点モードを選択してください。(Choose scoring mode.)\n'1': デフォルト/Default\n'2': 大会モード/Tournament Mode (PERFECT = +3)\n> ")
		before, _ := rawmode.Enable()
		tmpScoreModeByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpScoreMode := string(tmpScoreModeByte)
		rawmode.Restore(before)
		switch tmpScoreMode {
		default:
			scoreMode = "default"
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpScoreMode))
			fmt.Println(color.GreenString("Score Mode: デフォルト/Default"))
		case "2":
			scoreMode = "tournament"
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpScoreMode))
			fmt.Println(color.GreenString("Score Mode: 大会/Tournament"))
		}
	}

	if !isOptionSpecified && scoreMode == "default" {
		fmt.Print("\n総合力を指定してください。 (Input your team power.)\n\n- 小数と科学的記数法が使える (Accepts decimals & scientific notation)\n- おすすめ (Recommended): 250000 - 300000\n- 例 (Example): 1234567; 1e+20; -300000\n> ")
		var tmpTeamPower string
		fmt.Scanln(&tmpTeamPower)
		if tmpTeamPower == "" {
			tmpTeamPower = "250000"
		}
		teamPower, err = strconv.ParseFloat(tmpTeamPower, 64)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}

		if teamPower >= math.Abs(math.Pow(2, 56)/10) && aviutlProcess == "aviutl.exe" {
			fmt.Printf("\033[A\033[2K\r> %s\n", color.HiYellowString(tmpTeamPower))
			fmt.Println(color.HiYellowString("WARN: スコアは大きすぎると精度が落ちる可能性がある。Score may decrease precision if it's too large."))
		} else {
			fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(tmpTeamPower))
		}
	}

	fmt.Print("- スコアを計算中 (Calculating score)... ")
	scoreData := sonooverlay.CalculateScore(chart, levelData, teamPower, scoreMode, allFlick)

	fmt.Println(color.GreenString("OK"))
	if !isOptionSpecified {
		fmt.Print("\n英語UIを使う？（イントロ + v3 UI）[y/n]\nUse English UI? (Intro + v3 UI) [y/n]\n> ")
		before, _ := rawmode.Enable()
		tmpEnableENByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpEnableEN := string(tmpEnableENByte)
		rawmode.Restore(before)
		if tmpEnableEN == "Y" || tmpEnableEN == "y" {
			enUI = true
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpEnableEN))
			fmt.Println(color.GreenString("TOGGLE: ON"))
		} else {
			enUI = false
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpEnableEN))
			fmt.Println(color.RedString("TOGGLE: OFF"))
		}
	}

	executableDir := filepath.Dir(executablePath)
	assets := filepath.Join(executableDir, "assets")

	fmt.Print("\n- pedファイルを生成中 (Generating ped file)... ")

	err = sonooverlay.WritePedFile(scoreData, assets, filepath.Join(formattedOutDir, "data.ped"), sonolus.LevelInfo{Rating: chart.Rating}, levelData, scoreMode, enUI)

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	fmt.Println(color.GreenString("OK"))

	var exoType = "exo"
	if aviutlProcess == "aviutl2.exe" {
		exoType = "alias(.object)"
	}

	fmt.Printf("- %sファイルを生成中 (Generating %s file)... ", exoType, exoType)

	var difficulty string
	difficultyStrings := []string{"EASY", "NORMAL", "HARD", "EXPERT", "MASTER", "APPEND", "ETERNAL"}

	for i := range chart.Tags {
		tags := checkSubstrings([]string{strings.ToUpper(chart.Tags[i].Title)}, difficultyStrings...)
		if tags != "" {
			difficulty = tags
			break
		}
	}

	if difficulty == "" {
		if title := checkSubstrings(strings.Fields(strings.ToUpper(chart.Title)), difficultyStrings...); title != "" {
			difficulty = title
		} else {
			difficulty = "APPEND"
		}
	}

	composerAndVocals := []string{chart.Artists, "-"}
	if separateAttempt := strings.Split(chart.Artists, " / "); chartSource.Id == "chart_cyanvas" && len(separateAttempt) <= 2 {
		composerAndVocals = separateAttempt
	}

	charter := []string{chart.Author, "-"}
	if charterTag := strings.Split(chart.Author, "#"); len(charterTag) <= 2 {
		charter = charterTag
	}

	description := []string{fmt.Sprintf("作詞：-    作曲：%s    編曲：-", composerAndVocals[0]), fmt.Sprintf("Vo：%s    譜面制作：%s", composerAndVocals[1], charter[0])}
	descriptionv1 := []string{fmt.Sprintf("作詞：-    作曲：%s    編曲：-", composerAndVocals[0]), fmt.Sprintf("歌：%s    譜面制作：%s", composerAndVocals[1], charter[0])}
	extra := "【追加情報】"
	exFile := "tournament-mode.png"
	exFileOpacity := "100.0"

	if enUI {
		description = []string{fmt.Sprintf("Lyrics: -    Music: %s    Arrangement: -", composerAndVocals[0]), fmt.Sprintf("Vo: %s    Chart Design: %s", composerAndVocals[1], charter[0])}
		descriptionv1 = []string{fmt.Sprintf("Lyrics: -    Music: %s    Arrangement: -", composerAndVocals[0]), fmt.Sprintf("Vocals: %s    Chart Design: %s", composerAndVocals[1], charter[0])}
		extra = "【Additional Info】"
		exFile = "tournament-mode-en.png"
	}
	if scoreMode == "tournament" {
		exFileOpacity = "0.0"
	}

	if aviutlProcess == "aviutl.exe" {
		err = sonooverlay.WriteExoFiles(assets, formattedOutDir, chart.Title, description, descriptionv1, difficulty, extra, exFile, exFileOpacity, mappingStr)
	} else {
		err = sonooverlay.WriteAliasFiles(assets, formattedOutDir, chart.Title, description, descriptionv1, difficulty, extra, exFile, exFileOpacity, mappingStr)
	}

	if err != nil {
		fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
		return
	}

	tryRunYtdlpAddon(resultDir, false)

	message := fmt.Sprintf("\n全ての処理が完了しました！READMEの規約を確認した上で、%sファイルを%sにインポートして下さい。\nExecution complete! Please import the %s file into %s after reviewing the README Terms of Use.", exoType, aviutlName, exoType, aviutlName)
	fmt.Println(color.GreenString(message))

	if !isOptionSpecified || !noExplorerAutoOpen {
		cmd := exec.Command(`explorer`, `/select,`, resultDir)
		cmd.Run()

		time.Sleep(2000 * time.Millisecond)
	}
}

func tryRunYtdlpAddon(resultDir string, isStandalone bool) {
	executablePath, err := os.Executable()
	if err != nil {
		return
	}
	cwd := filepath.Dir(executablePath)

	ytdlpPath := filepath.Join(cwd, "addons", "yt-dlp.exe")

	if _, err := os.Stat(ytdlpPath); os.IsNotExist(err) {
		return
	}

	fmt.Println(color.HiMagentaString("\n[Addon] yt-dlp addon detected."))
	if !isStandalone {
		fmt.Print("オーディオ／動画ソースをダウンロードしますか？ (Do you want to download an audio/video source?) [y/n]\n> ")

		before, _ := rawmode.Enable()
		tmpChoiceByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpChoice := string(tmpChoiceByte)
		rawmode.Restore(before)

		if tmpChoice != "Y" && tmpChoice != "y" {
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.RedString(tmpChoice))
			return
		}
		fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString(tmpChoice))
	}

	// Loop
	for {
		fmt.Print("\nYouTube などの動画URLを入力してください。('exit' で終了)\nEnter the media source URL (or type 'exit' to quit):\n> ")
		var videoURL string
		fmt.Scanln(&videoURL)
		videoURL = strings.TrimSpace(videoURL)

		// Clean exit door out of the downloader thread loop
		if strings.ToLower(videoURL) == "exit" || videoURL == "" {
			fmt.Println(color.HiYellowString("Exiting downloader menu."))
			break
		}

		fmt.Print("ダウンロード形式を選択してください (Choose download format):\n'1': 楽曲音声のみ (Audio Only - mp3)\n'2': 背景動画 (Background Video - mp4)\n> ")
		beforeMode, _ := rawmode.Enable()
		tmpModeByte, _ := bufio.NewReader(os.Stdin).ReadByte()
		tmpMode := string(tmpModeByte)
		rawmode.Restore(beforeMode)

		var cmd *exec.Cmd
		if tmpMode == "2" {
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString("Video (mp4)"))

			// Dynamic naming rule definition
			fileName := "video.%(ext)s"
			if isStandalone {
				// Standalone Mode: Sets name to "MV_Title_UnixTimestamp.mp4"
				// --restrict-filenames automatically cleans up spaces/special characters for AviUtl stability
				fileName = fmt.Sprintf("%%(title)s_%d.%%(ext)s", time.Now().Unix())
			}
			outputTemplate := filepath.Join(resultDir, fileName)

			cmd = exec.Command(ytdlpPath,
				"--ffmpeg-location", filepath.Join(cwd, "addons"),
				"--restrict-filenames", // ◄ ADDED: Safely cleans up video titles for Windows/AviUtl compatibility
				"-f", "bv*[height<=1080][ext=mp4]+ba[ext=m4a]/b[height<=1080][ext=mp4]",
				"--merge-output-format", "mp4",
				"-o", outputTemplate,
				videoURL,
			)
		} else {
			fmt.Printf("\n\033[A\033[2K\r> %s\n", color.GreenString("Audio (mp3)"))

			fileName := "audio.%(ext)s"
			if isStandalone {
				// Standalone Mode: Sets name to "Audio_Title_UnixTimestamp.mp3"
				fileName = fmt.Sprintf("%%(title)s_%d.%%(ext)s", time.Now().Unix())
			}
			outputTemplate := filepath.Join(resultDir, fileName)

			cmd = exec.Command(ytdlpPath,
				"--ffmpeg-location", filepath.Join(cwd, "addons"),
				"--restrict-filenames", // ◄ ADDED: Safely cleans up audio titles
				"-f", "ba/b",
				"-x",
				"--audio-format", "mp3",
				"-o", outputTemplate,
				videoURL,
			)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: Download execution failed: %s", err.Error())))
		} else {
			fmt.Println(color.GreenString("OK: Media assets successfully saved to target folder!"))
		}

		// If running inside Option 1 (Normal Mode), we only want one file for the chart project, so exit after one run
		if !isStandalone {
			break
		}
	}
}

func main() {
	isOptionSpecified := len(os.Args) > 1
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	origMain(isOptionSpecified)

	if !isOptionSpecified {
		fmt.Print(color.CyanString("\n- 何かキーを押すと終了します...\n- Press any key to exit..."))

		before, _ := rawmode.Enable()
		bufio.NewReader(os.Stdin).ReadByte()
		rawmode.Restore(before)
	}
}
