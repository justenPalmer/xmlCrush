package xmlCrush

import (
	"log"
	"strings"
	"testing"
)

func TestCrush(t *testing.T) {

	//define xml string
	testXml := `
		<?xml version="1.0" encoding="UTF-8"?>
		<note>
			<to>Tove</to>
			<from>Jani</from>
			<heading>Reminder</heading>
			<body>Don't forget me this weekend!</body>
		</note>`

	reader := strings.NewReader(testXml)

	nodes, content, err := Crush(reader)
	if err != nil {
		t.Errorf("Crush XML failed %v", err)
	}

	compareNodes := []Node{
		Node{Ind: 0, Tag: "note"},
		Node{Ind: 1, Tag: "to"},
		Node{Ind: 2, Tag: "from"},
		Node{Ind: 3, Tag: "heading"},
		Node{Ind: 4, Tag: "body"},
	}

	//ensure the nodes are correct
	for i := range nodes {
		if compareNodes[i].Ind != nodes[i].Ind || compareNodes[i].Tag != nodes[i].Tag {
			t.Errorf("Crush XML failed %v", nodes[i])
		}
	}

	compareContent := "				~~0~~			~~1~~Tove~~/1~~			~~2~~Jani~~/2~~			~~3~~Reminder~~/3~~			~~4~~Don't forget me this weekend!~~/4~~		~~/0~~"
	if compareContent != content {
		t.Errorf("Crush XML failed content string")
	}

	//log.Printf("%+v\n", nodes)
	//log.Println("content", len(content), len(compareContent))

}
