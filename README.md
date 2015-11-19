# XML Crush
Library for flattening and making large XML documents accessible in Go.

## Flattened XML
XML Crush takes a different approach to XML parsing. It developed out of the need to parse massive, heavily nested and non-standardized XML documents into a usable format within Go. Traditional methods of defining structs that reflect the structure of the XML document were not cutting it because of the flexible nature of the documents. Crawling through the document one node at a time was tedious, produced horrendously ugly code, and was prone to breaking. The solution, XML Crush, allows all data to be brought in no matter the format - it allows for easy extraction of data based on tags, properties, or even by structural relationship. The code to utilize XML Crush comes out clean and semantic, and allows for easy storage and manipulation.

In XML Crush, an XML document is flattened into a content string and array of nodes. The array of nodes define the tags in the XML document along with their corresponding attributes. The content string defines data defined in and around the nodes - it also defines the position of each node, preserving the relationship between tags. This approach, while unconventional, has a number of benefits:

- Full XML document becomes available for manipulation and data extraction with one simple method call
- Flexible and reliable - no matter what the structure of the XML document, XML Crush can parse it
- All values and properties of the XML document get parsed and are made accessible - it is completely lossless
- Easily store XML Crush output into any database as an array of nodes and a string
- Extract data from the document with ease by tag, or even tag chains (by relationship of tags)
- Iterate recursively through crushed XML data with a single method call

Go Doc: https://godoc.org/github.com/justenPalmer/xmlCrush

## Crushing
To crush an XML document from an HTTP endpoint:
```` go
import(
	"fmt"
	"net/http"
	"github.com/justenPalmer/xmlCrush"
)
resp, err := http.Get("http://pathtoxmldocument.xml")
defer resp.Body.Close()
if err != nil {
	//do some error handling
}
nodes, content, err := xmlCrush.Crush(resp.Body)
if err != nil {
	//do some error handling
}
//nodes is an array (slice) of struct Node defined in XML crush
fmt.Println("nodes",nodes)

//content is a string that defines the positions of each node as well as content in and around the nodes
fmt.Println("content",content)
````
The crush method in XML Crush takes in any io.Reader and returns three arguments. The first is an array of nodes, these nodes define all the tags and their properties as defined in the XML document. The second is a string that defined all the content of the XML document as well as the positions and relationships of the nodes. The third is an error returned in case of crush failure.

## Extract one node
To extract the first node from xmlCrush output:
```` go
//extract the first "<author>" tag from the XML document, 
node, author, err := xmlCrush.ExtractOne(&nodes, content, "author")
if err != nil {
	//handle error
}

//node contains the author tag and all properties on that tag
fmt.Println("node:",node)

//author contains all data found inside the author tag
fmt.Println("author:",author)
````

## Extract all nodes
To extract all nodes of a certain tag type:
```` go
//extract all "<author>" tags from the XML document
extAry, err := xmlCrush.ExtractAll(&nodes, content, "author")
if err != nil {
	//handle error
}
for i := 0; i < len(extAry); i++ {
	//each element is of the struct type ExtractCell and contains the node, content, and position string of the node
	elem := extAry[i]

	//n will be the node found with tag type and property data attached
	fmt.Println("node:",elem.Node)

	//c will be the content string found inside the node, nested nodes will be defined by position in this string
	fmt.Println("inner content:",elem.Content)

	//pos is the position string of the extracted node, it will include every parent node by name separated by spaces
	fmt.Println("pos:",elem.Pos)
}
````

## Extracting lots of nodes with one pass
This is the most efficient way to extract lots of data. To extract data from XML Crush output with one pass:
```` go
//slices for storing the extracted data
authors := []string
links := []string

//first define an array of extraction rules, add as many rules as needed
exts := []xmlCrush.Ext{}
//define first extraction rule
ext := xmlCrush.Ext{}
//define the tags to be extracted from the document, in this case all "<author>" tags will be found in the document
ext.Tags = []string{
	"author", 
}
//define the callback to be ran for each instance of the tag found
ext.Callback = func(n xmlCrush.Node, c string, pos string) (err error) {
	//n will be the node found with tag type and property data attached
	fmt.Println("node:",n)

	//c will be the content string found inside the node, nested nodes will be defined by position in this string
	fmt.Println("inner content:",c)

	//pos is the position string of the extracted node, it will include every parent node by name separated by spaces
	fmt.Println("pos:",pos)

	//add the inner content of the author tag into the array of authors
	authors = append(authors, c)

	return
}
//attach this rule to the list of extraction rules
exts = append(exts, ext)

//second extraction rule
ext = xmlCrush.Ext{}
//define nested tag rules with parent nodes ahead of the nodes of interest
//in this case, all "<a>" tags that are found within the "<footer>" will be extracted
ext.Tags = []string{
	"footer",
	"a",
}
ext.Callback = func(n xmlCrush.Node, c string, pos string) (err error) {
	//store the href attribute on the node into the links array
	link, exists := n.Attr["href"]
	if exists {
		links = append(links, link)
	}
	return
}
exts = append(exts, ext)

//do all the extractions in one pass, pass in the nodes array and content string from the extraction as well as the array of extraction rules
err = xmlCrush.Extract(&nodes, content, exts)
if err != nil {
	//handle error
}
````

## Crawling
To crawl through all nodes of an XML document:
```` go
//define the crawl struct
crawl := xmlCrush.Crawl{
	Content: content,
	Nodes:   &nodes,
}

//iterate through each node
for !crawl.Done {
	//next will return the data associated with the next found node
	node, c, pos, err := crawl.Next()
	if err != nil {
		//handle error
	}

	//n will be the node found with tag type and property data attached
	fmt.Println("node:",n)

	//c will be the content string found inside the node, nested nodes will be defined by position in this string
	fmt.Println("inner content:",c)

	//pos is the position string of the extracted node, it will include every parent node by name separated by spaces
	fmt.Println("pos:",pos)
}
````
