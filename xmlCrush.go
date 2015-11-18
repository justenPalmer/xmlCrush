package xmlCrush

import (
	"bufio"
	"io"
	//"log"
	"strconv"
	"strings"
)

type Node struct {
	Ind    int               `bson:"ind,omitempty" json:"ind,omitempty"`
	Tag    string            `bson:"tag,omitempty" json:"tag,omitempty"`
	Attr   map[string]string `bson:"attr,omitempty" json:"attr,omitempty"`
	Closed bool              `bson:"closed,omitempty" json:"closed,omitempty"`
}

type inNode struct {
	Tag string
	Ind int
}

type Ext struct {
	Tags     []string
	Callback func(Node, string, string) error
}

func Crush(reader io.Reader) (nodes []Node, content string, err error) {

	s := bufio.NewScanner(reader)
	start := 0
	end := 0
	ind := 0
	inNodes := []inNode{}
	for s.Scan() {
		line := strings.Replace(s.Text(), "~~", "&#126;&#126;", -1)
		start = 0
		for {
			//log.Println("find start:", line)
			start = strings.Index(line, "<")
			if start == -1 {
				break
			}
			end = strings.Index(line, ">")
			if end == -1 {
				break
			}

			//start node
			tagStr := line[start+1 : end]

			if tagStr[0:1] == "/" { //is a close

				tag := tagStr[1:]
				tag = strings.Replace(tag, " ", "", -1)
				//find if in tag
				found := false
				for i := len(inNodes) - 1; i >= 0; i-- {
					if inNodes[i].Tag == tag {
						//log.Println("close:", inNodes[i][0], start, end)
						found = true
						line = line[0:start] + "~~/" + strconv.Itoa(inNodes[i].Ind) + "~~" + line[end+1:]
						if len(inNodes) == 1 {
							inNodes = []inNode{}
						} else {
							inNodes = append(inNodes[:i], inNodes[i+1:]...) //remove node
						}
						break
					}
				}
				if !found {
					line = line[0:start] + line[end+1:] //ignore tag
				}
			} else if len(tagStr) > 2 && tagStr[0:3] == "!--" { //remove comments
				end = strings.Index(line, "-->")
				if end == -1 {
					line = line[0:start]
				} else {
					line = line[0:start] + line[end+3:]
				}
			} else if tagStr[0:1] == "?" || tagStr[0:1] == "!" { //remove doc type and xml declarations
				line = line[0:start] + line[end+1:]
			} else {
				node := Node{}
				node.Attr = make(map[string]string)
				node.Ind = ind

				spaceInd := strings.Index(tagStr, " ")
				if spaceInd == -1 {
					node.Tag = tagStr
					if len(node.Tag) > 1 && node.Tag[len(node.Tag)-1:] == "/" { //remove "/" after <a/>
						node.Tag = node.Tag[:len(node.Tag)-1]
					}
				} else { //tag has a space (may have attrs)
					node.Tag = tagStr[0:spaceInd]
					tagStr = tagStr[spaceInd+1:]

					//log.Println("tag attrs:", tagStr)

					tagA := strings.Split(tagStr, "\"")

					if len(tagA) > 0 {
						for i := 0; i < len(tagA); i += 2 {
							if len(tagA[i]) < 1 {
								continue
							}
							//get rid of space at start
							tagA[i] = strings.Replace(tagA[i], " ", "", -1)
							tagA[i] = strings.Replace(tagA[i], "=", "", -1)

							if len(tagA) > i+1 {
								node.Attr[tagA[i]] = tagA[i+1]
							} else { //no attr prop
								node.Attr[tagA[i]] = ""
							}
						}
					}
				}

				line = line[0:start] + "~~" + strconv.Itoa(node.Ind) + "~~" + line[end+1:]

				//if ended with "/", close
				if tagStr[len(tagStr)-1:] == "/" {
					node.Closed = true

				} else { //find end tag
					inN := inNode{
						Tag: node.Tag,
						Ind: node.Ind,
					}
					inNodes = append(inNodes, inN)
				}
				nodes = append(nodes, node)

				ind++
			}
		}

		content += line
	}
	return
}

