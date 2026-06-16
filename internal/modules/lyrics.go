package modules

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/Laky-64/gologging"
)

func LyricsCommand(ctx Context) {
	if len(ctx.Args) == 0 {
		ctx.Reply("🎤 Usage: /lyrics <song name>")
		return
	}
	song := strings.Join(ctx.Args, " ")
	url := fmt.Sprintf("https://api.lyrics.ovh/v1/%s", song)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		ctx.Reply("❌ Couldn't find lyrics for that song.")
		return
	}
	defer resp.Body.Close()
	var result struct { Lyrics string `json:"lyrics"` }
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		ctx.Reply("❌ Error parsing lyrics.")
		return
	}
	if result.Lyrics == "" {
		ctx.Reply("No lyrics found.")
		return
	}
	// Send in chunks if too long
	ctx.Reply(fmt.Sprintf("📖 *Lyrics for %s*\n\n%s", song, result.Lyrics[:min(4096, len(result.Lyrics))]))
}
