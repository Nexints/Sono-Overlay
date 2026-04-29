package main

import (
	"fmt"
	"strings"

	"github.com/Nexints/pjsekai-overlay-APPEND-maintenance/pkg/pjsekaioverlay"
	"github.com/lithammer/dedent"
)

func Title() {
	fmt.Printf(
		strings.TrimSpace(dedent.Dedent(`
    %s-- Sonolus %sOverlay %s -----------------------------------------------------------%s
    %sOverlay Software used to add MVs and custom skins to Sonolus%s

        Version: %s%s%s*
            * This does not use any Sekai assets.
            * This tool is primarily intended for English users.

        Developed by %s名無し｡(@sevenc-nanashi)%s
            Link redacted for legal reasons.
        Forked by %sTootieJin & ぴぃまん(@Piliman22)%s
            Link redacted for legal reasons.
        Maintenance Fork by %sNexint%s
         -> https://github.com/Nexints/Sono-Overlay %s(In use)%s

    %s[INFO] This tool does NOT use Project Sekai assets.%s
    %sNo documentation will be provided.%s
    %sI am not affiliated with the original developer in any way, shape or form.%s
	
    %s[CAUTION] This tool is primarily only for people with technical know-how and basic knowledge of AviUtl / AviUtl ExEdit2.%s 
    %sIf you have any questions/problems, please make a discussion thread. Refer to the wiki for how to set it up.%s
    %s-------------------------------------------------------------------------------------%s
    `))+"\n",
		RgbColorEscape(0x00afc7), RgbColorEscape(0xab93ff), RgbColorEscape(0x00afc7), ResetEscape(),
		RgbColorEscape(0x00afc7), ResetEscape(),
		RgbColorEscape(0x0f6ea3), pjsekaioverlay.Version, ResetEscape(),
		RgbColorEscape(0x48b0d5), ResetEscape(),
		RgbColorEscape(0x48b0d5), ResetEscape(),
		RgbColorEscape(0x48b0d5), ResetEscape(),
		RgbColorEscape(0xadff2f), ResetEscape(),
		RgbColorEscape(0x00ff00), ResetEscape(),
		RgbColorEscape(0x00ff00), ResetEscape(),
		RgbColorEscape(0x00ff00), ResetEscape(),
		RgbColorEscape(0xff0000), ResetEscape(),
		RgbColorEscape(0xff0000), ResetEscape(),
		RgbColorEscape(0xff5a91), ResetEscape(),
	)

}

func RgbColorEscape(rgb int) string {
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", (rgb>>16)&0xff, (rgb>>8)&0xff, rgb&0xff)
}

func AnsiColorEscape(color int) string {
	return fmt.Sprintf("\033[38;5;%dm", color)
}

func ResetEscape() string {
	return "\033[0m"
}