func Extract(nodes *[]Node, content string, exts []Ext) (err error) {
	contentA := strings.Split(content, "~~")
	if len(contentA) < 2 {
		return
	}

	pos := ""

	for i := 1; i < len(contentA); i += 2 {
		//loop though nodes
		//log.Println("node:", contentA[i])
		if contentA[i][0:1] == "/" { //is an ending tag, remove from pos
			var closeInd int
			closeInd, err = strconv.Atoi(contentA[i][1:])

			closeNode := getNode(nodes, closeInd)
			posInd := strings.LastIndex(pos, closeNode.Tag)
			if posInd == -1 {
				continue
			}
			pos = pos[:posInd] + pos[posInd+len(closeNode.Tag)+1:]
			//log.Println("close Tag", closeNode.Tag)
		} else {
			//not a close tag, attach to call

			//set possible matches
			//loop through set of rules and eliminate those that are not matching

			ind, err := strconv.Atoi(contentA[i])
			if err != nil {
				return err
			}
			node := getNode(nodes, ind)

			//posPrevLen := len(pos)

			//if node.Tag != "" {
			pos += node.Tag + " "
			//}

			//log.Println("pos:", pos)

			for e := 0; e < len(exts); e++ {
				match := false
				ext := exts[e]
				for m := 0; m < len(ext.Tags); m++ {
					tagPos := strings.Index(pos, ext.Tags[m])

					//last tag must match end of pos
					if m == len(ext.Tags)-1 {
						//log.Println("last match:", tagPos, posPrevLen)
						if ext.Tags[m] == node.Tag {
							match = true
							break
						} else {
							match = false
							break
						}
					}
					if tagPos != -1 {
						match = true
					} else {
						match = false
						break
					}
				}
				if match {
					//log.Println("matched:", pos)
					content := ""

					if !node.Closed {
						for j := i + 1; j < len(contentA); j++ {
							if j%2 == 0 { //even, text
								content += contentA[j]
							} else if contentA[j][0:1] == "/" && contentA[j][1:] == contentA[i] {
								break
							} else {
								content += "~~" + contentA[j] + "~~"
							}
							//log.Println("ext pos:", pos)
						}
					}
					//log.Println("extract:", node, content)
					err = ext.Callback(node, content, pos)
				}
			}

			if node.Closed {
				pos = pos[0 : len(pos)-len(node.Tag)-1]
			}
		}
	}

	return
}

func ExtractOne(nodes *[]Node, content string, tag string) (node Node, c string, err error) {
	contentA := strings.Split(content, "~~")
	if len(contentA) < 2 {
		return
	}

	pos := ""

	for i := 1; i < len(contentA); i += 2 {
		//loop though nodes
		//log.Println("node:", contentA[i])
		if contentA[i][0:1] == "/" { //is an ending tag, remove from pos
			var closeInd int
			closeInd, err = strconv.Atoi(contentA[i][1:])

			closeNode := getNode(nodes, closeInd)
			posInd := strings.LastIndex(pos, closeNode.Tag)
			if posInd == -1 {
				continue
			}
			pos = pos[:posInd] + pos[posInd+len(closeNode.Tag):]
		} else {
			//not a close tag, attach to call

			//set possible matches
			//loop through set of rules and eliminate those that are not matching

			ind, err := strconv.Atoi(contentA[i])
			if err != nil {
				return node, c, err
			}
			node := getNode(nodes, ind)

			pos += node.Tag + " "

			//log.Println("pos:", pos)

			tagPos := strings.Index(pos, tag)
			if tagPos == -1 {
				continue
			}

			if !node.Closed {
				for j := i + 1; j < len(contentA); j++ {
					if j%2 == 0 { //even, text
						c += contentA[j]
					} else if contentA[j][0:1] == "/" && contentA[j][1:] == contentA[i] {
						break
					} else {
						c += "~~" + contentA[j] + "~~"
					}
				}
			}
			//log.Println("extract:", node, content)
			break
		}
	}

	return
}

type ExtractCell struct {
	Pos     string
	Node    Node
	Content string
}

