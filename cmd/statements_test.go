package cmd

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dimboknv/p24"
	"github.com/stretchr/testify/require"
)

func Test_SplitStatementsDateRange(t *testing.T) {
	cases := []struct {
		startDate time.Time
		endDate   time.Time
		card      string
		expected  []p24.StatementsOpts
	}{
		{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
			time.Date(2000, 2, 1, 0, 0, 0, 0, time.Local),
			"1111111111111112",
			[]p24.StatementsOpts{
				{
					StartDate:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
					EndDate:    time.Date(2000, 2, 1, 0, 0, 0, 0, time.Local),
					CardNumber: "1111111111111112",
				},
			},
		},
		{
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
			time.Date(2000, 9, 1, 0, 0, 0, 0, time.Local),
			"1111111111111113",
			[]p24.StatementsOpts{
				{
					StartDate:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
					EndDate:    time.Date(2000, 3, 31, 0, 0, 0, 0, time.Local),
					CardNumber: "1111111111111113",
				},
				{
					StartDate:  time.Date(2000, 4, 1, 0, 0, 0, 0, time.Local),
					EndDate:    time.Date(2000, 6, 30, 0, 0, 0, 0, time.Local),
					CardNumber: "1111111111111113",
				},
				{
					StartDate:  time.Date(2000, 7, 1, 0, 0, 0, 0, time.Local),
					EndDate:    time.Date(2000, 9, 1, 0, 0, 0, 0, time.Local),
					CardNumber: "1111111111111113",
				},
			},
		},
		{
			time.Date(2001, 1, 1, 0, 0, 0, 0, time.Local),
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
			"1111111111111114",
			[]p24.StatementsOpts{
				{
					StartDate:  time.Date(2001, 1, 1, 0, 0, 0, 0, time.Local),
					EndDate:    time.Date(2000, 1, 1, 0, 0, 0, 0, time.Local),
					CardNumber: "1111111111111114",
				},
			},
		},
	}

	l := inputTimeLayout
	check := func(expected, actual p24.StatementsOpts) {
		require.Equal(t, expected.StartDate.Format(l), actual.StartDate.Format(l), "startDates not equal")
		require.Equal(t, expected.EndDate.Format(l), actual.EndDate.Format(l), "endDates not equal")
		require.Equal(t, expected.CardNumber, actual.CardNumber, "Cards not equal")
	}
	for i, c := range cases {
		c := c
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := SplitStatementsDateRange(c.startDate, c.endDate, c.card)
			require.Equal(t, len(c.expected), len(actual), "invalid statements len")

			for k := range actual {
				k := k
				t.Run(fmt.Sprintf("%d", k), func(t *testing.T) {
					check(c.expected[k], actual[k])
				})
			}
		})
	}
}
