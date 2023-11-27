package serr_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"gerardus/serr"
)

var Xs = strings.Repeat("-", 256)

func TestExcerpt(t *testing.T) {
	var tests = []struct {
		name   string
		source string
		length int
		want   string
	}{
		{
			source: "ABCDEFGHIJ",
			want:   "ABCDEFGHIJ",
			length: 13,
		},
		{
			source: "ABCDEFGHIJ",
			want:   fmt.Sprintf("%s%s%s", "ABC", serr.EllipsisRune, "HIJ"),
			length: 7,
		},
		{
			source: "ABCDEFGHIJ",
			want:   fmt.Sprintf("%s%s%s", "A", serr.EllipsisRune, "J"),
			length: 3,
		},
	}
	for _, test := range tests {
		t.Run(strconv.Itoa(test.length), func(t *testing.T) {
			got := serr.Excerpt(test.source, test.length)
			wantLen := min(test.length, len(test.source))
			gotLen := utf8.RuneCountInString(got)
			if wantLen != gotLen {
				t.Errorf("Result did not return expected length\n\t\twant=%d\n\t\t got=%d",
					wantLen,
					gotLen,
				)
			}
			if test.want != got {
				t.Errorf("Result not equal\n\t\twant=%s\n\t\t got=%s",
					test.want,
					got,
				)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	var tests = []struct {
		name             string
		source1, source2 string
		want1, want2     string
		n                int
	}{
		{
			name:    "Complex, Even; Shorter and Longer",
			source1: Xs[:40] + "ABCD" + Xs[:55] + "EFGH" + Xs[:40],
			source2: Xs[:40] + "PQR" + Xs[:60] + "STUVW" + Xs[:60] + "XYZ" + Xs[:40],
			want1:   fmt.Sprintf("ABCD%sEFGH", Xs[:55]),
			want2:   fmt.Sprintf("PQR%s%s%sXYZ", Xs[:47], serr.EllipsisRune, Xs[:46]),
			n:       100,
		},
		{
			name:    "Complex, Odd; Shorter and Longer",
			source1: Xs[:40] + "ABCD" + Xs[:55] + "EFGH" + Xs[:40],
			source2: Xs[:40] + "PQR" + Xs[:60] + "STUVW" + Xs[:60] + "XYZ" + Xs[:40],
			want1:   fmt.Sprintf("ABCD%sEFGH", Xs[:55]),
			want2:   fmt.Sprintf("PQR%s%s%sXYZ", Xs[:47], serr.EllipsisRune, Xs[:47]),
			n:       101,
		},
		{
			name:    "Mismatched lengths, one diff shorter than n, diff in middle",
			source1: Xs[:50] + "ABCDEFGHI" + Xs[:50],
			source2: Xs[:50] + "XYZ" + Xs[:50],
			want1:   fmt.Sprintf("AB%sHI", serr.EllipsisRune),
			want2:   "XYZ",
			n:       5,
		},
		{
			name:    "n is shorter than source, diff in middle",
			source1: Xs[:50] + "ABCDEFGHI" + Xs[:50],
			source2: Xs[:50] + "QRSTUVXYZ" + Xs[:50],
			want1:   fmt.Sprintf("AB%sHI", serr.EllipsisRune),
			want2:   fmt.Sprintf("QR%sYZ", serr.EllipsisRune),
			n:       5,
		},
		{
			name:    "n is shorter than source, starting diff",
			source1: "ABCDEFGHI" + Xs[:50],
			source2: "QRSTUVXYZ" + Xs[:50],
			want1:   fmt.Sprintf("AB%sHI", serr.EllipsisRune),
			want2:   fmt.Sprintf("QR%sYZ", serr.EllipsisRune),
			n:       5,
		},
		{
			name:    "n is shorter than source, ending diff",
			source1: Xs[:50] + "ABCDEFGHI",
			source2: Xs[:50] + "QRSTUVXYZ",
			want1:   fmt.Sprintf("AB%sHI", serr.EllipsisRune),
			want2:   fmt.Sprintf("QR%sYZ", serr.EllipsisRune),
			n:       5,
		},
		{
			name:    "Shorter than n, diff in middle",
			source1: Xs[:5] + "ABC" + Xs[:5],
			source2: Xs[:5] + "XYZ" + Xs[:5],
			want1:   "ABC",
			want2:   "XYZ",
			n:       25,
		},
		{
			name:    "Longer than n, diff in middle",
			source1: Xs[:50] + "ABC" + Xs[:50],
			source2: Xs[:50] + "XYZ" + Xs[:50],
			want1:   "ABC",
			want2:   "XYZ",
			n:       25,
		},
		{
			name:    "Shorter than n, start diff",
			source1: "ABC" + Xs[:10],
			source2: "XYZ" + Xs[:10],
			want1:   "ABC",
			want2:   "XYZ",
			n:       25,
		},
		{
			name:    "Longer than n, start diff",
			source1: "ABC" + Xs[:50],
			source2: "XYZ" + Xs[:50],
			want1:   "ABC",
			want2:   "XYZ",
			n:       25,
		},
		{
			name:    "Shorter than n, end diff",
			source1: Xs[:10] + "ABC",
			source2: Xs[:10] + "XYZ",
			want1:   "ABC",
			want2:   "XYZ",
			n:       25,
		},
		{
			name:    "Longer than n, end diff",
			source1: Xs[:50] + "ABC",
			source2: Xs[:50] + "XYZ",
			want1:   "ABC",
			want2:   "XYZ",
			n:       25,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got1, got2, _, _ := serr.Diff(test.source1, test.source2, test.n)
			verifyDiffResult(t, 1, test.source1, test.want1, got1)
			verifyDiffResult(t, 2, test.source2, test.want2, got2)
		})
	}
}

func verifyDiffResult(t *testing.T, n int, source, want, got string) {
	wantLen := utf8.RuneCountInString(want)
	gotLen := utf8.RuneCountInString(got)
	if wantLen != gotLen {
		t.Errorf("Length mismatch [source %d]\n\twant=%d\n\t got=%d",
			n,
			wantLen,
			gotLen,
		)
	}
	if want != got {
		t.Errorf("Value mismatch [source %d]\n\twant=%s\n\t got=%s",
			n,
			want,
			got,
		)
	}
}