func ExtractAll(nodes *[]Node, content string, tag string) (cells []ExtractCell, err error) {
	contentA := strings.Split(content, "~~")
	if len(contentA) < 2 {
		return
	}

	pos := ""

	for i := 1; i < len(contentA); i += 2 {
		//loop though nodes
		//log.Println("node:", contentA[i])
		if contentA[i][0:1] == "/" { //is an ending tag, remove from pos
			var closeInd int
			closeInd, err = strconv.Atoi(contentA[i][1:])
			closeNode := getNode(nodes, closeInd)
			posInd := strings.LastIndex(pos, closeNode.Tag)
			if posInd == -1 {
				continue
			}
			pos = pos[:posInd] + pos[posInd+len(closeNode.Tag):]
		} else {
			//not a close tag, attach to call

			cell := ExtractCell{}

			//set possible matches
			//loop through set of rules and eliminate those that are not matching

			ind, err := strconv.Atoi(contentA[i])
			if err != nil {
				return cells, err
			}
			cell.Node = getNode(nodes, ind)

			pos += cell.Node.Tag + " "

			//log.Println("pos:", pos)

			tagPos := strings.Index(pos, tag)
			if tagPos == -1 {
				continue
			}
			if !cell.Node.Closed {
				for j := i + 1; j < len(contentA); j++ {
					if j%2 == 0 { //even, text
						cell.Content += contentA[j]
					} else if contentA[j][0:1] == "/" && contentA[j][1:] == contentA[i] {
						break
					} else {
						cell.Content += "~~" + contentA[j] + "~~"
					}
				}
			}

			cells = append(cells, cell)

			//log.Println("extract:", node, content)
			//break

		}
	}

	return
}

type Crawl struct {
	Nodes   *[]Node
	Content string //string of content, this can be updated
	Pos     string // "div a "
	Done    bool
}

func (this *Crawl) Next() (node Node, content string, pos string, err error) {
	start := strings.Index(this.Content, "~~")

	pos = this.Pos

	if start == -1 { //no tags found, return full content string
		content = this.Content
		this.Done = true
		return
	}
	if start != 0 { //tag not start, return everything before tag
		content = this.Content[0:start]
		this.Content = this.Content[start:] //trim off start of content
		return
	}

	//is a tag

	startClose := strings.Index(this.Content[2:], "~~")

	if startClose == -1 {
		this.Done = true
		return
	}

	startClose += 2

	//log.Println("content:", this.Content)

	nodeStr := this.Content[2:startClose]
	//log.Println("nodeStr:", nodeStr)

	if len(nodeStr) < 1 {
		this.Content = this.Content[startClose+2:]
		if len(this.Content) < 1 {
			this.Done = true
		}
		return
	}

	if nodeStr[0:1] == "/" { // is a close tag - assume there was a parsing bug and ignore
		this.Content = this.Content[startClose+2:]
		if len(this.Content) < 1 {
			this.Done = true
		}
		return
	}

	ind, err := strconv.Atoi(nodeStr)
	if err != nil {
		return
	}

	nodes := *this.Nodes
	node = nodes[ind]

	pos += node.Tag + " "

	if node.Closed {
		this.Content = this.Content[startClose+2:]
	} else {
		//find node close
		endNode := "~~/" + nodeStr + "~~"
		end := strings.Index(this.Content, endNode)
		content = this.Content[startClose+2 : end]
		this.Content = this.Content[end+len(endNode):]
	}

	if len(this.Content) < 1 {
		this.Done = true
	}

	return
}

func getNode(nodes *[]Node, ind int) (node Node) {
	nP := *nodes
	node = nP[ind]
	return
}

func GetNodeAsContent(nodes *[]Node, c string, ind int, wrap bool) (content string) {
	start := "~~" + strconv.Itoa(ind) + "~~"
	end := "~~/" + strconv.Itoa(ind) + "~~"

	startI := strings.Index(c, start)
	endI := strings.Index(c, end)

	if startI == -1 || endI == -1 {
		return
	}

	if wrap {
		endI += len(end)
	} else {
		startI += len(start)
	}

	//log.Println("start:", startI, endI)

	content = c[startI:endI]
	return
}

func StripNodes(content string) (s string) {
	contentA := strings.Split(content, "~~")
	for i := 0; i < len(contentA); i += 2 {
		s += contentA[i]
	}
	return
}
