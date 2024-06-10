package scraper

import (
	"strings"
	"testing"
)

// TESTS

func TestParseXMLOneEntry(t *testing.T) {
	xml := `<rss>
<channel>
	<item>
		<title>1st entry</title>
		<link>www.entry-1.com</link>
		<pubDate>Wed, 05 Jun 2024 00:00:00 +0000</pubDate>
	</item>
</channel>
</rss>`

	reader := strings.NewReader(xml)

	feedData, err := parseXML(reader)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if len(feedData.Items) != 1 {
		t.Fatalf("Invalid number of items expected 1 got %d", len(feedData.Items))
	}

	if feedData.Items[0].Title != "1st entry" {
		t.Fatalf("Invalid title: expected '1st entry' got '%s'", feedData.Items[0].Title)
	}
}

func TestParseXMLManyEntries(t *testing.T) {
	xml := `<rss>
<channel>
	<item>
		<title>1st entry</title>
		<link>www.entry-1.com</link>
		<pubDate>Wed, 01 Jun 2024 00:00:00 +0000</pubDate>
	</item>
	<item>
		<title>2nd entry</title>
		<link>www.entry-2.com</link>
		<pubDate>Wed, 02 Jun 2024 00:00:00 +0000</pubDate>
	</item>
	<item>
		<title>3rd entry</title>
		<link>www.entry-3.com</link>
		<pubDate>Wed, 04 Jun 2024 00:00:00 +0000</pubDate>
	</item>
</channel>
</rss>`

	reader := strings.NewReader(xml)

	feedData, err := parseXML(reader)
	if err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if len(feedData.Items) != 3 {
		t.Fatalf("Invalid number of items expected 3 got %d", len(feedData.Items))
	}

	if feedData.Items[0].Title != "1st entry" {
		t.Fatalf("Invalid title: expected '1st entry' got '%s'", feedData.Items[0].Title)
	}

	if feedData.Items[1].Title != "2nd entry" {
		t.Fatalf("Invalid title: expected '2nd entry' got '%s'", feedData.Items[0].Title)
	}

	if feedData.Items[2].Title != "3rd entry" {
		t.Fatalf("Invalid title: expected '3rd entry' got '%s'", feedData.Items[0].Title)
	}
}
