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
	"unsafe"

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

	// 1. Set up a native Windows Job Object to group child windows together
	job, err := windows.CreateJobObject(nil, nil)
	if err == nil {
		// Configure the Job Object to automatically force-kill all children when the handle closes
		info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
			BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
				LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
			},
		}
		_, _ = windows.SetInformationJobObject(
			job,
			windows.JobObjectExtendedLimitInformation,
			uintptr(unsafe.Pointer(&info)),
			uint32(unsafe.Sizeof(info)),
		)
		// Ensure the entire job container is destroyed when origMain exits
		defer windows.CloseHandle(job)
	}

	// 2. Execute the start command to open your pop-up window
	serverCmd := exec.Command("cmd", "/C", "start", "Sono-Server Launcher", "node", "../server/index.js")

	// Crucial: Use CREATE_BREAKAWAY_FROM_JOB so 'start' can spawn its new window hierarchy cleanly
	serverCmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_BREAKAWAY_FROM_JOB,
	}

	if err := serverCmd.Start(); err != nil {
		fmt.Println(color.RedString("\nFAIL: Could not auto-boot Sono-Server: %s", err.Error()))
	} else {
		fmt.Println(color.GreenString("\n[Process Manager] Success: Opened terminal window running Sono-Server."))

		// 3. FIX: Change .Id to .Pid so Go can correctly pull the Process ID integer
		if job != 0 {
			handle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(serverCmd.Process.Pid))
			if err == nil {
				_ = windows.AssignProcessToJobObject(job, handle)
				defer windows.CloseHandle(handle) // Close the temporary process handle safely
			}
		}
	}

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
	var isLocalJson bool // ◄ ADDED: Track if we are running in local offline mode

	if flag.Arg(0) != "" {
		chartId = flag.Arg(0)
		chartId = strings.Trim(strings.TrimSpace(chartId), "\"'") // Clean quote tags
		if info, err := os.Stat(chartId); err == nil && !info.IsDir() {
			isLocalJson = true
		}
		if isLocalJson {
			fmt.Println(color.GreenString("Targeting Local File: " + filepath.Base(chartId)))
		} else {
			fmt.Printf("譜面ID (Chart ID): %s\n", color.GreenString(chartId))
		}
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
		sb.WriteString("\n'local-': Local Server (Sono-Utils) - Append the \"local-\" tag to the existing chart ID")
		sb.WriteString("\n'coconut-next-sekai-': Next SEKAI (coconut.sonolus.com/next-sekai)")
		sb.WriteString("\n'coconut-horizon-': Sonolus Horizon (coconut.sonolus.com/horizon) <-- Original Sonolus Rhythm Game")
		sb.WriteString("\n(EXPERIMENTAL) Alternatively, drag & drop a LevelData (.json.gz) file!\n")
		sb.WriteString("\n> ")

		// Convert back to a single string when done
		result := sb.String()
		fmt.Print(result)

		// FIXED: Use bufio Scanner to read the entire line, spaces included!
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			chartId = scanner.Text()
		}

		// Clean up trailing/leading spaces or quotation mark artifacts appended by Windows drag-and-drop
		chartId = strings.Trim(strings.TrimSpace(chartId), "\"'")

		// Check if the input path string points to a real local file
		if info, err := os.Stat(chartId); err == nil && !info.IsDir() {
			isLocalJson = true
		}

		if isLocalJson {
			fmt.Printf("\033[A\033[2K\r> Targeting Local File: %s\n", color.GreenString(filepath.Base(chartId)))
		} else {
			fmt.Printf("\033[A\033[2K\r> %s\n", color.GreenString(chartId))
		}
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
	var chart sonolus.LevelInfo

	if isLocalJson {
		// Mock a local source profile to route correctly inside the pkg library
		absChartPath, err := filepath.Abs(chartId)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: Failed to resolve absolute path: %s", err.Error())))
			return
		}
		chartId = absChartPath // Update chartId with the absolute path string

		// Mock a local source profile to route correctly inside the pkg library
		chartSource = sonooverlay.Source{
			Id:     "local_json",
			Name:   "Local Offline JSON Chart",
			Color:  0xeeaa00,
			Host:   "local_disk",
			Status: 0,
		}

		// Run your overridden internal FetchChart module
		chart, err = sonooverlay.FetchChart(chartSource, chartId)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println(color.HiCyanString("\n[Local Setup] Please configure your chart metadata tags:"))

		fmt.Print("曲名を入力してください（空欄でファイル名を使用）\nEnter Song Title (Leave blank to use filename):\n> ")
		if scanner.Scan() {
			inputTitle := strings.TrimSpace(scanner.Text())
			if inputTitle != "" {
				chart.Title = inputTitle
			}
		}

		fmt.Print("アーティスト名を入力してください (Enter Music Artist / Composer):\n> ")
		if scanner.Scan() {
			inputArtist := strings.TrimSpace(scanner.Text())
			if inputArtist != "" {
				chart.Artists = inputArtist
			} else {
				chart.Artists = "Unknown Artist"
			}
		}

		fmt.Print("譜面制作者名を入力してください (Enter Chart Author / Charter):\n> ")
		if scanner.Scan() {
			inputAuthor := strings.TrimSpace(scanner.Text())
			if inputAuthor != "" {
				chart.Author = inputAuthor
			} else {
				chart.Author = "Unknown Charter"
			}
		}

		fmt.Print("譜面の難易度（数字）を入力してください (Enter Chart Difficulty Level Number):\n> ")
		if scanner.Scan() {
			inputRating := strings.TrimSpace(scanner.Text())
			if val, err := strconv.Atoi(inputRating); err == nil {
				chart.Rating = val
			} else {
				chart.Rating = 26
			}
		}

		fmt.Print("\n背景画像の設定を選択してください (Select Background Option):\n'1': ジャケットからプロセカ風背景を自動生成 (Generate Project Sekai background from cover)\n'2': カスタム画像を自分で指定 (Import your own custom image file)\n'3': 設定なし (Skip / Keep transparent or default)\n> ")
		if scanner.Scan() {
			bgChoice := strings.TrimSpace(scanner.Text())
			if bgChoice == "1" {
				customBG = false // Triggers local generator pipeline down below
			} else if bgChoice == "2" {
				customBG = true // Switches pipeline path to flag custom image loading loops
			} else {
				chartSource.Id = "local_json_no_bg" // Set custom identity to skip all backgrounds entirely
			}
		}
	} else {
		// Classic Web Domain Source Mapping Loop
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
			chartId = strings.TrimPrefix(chartId, "local-")
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

		prefixTrim := checkSubstrings([]string{chartId}, "lalo-", "skyra-")
		chart, err = sonooverlay.FetchChart(chartSource, strings.TrimPrefix(chartId, prefixTrim))
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
	}

	// Additional BG
	var chartCCv1, chartUNv3, chartUNv1, chartUNv1def sonolus.LevelInfo
	if !isLocalJson {
		chartCCv1, _ = sonooverlay.FetchChart(chartSource, chartId+"?c_background=v1")
		chartUNv3, _ = sonooverlay.FetchChart(chartSource, chartId+"?levelbg=v3")
		chartUNv1, _ = sonooverlay.FetchChart(chartSource, chartId+"?levelbg=v1")
		chartUNv1def, _ = sonooverlay.FetchChart(chartSource, chartId+"?levelbg=default_or_v1")

		if chart.Engine.Version != 13 {
			fmt.Println(color.RedString(fmt.Sprintf("\nFAIL (ver.%d): エンジンのバージョンが古い。Sono-Overlayを最新版に更新してください。\nUnsupported engine version. Please update Sono-Overlay to latest version.", chart.Engine.Version)))
			return
		}
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
	if strings.HasPrefix(chartSource.Id, "local_json") {
		cleanBaseName := strings.TrimSuffix(filepath.Base(chartId), filepath.Ext(chartId))
		formattedOutDir = filepath.Join(cwd, strings.ReplaceAll(outDir, "_chartId_", cleanBaseName))
	}
	resultDir := filepath.Dir(formattedOutDir) + "\\" + filepath.Base(formattedOutDir)

	fmt.Println(color.GreenString("OK"))
	fmt.Printf("- 出力先ディレクトリ (Output path): %s\n", color.CyanString(resultDir))

	if isLocalJson {
		os.MkdirAll(formattedOutDir, 0755)
		fmt.Println(color.HiYellowString("[Notice] Local offline mode active: Skipping remote download pipelines."))

		scanner := bufio.NewScanner(os.Stdin)

		if chartSource.Id == "local_json_no_bg" {
			fmt.Println(color.HiYellowString("Background generation skipped."))
		} else if customBG {
			fmt.Print("\nカスタム背景画像 (.png / .jpg) をここにドラッグ＆ドロップしてください:\nDrag & drop your custom background image file here:\n> ")
			if scanner.Scan() {
				bgPath := strings.Trim(strings.TrimSpace(scanner.Text()), "\"'")
				if bgPath != "" {
					if info, err := os.Stat(bgPath); err == nil && !info.IsDir() {
						fmt.Print("- 背景画像を処理中 (Processing custom background)... ")
						err = sonooverlay.CopyFile(bgPath, filepath.Join(formattedOutDir, "background.png"))
						if err != nil {
							fmt.Println(color.RedString(fmt.Sprintf("WARN: Background import failed: %s", err.Error())))
						} else {
							_ = sonooverlay.CopyFile(bgPath, filepath.Join(formattedOutDir, "background-v1.png"))
							fmt.Println(color.GreenString("OK"))
						}
					} else {
						fmt.Println(color.HiYellowString("Image file not found. Running with default blank assets instead."))
					}
				}
			}
		} else {
			fmt.Print("\nプロセカ風背景の生成に使用するジャケット画像 (.png / .jpg) をドラッグ＆ドロップしてください:\nDrag & drop a jacket image here to generate the Project Sekai background:\n> ")
			if scanner.Scan() {
				coverSrcPath := strings.Trim(strings.TrimSpace(scanner.Text()), "\"'")
				if coverSrcPath != "" {
					if info, err := os.Stat(coverSrcPath); err == nil && !info.IsDir() {
						_ = sonooverlay.CopyFile(coverSrcPath, filepath.Join(formattedOutDir, "cover.png"))

						fmt.Print("- プロセカ風背景を生成中 - お待ちください (Generating background locally - please wait)... ")

						err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId, "-v 1", customBG)
						if err != nil {
							fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: %s", err.Error())))
							return
						}

						err = sonooverlay.DownloadBackground(chartSource, chart, formattedOutDir, chartId, "-v 3", customBG)
						if err != nil {
							fmt.Println(color.RedString(fmt.Sprintf("\nFAIL: %s", err.Error())))
							return
						}
						fmt.Println(color.GreenString("OK"))
					} else {
						fmt.Println(color.RedString("FAIL: Jacket file not found. Skipping automatic background generation."))
					}
				}
			}
		}
	} else {
		// Encapsulate the entire classic web asset download loop inside this clean else block
		fmt.Print("- ジャケットをダウンロード中 (Downloading jacket)... ")
		err = sonooverlay.DownloadJacket(chartSource, chart, formattedOutDir)
		if err != nil {
			fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
			return
		}
		fmt.Println(color.GreenString("OK"))

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
	}

	// fmt.Print("- 音声のプレビューをダウンロード中 (Downloading preview audio)... ")
	// err = sonooverlay.DownloadPreview(chartSource, chart, formattedOutDir)
	// if err != nil {
	// 	fmt.Println(color.RedString(fmt.Sprintf("FAIL: %s", err.Error())))
	// 	return
	// }

	// fmt.Println(color.GreenString("OK"))

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

	isEnglishTarget := enUI || (len(mappingStr) > 2 && mappingStr[2] == "1")

	description := []string{fmt.Sprintf("作詞：-    作曲：%s    編曲：-", composerAndVocals[0]), fmt.Sprintf("Vo：%s    譜面制作：%s", composerAndVocals[1], charter[0])}
	descriptionv1 := []string{fmt.Sprintf("作詞：-    作曲：%s    編曲：-", composerAndVocals[0]), fmt.Sprintf("歌：%s    譜面制作：%s", composerAndVocals[1], charter[0])}
	extra := "【追加情報】"
	exFile := "tournament-mode.png"
	exFileOpacity := "100.0"

	if isEnglishTarget {
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

	message := fmt.Sprintf("\n全ての処理が完了しました！READMEの規約を確認した上で、%sファイルを%sにインポートして下さい。\nExecution complete! Please import the %s file into %s after reviewing the README Terms of Use.", exoType, aviutlName, exoType, aviutlName)
	fmt.Println(color.GreenString(message))

	if !isOptionSpecified || !noExplorerAutoOpen {
		cmd := exec.Command(`explorer`, `/select,`, resultDir)
		cmd.Run()

		time.Sleep(2000 * time.Millisecond)
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
