package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_displayBuild(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		wantV, wantD, wantC := "1", "1", "1"

		gotV, gotD, gotC := displayBuild(wantV, wantD, wantC)

		require.Equal(t, wantV, gotV)
		require.Equal(t, wantD, gotD)
		require.Equal(t, wantC, gotC)
	})

	t.Run("test N/A", func(t *testing.T) {
		wantV, wantD, wantC := "", "", ""

		gotV, gotD, gotC := displayBuild(wantV, wantD, wantC)

		require.Equal(t, "N/A", gotV)
		require.Equal(t, "N/A", gotD)
		require.Equal(t, "N/A", gotC)
	})
}
