package mapper

import (
	"testing"

	"github.com/GoMudEngine/ansitags"
	"github.com/stretchr/testify/assert"
)

func TestRenderANSILineDoesNotCorruptInjectedTags(t *testing.T) {
	line := []rune{'Y', '=', '-'}
	legend := map[rune]string{
		'Y': "Yard",
		'=': "Bus Stop",
		'-': "Hall-Way",
	}

	got := RenderANSILine(line, legend)

	assert.Equal(
		t,
		`<ansi fg="map-yard" bg="mapbg-yard">Y</ansi><ansi fg="map-bus-stop" bg="mapbg-bus-stop">=</ansi><ansi fg="map-hall-way" bg="mapbg-hall-way">-</ansi>`,
		got,
	)
	assert.Equal(t, "Y=-", ansitags.Parse(got, ansitags.StripTags))
	assert.NotContains(t, got, `fg"map-`)
}

func TestSanitizeLegendAlias(t *testing.T) {
	assert.Equal(t, "bus-stop", sanitizeLegendAlias("Bus Stop"))
	assert.Equal(t, "hall-way", sanitizeLegendAlias("Hall-Way"))
	assert.Equal(t, "rileys-house", sanitizeLegendAlias("Riley's House"))
}
