package render

import "testing"

func TestFuriganaFilters(t *testing.T) {
	const expr = "日本語[にほんご]を 勉強[べんきょう]する"
	cases := []struct {
		name string
		fn   func(string) string
		in   string
		want string
	}{
		{"ruby text (spec)", furigana, expr, "<ruby>日本語<rt>にほんご</rt></ruby>を<ruby>勉強<rt>べんきょう</rt></ruby>する"},
		{"kana only (spec)", kana, expr, "にほんごをべんきょうする"},
		{"kanji only (spec)", kanji, expr, "日本語を勉強する"},
		{"sound tag untouched (spec)", furigana, "猫[ねこ] [sound:neko.mp3]", "<ruby>猫<rt>ねこ</rt></ruby> [sound:neko.mp3]"},
		{"sound tag directly after a base", kana, "猫[sound:neko.mp3]", "猫[sound:neko.mp3]"},
		{"tags around the base", furigana, "<b>日本[にほん]</b>", "<b><ruby>日本<rt>にほん</rt></ruby></b>"},
		{"base starts after a tag", kanji, "<i>x</i>猫[ねこ]", "<i>x</i>猫"},
		{"nbsp counts as a space", kana, "を&nbsp;勉強[べんきょう]", "をべんきょう"},
		{"text without readings", furigana, "<b>日本語</b> です", "<b>日本語</b> です"},
		{"text without readings, kana", kana, "plain text", "plain text"},
		{"text without readings, kanji", kanji, "", ""},
	}
	for _, c := range cases {
		if got := c.fn(c.in); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}
