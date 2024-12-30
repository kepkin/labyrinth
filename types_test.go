package labyrinth

import "testing"

func TestGetXFromLetterMust(t *testing.T) {

	tests := []struct {
		name string
		l    rune
		want int
	}{
		{
			name: "latin a",
			l:    'a',
			want: 0,
		},
		{
			name: "latin b",
			l:    'b',
			want: 1,
		},
		{
			name: "latin c",
			l:    'c',
			want: 2,
		},
		{
			name: "latin z",
			l:    'z',
			want: 25,
		},
		{
			name: "cyrillic A",
			l:    'A',
			want: 0,
		},
		{
			name: "cyrillic C",
			l:    'C',
			want: 2,
		},
		{
			name: "cyrillic Я",
			l:    'Я',
			want: 31,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetXFromLetterMust(tt.l); got != tt.want {
				t.Errorf("GetXFromLetterMust() = %v, want %v", got, tt.want)
			}
		})
	}
}
