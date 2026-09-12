package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

// これを見て何か追加したいTipsがあれば、PRを送ってください
// if you see this and you have some tips you would like to add, make a PR pls

func Tips() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	start := time.Date(2026, 04, 27, 14, 29, 30, 0, time.UTC)
	now := time.Now()
	duration := now.Sub(start).Truncate(time.Second)
	tips := []string{
		fmt.Sprintf("Sono-Overlayの最初のコミットからの経過時間: %v", duration),
		fmt.Sprintf("Time since the first commit of Sono-Overlay: %v", duration),

		"このツールは、主に英語圏のユーザーのみを対象としています。",
		"This tool is primarily intended for English Users Only.",

		"このテキストを日本語に翻訳するのにDeepLを使っています！",
		"I use DeepL to translate this text to Japanese!",
	}

	// これを見て何か追加したいTipsがあれば、PRを送ってください
	// if you see this and you have some tips you would like to add, make a PR pls

	a := r.Intn(len(tips) - 1)
	if a%2 != 0 {
		a--
	}
	fmt.Printf(color.CyanString("◆ Tip: %s\n◆ Tip: %s\n\n"), tips[a], tips[a+1])
}
