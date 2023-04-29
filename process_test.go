package main

import (
	"testing"
)

func TestProcessHtmlData_SingleWord(t *testing.T) {
	setup()

	data := "<body><p>{\"Lantern?\",..}  CENTIME the, chanting!?!</p></body>"
	expected := "<body><p>{\"<ruby>Lantern<rt>a light carried by a handle</rt></ruby>?\",..}  <ruby>CENTIME<rt>a money unit</rt></ruby> the, <ruby>chanting<rt>act of singing in a certain way</rt></ruby>!?!</p></body>"
	assertString(t, expected, processHtmlData(data))
}

func TestProcessHtmlData_Hyphens(t *testing.T) {
	setup()

	data := "<body><span>-Disunion--WAND—warden.</span></body>"
	expected := "<body><span>-<ruby>Disunion<rt>the ending of an association</rt></ruby>--<ruby>WAND<rt>thin stick used by a magician</rt></ruby>—<ruby>warden<rt>one who is in charge</rt></ruby>.</span></body>"
	assertString(t, expected, processHtmlData(data))
}

func TestProcessHtmlData_Pharse(t *testing.T) {
	setup()

	data := "<body><div>- (Whole life insurance, workman---fresh water?]</div></body>"
	expected := "<body><div>- (<ruby>Whole life insurance<rt>a type of life insurance</rt></ruby>, <ruby>workman<rt>a skilled worker</rt></ruby>---<ruby>fresh water<rt>water that is not salty</rt></ruby>?]</div></body>"
	assertString(t, expected, processHtmlData(data))
}

func setup() {
	if wordwiseDict == nil {
		loadWordwiseDict()
		loadLemmatizerDict()
	}
}

func assertString(t *testing.T, expected string, actual string) {
	if actual != expected {
		t.Errorf("Expected: %s\nActual: %s", expected, actual)
	}
}
